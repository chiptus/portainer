package security

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/dataservices"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/ssl"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

type (
	// RequestBouncer represents an entity that manages API request accesses
	RequestBouncer struct {
		dataStore      dataservices.DataStore
		jwtService     portaineree.JWTService
		apiKeyService  apikey.APIKeyService
		licenseService portaineree.LicenseService
		sslService     *ssl.Service
	}

	// RestrictedRequestContext is a data structure containing information
	// used in AuthenticatedAccess
	RestrictedRequestContext struct {
		IsAdmin         bool
		IsTeamLeader    bool
		UserID          portaineree.UserID
		UserMemberships []portaineree.TeamMembership
	}

	// tokenLookup looks up a token in the request
	tokenLookup func(*http.Request) *portaineree.TokenData
)

const apiKeyHeader = "X-API-KEY"

// NewRequestBouncer initializes a new RequestBouncer
func NewRequestBouncer(dataStore dataservices.DataStore, licenseService portaineree.LicenseService, jwtService portaineree.JWTService, apiKeyService apikey.APIKeyService, sslService *ssl.Service) *RequestBouncer {
	return &RequestBouncer{
		dataStore:      dataStore,
		jwtService:     jwtService,
		apiKeyService:  apiKeyService,
		licenseService: licenseService,
		sslService:     sslService,
	}
}

// PublicAccess defines a security check for public API environments(endpoints).
// No authentication is required to access these environments(endpoints).
func (bouncer *RequestBouncer) PublicAccess(h http.Handler) http.Handler {
	h = mwSecureHeaders(h)
	return h
}

// AdminAccess is an alias for RestrictedAddress
// It's not removed as it's used across our codebase and removing will cause conflicts with CE
func (bouncer *RequestBouncer) AdminAccess(h http.Handler) http.Handler {
	return bouncer.RestrictedAccess(h)
}

// RestrictedAccess defines a security check for restricted API environments(endpoints).
// Authentication and authorizations are required to access these environments(endpoints).
// The request context will be enhanced with a RestrictedRequestContext object
// that might be used later to inside the API operation for extra authorization validation
// and resource filtering.
//
// Bouncer operations are applied backwards:
//  - Parse the JWT from the request and stored in context, user has to be authenticated
//  - Validate the software license
//  - Authorize the user to the request from the token data
//  - Upgrade to the restricted request
func (bouncer *RequestBouncer) RestrictedAccess(h http.Handler) http.Handler {
	h = bouncer.mwUpgradeToRestrictedRequest(h)
	h = bouncer.mwCheckPortainerAuthorizations(h)
	h = bouncer.mwCheckLicense(h)
	h = bouncer.mwAuthenticatedUser(h)
	return h
}

// AuthenticatedAccess defines a security check for restricted API environments(endpoints).
// Authentication is required to access these environments(endpoints).
// The request context will be enhanced with a RestrictedRequestContext object
// that might be used later to inside the API operation for extra authorization validation
// and resource filtering.
//
// Bouncer operations are applied backwards:
//  - Parse the JWT from the request and stored in context, user has to be authenticated
//  - Upgrade to the restricted request
func (bouncer *RequestBouncer) AuthenticatedAccess(h http.Handler) http.Handler {
	h = bouncer.mwUpgradeToRestrictedRequest(h)
	h = bouncer.mwAuthenticatedUser(h)
	return h
}

// AuthorizedEndpointOperation retrieves the JWT token from the request context and verifies
// that the user can access the specified environment(endpoint).
// If the authorizationCheck flag is set, it will also validate that the user can execute the specified operation.
// An error is returned when access to the environment(endpoint) is denied or if the user do not have the required
// authorization to execute the operation.
func (bouncer *RequestBouncer) AuthorizedEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint, authorizationCheck bool) error {
	tokenData, err := RetrieveTokenData(r)
	if err != nil {
		return err
	}

	if tokenData.Role == portaineree.AdministratorRole {
		return nil
	}

	memberships, err := bouncer.dataStore.TeamMembership().TeamMembershipsByUserID(tokenData.ID)
	if err != nil {
		return err
	}

	group, err := bouncer.dataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	if !authorizedEndpointAccess(endpoint, group, tokenData.ID, memberships) {
		return httperrors.ErrEndpointAccessDenied
	}

	if authorizationCheck {
		err = bouncer.checkEndpointOperationAuthorization(r, endpoint)
		if err != nil {
			return ErrAuthorizationRequired
		}
	}

	return nil
}

func (bouncer *RequestBouncer) AuthorizedClientTLSConn(r *http.Request) error {
	sslSettings, err := bouncer.dataStore.SSLSettings().Settings()
	if err != nil {
		return fmt.Errorf("could not retrieve the TLS settings: %w", err)
	}

	if sslSettings.CACertPath != "" {
		caCertErr := bouncer.sslService.ValidateCACert(r.TLS)
		if caCertErr != nil {
			return caCertErr
		}
	}

	return nil
}

// AuthorizedEdgeEndpointOperation verifies that the request was received from a valid Edge environment(endpoint)
func (bouncer *RequestBouncer) AuthorizedEdgeEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint) error {
	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return errors.New("Invalid environment type")
	}

	edgeIdentifier := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)
	if edgeIdentifier == "" {
		return errors.New("missing Edge identifier")
	}

	if endpoint.EdgeID != "" && endpoint.EdgeID != edgeIdentifier {
		return errors.New("invalid Edge identifier")
	}

	err := bouncer.AuthorizedClientTLSConn(r)
	if err != nil {
		return err
	}

	if endpoint.LastCheckInDate > 0 || endpoint.UserTrusted {
		return nil
	}

	settings, err := bouncer.dataStore.Settings().Settings()
	if err != nil {
		return fmt.Errorf("could not retrieve the settings: %w", err)
	}

	if !settings.TrustOnFirstConnect {
		return errors.New("the device has not been trusted yet")
	}

	return nil
}

func (bouncer *RequestBouncer) checkEndpointOperationAuthorization(r *http.Request, endpoint *portaineree.Endpoint) error {
	tokenData, err := RetrieveTokenData(r)
	if err != nil {
		return err
	}

	if tokenData.Role == portaineree.AdministratorRole {
		return nil
	}

	user, err := bouncer.dataStore.User().User(tokenData.ID)
	if err != nil {
		return err
	}

	apiOperation := &portaineree.APIOperationAuthorizationRequest{
		Path:           r.URL.String(),
		Method:         r.Method,
		Authorizations: user.EndpointAuthorizations[endpoint.ID],
	}

	if !authorizedOperation(apiOperation) {
		return errors.New("Unauthorized")
	}

	return nil
}

// mwAuthenticatedUser authenticates a request by
// - adding a secure handlers to the response
// - authenticating the request with a valid token
func (bouncer *RequestBouncer) mwAuthenticatedUser(h http.Handler) http.Handler {
	h = bouncer.mwAuthenticateFirst([]tokenLookup{
		bouncer.JWTAuthLookup,
		bouncer.apiKeyLookup,
	}, h)
	h = mwSecureHeaders(h)
	return h
}

// mwCheckLicense will verify that the instance license is valid
func (bouncer *RequestBouncer) mwCheckLicense(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenData, err := RetrieveTokenData(r)
		if err != nil {
			httperror.WriteError(w, http.StatusForbidden, "Access denied", httperrors.ErrUnauthorized)
			return
		}

		if tokenData.Role == portaineree.AdministratorRole {
			next.ServeHTTP(w, r)
			return
		}

		info := bouncer.licenseService.Info()

		if !info.Valid {
			log.Printf("[INFO] [http,security,bouncer] [msg: licenses are invalid]")
			httperror.WriteError(w, http.StatusForbidden, "License is not valid", httperrors.ErrUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// mwCheckPortainerAuthorizations will verify that the user has the required authorization to access
// a specific API environment(endpoint).
func (bouncer *RequestBouncer) mwCheckPortainerAuthorizations(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenData, err := RetrieveTokenData(r)
		if err != nil {
			httperror.WriteError(w, http.StatusForbidden, "Access denied", httperrors.ErrUnauthorized)
			return
		}

		if tokenData.Role == portaineree.AdministratorRole {
			next.ServeHTTP(w, r)
			return
		}

		user, err := bouncer.dataStore.User().User(tokenData.ID)
		if err != nil && err == bolterrors.ErrObjectNotFound {
			httperror.WriteError(w, http.StatusUnauthorized, "Unauthorized", httperrors.ErrUnauthorized)
			return
		} else if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable to retrieve user details from the database", err)
			return
		}

		apiOperation := &portaineree.APIOperationAuthorizationRequest{
			Path:           r.URL.String(),
			Method:         r.Method,
			Authorizations: user.PortainerAuthorizations,
		}

		if !authorizedOperation(apiOperation) {
			httperror.WriteError(w, http.StatusForbidden, "Access denied", ErrAuthorizationRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// mwUpgradeToRestrictedRequest will enhance the current request with
// a new RestrictedRequestContext object.
func (bouncer *RequestBouncer) mwUpgradeToRestrictedRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenData, err := RetrieveTokenData(r)
		if err != nil {
			httperror.WriteError(w, http.StatusForbidden, "Access denied", httperrors.ErrResourceAccessDenied)
			return
		}

		requestContext, err := bouncer.newRestrictedContextRequest(tokenData.ID, tokenData.Role)
		if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable to create restricted request context ", err)
			return
		}

		ctx := StoreRestrictedRequestContext(r, requestContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mwAuthenticateFirst authenticates a request an auth token.
// A result of a first succeded token lookup would be used for the authentication.
func (bouncer *RequestBouncer) mwAuthenticateFirst(tokenLookups []tokenLookup, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token *portaineree.TokenData

		for _, lookup := range tokenLookups {
			token = lookup(r)

			if token != nil {
				break
			}
		}

		if token == nil {
			httperror.WriteError(w, http.StatusUnauthorized, "A valid authorisation token is missing", httperrors.ErrUnauthorized)
			return
		}

		user, _ := bouncer.dataStore.User().User(token.ID)
		if user == nil {
			httperror.WriteError(w, http.StatusUnauthorized, "An authorisation token is invalid", httperrors.ErrUnauthorized)
			return
		}

		ctx := StoreTokenData(r, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWTAuthLookup looks up a valid bearer in the request.
func (bouncer *RequestBouncer) JWTAuthLookup(r *http.Request) *portaineree.TokenData {
	// get token from the Authorization header or query parameter
	token, err := extractBearerToken(r)
	if err != nil {
		return nil
	}

	tokenData, err := bouncer.jwtService.ParseAndVerifyToken(token)
	if err != nil {
		return nil
	}

	return tokenData
}

// apiKeyLookup looks up an verifies an api-key by:
// - computing the digest of the raw api-key
// - verifying it exists in cache/database
// - matching the key to a user (ID, Role)
// If the key is valid/verified, the last updated time of the key is updated.
// Successful verification of the key will return a TokenData object - since the downstream handlers
// utilise the token injected in the request context.
func (bouncer *RequestBouncer) apiKeyLookup(r *http.Request) *portaineree.TokenData {
	rawAPIKey, ok := extractAPIKey(r)
	if !ok {
		return nil
	}

	digest := bouncer.apiKeyService.HashRaw(rawAPIKey)

	user, apiKey, err := bouncer.apiKeyService.GetDigestUserAndKey(digest)
	if err != nil {
		return nil
	}

	tokenData := &portaineree.TokenData{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}
	if _, err := bouncer.jwtService.GenerateToken(tokenData); err != nil {
		return nil
	}

	// update the last used time of the key
	apiKey.LastUsed = time.Now().UTC().Unix()
	bouncer.apiKeyService.UpdateAPIKey(&apiKey)

	return tokenData
}

// extractBearerToken extracts the Bearer token from the request header or query parameter and returns the token.
func extractBearerToken(r *http.Request) (string, error) {
	// Token might be set via the "token" query parameter.
	// For example, in websocket requests
	// For these cases, hide the token from the query
	query := r.URL.Query()
	token := query.Get("token")
	if token != "" {
		query.Del("token")
		r.URL.RawQuery = query.Encode()
	}

	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token == "" {
		return "", httperrors.ErrUnauthorized
	}
	return token, nil
}

// extractAPIKey extracts the api key from the api key request header or query params.
func extractAPIKey(r *http.Request) (apikey string, ok bool) {
	// extract the API key from the request header
	apikey = r.Header.Get(apiKeyHeader)
	if apikey != "" {
		return apikey, true
	}

	// extract the API key from query params.
	// Case-insensitive check for the "X-API-KEY" query param.
	query := r.URL.Query()
	for k, v := range query {
		if strings.EqualFold(k, apiKeyHeader) {
			return v[0], true
		}
	}

	return "", false
}

// mwSecureHeaders provides secure headers middleware for handlers.
func mwSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-XSS-Protection", "1; mode=block")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func (bouncer *RequestBouncer) newRestrictedContextRequest(userID portaineree.UserID, userRole portaineree.UserRole) (*RestrictedRequestContext, error) {
	if userRole == portaineree.AdministratorRole {
		return &RestrictedRequestContext{
			IsAdmin: true,
			UserID:  userID,
		}, nil
	}

	memberships, err := bouncer.dataStore.TeamMembership().TeamMembershipsByUserID(userID)
	if err != nil {
		return nil, err
	}

	isTeamLeader := false
	for _, membership := range memberships {
		if membership.Role == portaineree.TeamLeader {
			isTeamLeader = true
		}
	}

	return &RestrictedRequestContext{
		IsAdmin:         false,
		UserID:          userID,
		IsTeamLeader:    isTeamLeader,
		UserMemberships: memberships,
	}, nil
}

// EdgeComputeOperation defines a restriced edge compute operation.
// Use of this operation will only be authorized if edgeCompute is enabled in settings
func (bouncer *RequestBouncer) EdgeComputeOperation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings, err := bouncer.dataStore.Settings().Settings()
		if err != nil {
			httperror.WriteError(w, http.StatusServiceUnavailable, "Unable to retrieve settings", err)
			return
		}

		if !settings.EnableEdgeComputeFeatures {
			httperror.WriteError(w, http.StatusServiceUnavailable, "Edge compute features are disabled", errors.New("Edge compute features are disabled"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/internal/authorization"
)

type oauthPayload struct {
	// OAuth code returned from OAuth Provided
	Code string
}

func (payload *oauthPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Code) {
		return errors.New("Invalid OAuth authorization code")
	}
	return nil
}

func (handler *Handler) authenticateOAuth(code string, settings *portainer.OAuthSettings) (*portainer.OAuthInfo, error) {
	if code == "" {
		return nil, errors.New("Invalid OAuth authorization code")
	}

	if settings == nil {
		return nil, errors.New("Invalid OAuth configuration")
	}

	authInfo, err := handler.OAuthService.Authenticate(code, settings)
	if err != nil {
		return nil, err
	}

	return authInfo, nil
}

// @id ValidateOAuth
// @summary Authenticate with OAuth
// @description **Access policy**: public
// @tags auth
// @accept json
// @produce json
// @param body body oauthPayload true "OAuth Credentials used for authentication"
// @success 200 {object} authenticateResponse "Success"
// @failure 400 "Invalid request"
// @failure 422 "Invalid Credentials"
// @failure 500 "Server error"
// @router /auth/oauth/validate [post]
func (handler *Handler) validateOAuth(w http.ResponseWriter, r *http.Request) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method: portainer.AuthenticationOAuth,
	}

	var payload oauthPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request payload",
			Err:        err,
		}
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to retrieve settings from the database",
			Err:        err,
		}
	}

	if settings.AuthenticationMethod != portainer.AuthenticationOAuth {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "OAuth authentication is not enabled",
			Err:        errors.New("OAuth authentication is not enabled"),
		}
	}

	authInfo, err := handler.authenticateOAuth(payload.Code, &settings.OAuthSettings)
	if err != nil {
		log.Printf("[DEBUG] - OAuth authentication error: %s", err)
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to authenticate through OAuth",
			Err:        httperrors.ErrUnauthorized,
		}
	}

	resp.Username = authInfo.Username

	user, err := handler.DataStore.User().UserByUsername(authInfo.Username)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to retrieve a user with the specified username from the database",
			Err:        err,
		}
	}

	if user == nil && !settings.OAuthSettings.OAuthAutoMapTeamMemberships && !settings.OAuthSettings.OAuthAutoCreateUsers {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "Auto OAuth team membership failed: user not created beforehand in Portainer and automatic user provisioning not enabled",
			Err:        httperrors.ErrUnauthorized,
		}
	}

	//try to match retrieved oauth teams with pre-set auto admin claims
	isValidAdminClaims, err := validateAdminClaims(settings.OAuthSettings, authInfo.Teams)
	if err != nil {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to validate OAuth auto admin claims",
			Err:        err,
		}
	}

	if user != nil {
		//if user exists, check oauth settings and update user's role if needed
		if err := handler.updateUser(user, settings.OAuthSettings, isValidAdminClaims); err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to update user OAuth membership",
				Err:        err,
			}
		}
	} else {
		//if user not exists, create a new user with the correct role (according to AdminAutoPopulate settings)
		user = &portainer.User{
			Username:                authInfo.Username,
			Role:                    portainer.StandardUserRole,
			PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
		}
		if err := handler.createUserAndDefaultTeamMembership(user, settings.OAuthSettings, isValidAdminClaims); err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to create user or OAuth membership",
				Err:        err,
			}
		}
	}

	if settings.OAuthSettings.OAuthAutoMapTeamMemberships {
		if settings.OAuthSettings.TeamMemberships.OAuthClaimName == "" {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to process user oauth team memberships",
				Err:        errors.New("empty value set for oauth team membership Claim name"),
			}
		}

		err = updateOAuthTeamMemberships(handler.DataStore, settings.OAuthSettings, *user, authInfo.Teams)
		if err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to update user oauth team memberships",
				Err:        err,
			}
		}

		err = handler.AuthorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to update user authorizations",
				Err:        err,
			}
		}
	}

	info := handler.LicenseService.Info()

	if user.Role != portainer.AdministratorRole && !info.Valid {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "License is not valid",
			Err:        httperrors.ErrNoValidLicense,
		}
	}

	return handler.writeToken(w, user, portainer.AuthenticationOAuth)
}

func (handler *Handler) updateUser(user *portainer.User, oauthSettings portainer.OAuthSettings, validAdminClaim bool) error {
	//if AdminAutoPopulate is switched off, no need to update user
	if !oauthSettings.TeamMemberships.AdminAutoPopulate {
		return nil
	}

	if validAdminClaim {
		user.Role = portainer.AdministratorRole
	} else {
		user.Role = portainer.StandardUserRole
	}

	if err := handler.DataStore.User().UpdateUser(user.ID, user); err != nil {
		return errors.New("Unable to persist user changes inside the database")
	}

	if err := handler.AuthorizationService.UpdateUsersAuthorizations(); err != nil {
		return errors.New("Unable to update user authorizations")
	}

	return nil
}

func (handler *Handler) createUserAndDefaultTeamMembership(user *portainer.User, oauthSettings portainer.OAuthSettings, validAdminClaim bool) error {
	//if AdminAutoPopulate is switched on and valid OAuth group is identified
	//set user role as admin
	if oauthSettings.TeamMemberships.AdminAutoPopulate && validAdminClaim {
		user.Role = portainer.AdministratorRole
	}

	err := handler.DataStore.User().CreateUser(user)
	if err != nil {
		return errors.New("Unable to persist user inside the database")
	}

	if oauthSettings.DefaultTeamID != 0 {
		membership := &portainer.TeamMembership{
			UserID: user.ID,
			TeamID: oauthSettings.DefaultTeamID,
			Role:   portainer.TeamMember,
		}

		err = handler.DataStore.TeamMembership().CreateTeamMembership(membership)
		if err != nil {
			return errors.New("Unable to persist team membership inside the database")
		}
	}

	err = handler.AuthorizationService.UpdateUsersAuthorizations()
	if err != nil {
		return errors.New("Unable to update user authorizations")
	}

	return nil
}

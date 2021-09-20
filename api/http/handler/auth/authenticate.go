package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/internal/authorization"
)

type authenticatePayload struct {
	// Username
	Username string `example:"admin" validate:"required"`
	// Password
	Password string `example:"mypassword" validate:"required"`
}

type authenticateResponse struct {
	// JWT token used to authenticate against the API
	JWT string `json:"jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MSwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOjEsImV4cCI6MTQ5OTM3NjE1NH0.NJ6vE8FY1WG6jsRQzfMqeatJ4vh2TWAeeYfDhP71YEE"`
}

func (payload *authenticatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Username) {
		return errors.New("Invalid username")
	}
	if govalidator.IsNull(payload.Password) {
		return errors.New("Invalid password")
	}
	return nil
}

// @id AuthenticateUser
// @summary Authenticate
// @description Use this environment(endpoint) to authenticate against Portainer using a username and password.
// @tags auth
// @accept json
// @produce json
// @param body body authenticatePayload true "Credentials used for authentication"
// @success 200 {object} authenticateResponse "Success"
// @failure 400 "Invalid request"
// @failure 422 "Invalid Credentials"
// @failure 500 "Server error"
// @router /auth [post]
func (handler *Handler) authenticate(rw http.ResponseWriter, r *http.Request) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method: portainer.AuthenticationInternal,
	}

	var payload authenticatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid request payload",
				Err:        err,
			}

	}

	resp.Username = payload.Username

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to retrieve settings from the database",
			Err:        err,
		}
	}

	u, err := handler.DataStore.User().UserByUsername(payload.Username)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to retrieve a user with the specified username from the database",
			Err:        err,
		}
	}

	if err == bolterrors.ErrObjectNotFound && (settings.AuthenticationMethod == portainer.AuthenticationInternal || settings.AuthenticationMethod == portainer.AuthenticationOAuth) {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    "Invalid credentials",
			Err:        httperrors.ErrUnauthorized,
		}

	}

	if settings.AuthenticationMethod == portainer.AuthenticationLDAP {
		if u == nil && settings.LDAPSettings.AutoCreateUsers {
			return handler.authenticateLDAPAndCreateUser(rw, payload.Username, payload.Password, &settings.LDAPSettings)
		} else if u == nil && !settings.LDAPSettings.AutoCreateUsers {
			return resp,
				&httperror.HandlerError{
					StatusCode: http.StatusUnprocessableEntity,
					Message:    "Invalid credentials",
					Err:        httperrors.ErrUnauthorized,
				}
		}
		return handler.authenticateLDAP(rw, u, payload.Password, &settings.LDAPSettings)
	}

	return handler.authenticateInternal(rw, u, payload.Password)
}

func (handler *Handler) authenticateLDAP(w http.ResponseWriter, user *portainer.User, password string, ldapSettings *portainer.LDAPSettings) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method:   portainer.AuthenticationLDAP,
		Username: user.Username,
	}

	err := handler.LDAPService.AuthenticateUser(user.Username, password, ldapSettings)
	if err != nil {
		return handler.authenticateInternal(w, user, password)
	}

	err = handler.addUserIntoTeams(user, ldapSettings)
	if err != nil {
		log.Printf("Warning: unable to automatically add user into teams: %s\n", err.Error())
	}

	err = handler.AuthorizationService.UpdateUsersAuthorizations()
	if err != nil {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to update user authorizations",
				Err:        err,
			}

	}

	info := handler.LicenseService.Info()

	if user.Role != portainer.AdministratorRole && !info.Valid {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusForbidden,
				Message:    "License is not valid",
				Err:        httperrors.ErrNoValidLicense,
			}

	}

	return handler.writeToken(w, user, resp.Method)
}

func (handler *Handler) authenticateInternal(w http.ResponseWriter, user *portainer.User, password string) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method:   portainer.AuthenticationInternal,
		Username: user.Username,
	}

	err := handler.CryptoService.CompareHashAndData(user.Password, password)
	if err != nil {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusUnprocessableEntity,
				Message:    "Invalid credentials",
				Err:        httperrors.ErrUnauthorized,
			}

	}

	info := handler.LicenseService.Info()

	if user.Role != portainer.AdministratorRole && !info.Valid {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusForbidden,
				Message:    "License is not valid",
				Err:        httperrors.ErrNoValidLicense,
			}

	}

	return handler.writeToken(w, user, resp.Method)
}

func (handler *Handler) authenticateLDAPAndCreateUser(w http.ResponseWriter, username, password string, ldapSettings *portainer.LDAPSettings) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method:   portainer.AuthenticationLDAP,
		Username: username,
	}

	err := handler.LDAPService.AuthenticateUser(username, password, ldapSettings)
	if err != nil {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusUnprocessableEntity,
				Message:    "Invalid credentials",
				Err:        err,
			}

	}

	user := &portainer.User{
		Username:                username,
		Role:                    portainer.StandardUserRole,
		PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
	}

	err = handler.DataStore.User().CreateUser(user)
	if err != nil {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to persist user inside the database",
				Err:        err,
			}

	}

	err = handler.addUserIntoTeams(user, ldapSettings)
	if err != nil {
		log.Printf("Warning: unable to automatically add user into teams: %s\n", err.Error())
	}

	err = handler.AuthorizationService.UpdateUsersAuthorizations()
	if err != nil {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to update user authorizations",
				Err:        err,
			}

	}

	info := handler.LicenseService.Info()

	if !info.Valid {
		return resp,
			&httperror.HandlerError{
				StatusCode: http.StatusForbidden,
				Message:    "License is not valid",
				Err:        httperrors.ErrNoValidLicense,
			}

	}

	return handler.writeToken(w, user, resp.Method)
}

func (handler *Handler) writeToken(w http.ResponseWriter, user *portainer.User, method portainer.AuthenticationMethod) (*authMiddlewareResponse, *httperror.HandlerError) {
	tokenData := composeTokenData(user)

	return handler.persistAndWriteToken(w, tokenData, nil, method)
}

func (handler *Handler) persistAndWriteToken(w http.ResponseWriter, tokenData *portainer.TokenData, expiryTime *time.Time, method portainer.AuthenticationMethod) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Username: tokenData.Username,
		Method:   method,
	}

	var token string
	var err error

	if method == portainer.AuthenticationOAuth {
		token, err = handler.JWTService.GenerateTokenForOAuth(tokenData, expiryTime)
		if err != nil {
			return resp,
				&httperror.HandlerError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Unable to generate JWT token for OAuth",
					Err:        err,
				}

		}
	} else {
		token, err = handler.JWTService.GenerateToken(tokenData)
		if err != nil {
			return resp,
				&httperror.HandlerError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Unable to generate JWT token",
					Err:        err,
				}

		}
	}
	return resp, response.JSON(w, &authenticateResponse{JWT: token})

}

func (handler *Handler) addUserIntoTeams(user *portainer.User, settings *portainer.LDAPSettings) error {
	teams, err := handler.DataStore.Team().Teams()
	if err != nil {
		return err
	}

	userGroups, err := handler.LDAPService.GetUserGroups(user.Username, settings)
	if err != nil {
		return err
	}

	userMemberships, err := handler.DataStore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return err
	}

	for _, team := range teams {
		if teamExists(team.Name, userGroups) {

			if teamMembershipExists(team.ID, userMemberships) {
				continue
			}

			membership := &portainer.TeamMembership{
				UserID: user.ID,
				TeamID: team.ID,
				Role:   portainer.TeamMember,
			}

			err := handler.DataStore.TeamMembership().CreateTeamMembership(membership)
			if err != nil {

				return err
			}
		}
	}

	return nil
}

func teamExists(teamName string, ldapGroups []string) bool {
	for _, group := range ldapGroups {
		if strings.ToLower(group) == strings.ToLower(teamName) {
			return true
		}
	}
	return false
}

func teamMembershipExists(teamID portainer.TeamID, memberships []portainer.TeamMembership) bool {
	for _, membership := range memberships {
		if membership.TeamID == teamID {
			return true
		}
	}
	return false
}

func composeTokenData(user *portainer.User) *portainer.TokenData {
	return &portainer.TokenData{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}
}

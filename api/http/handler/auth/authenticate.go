package auth

import (
	"net/http"
	"strings"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
// @description **Access policy**: public
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
		Method: portaineree.AuthenticationInternal,
	}

	var payload authenticatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return resp, httperror.BadRequest("Invalid request payload", err)
	}

	resp.Username = payload.Username

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return resp, httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	user, err := handler.DataStore.User().UserByUsername(payload.Username)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return resp, httperror.InternalServerError("Unable to retrieve a user with the specified username from the database", err)
	}

	if err == bolterrors.ErrObjectNotFound &&
		(settings.AuthenticationMethod == portaineree.AuthenticationInternal ||
			settings.AuthenticationMethod == portaineree.AuthenticationOAuth ||
			(settings.AuthenticationMethod == portaineree.AuthenticationLDAP && !settings.LDAPSettings.AutoCreateUsers)) {
		return resp, &httperror.HandlerError{StatusCode: http.StatusUnprocessableEntity, Message: "Invalid credentials", Err: httperrors.ErrUnauthorized}
	}

	// If the free subscription license is enforced, the standard user are not allowed to log in
	if user != nil && user.Role != portaineree.AdministratorRole && handler.LicenseService.ShouldEnforceOveruse() {
		return resp, &httperror.HandlerError{StatusCode: http.StatusPaymentRequired, Message: "Node limit exceeds the 5 node free license, please contact your administrator", Err: httperrors.ErrLicenseOverused}
	}

	if user != nil && isUserInitialAdmin(user) || settings.AuthenticationMethod == portaineree.AuthenticationInternal {
		return handler.authenticateInternal(rw, user, payload.Password)
	}

	if settings.AuthenticationMethod == portaineree.AuthenticationOAuth {
		resp.Method = portaineree.AuthenticationOAuth
		return resp, &httperror.HandlerError{StatusCode: http.StatusUnprocessableEntity, Message: "Only initial admin is allowed to login without oauth", Err: httperrors.ErrUnauthorized}
	}

	if settings.AuthenticationMethod == portaineree.AuthenticationLDAP {
		return handler.authenticateLDAP(rw, user, payload.Username, payload.Password, &settings.LDAPSettings)
	}

	return resp, &httperror.HandlerError{StatusCode: http.StatusUnprocessableEntity, Message: "Login method is not supported", Err: httperrors.ErrUnauthorized}
}

func isUserInitialAdmin(user *portaineree.User) bool {
	return int(user.ID) == 1
}

func (handler *Handler) authenticateLDAP(w http.ResponseWriter, user *portaineree.User, username, password string, ldapSettings *portaineree.LDAPSettings) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method:   portaineree.AuthenticationLDAP,
		Username: username,
	}

	err := handler.LDAPService.AuthenticateUser(username, password, ldapSettings)
	if err != nil {
		return resp, httperror.Forbidden("Only initial admin is allowed to login without oauth", err)
	}

	if user == nil {
		user = &portaineree.User{
			Username:                username,
			Role:                    portaineree.StandardUserRole,
			PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
		}

		err = handler.DataStore.User().Create(user)
		if err != nil {
			return resp, httperror.InternalServerError("Unable to persist user inside the database", err)
		}
	}

	if ldapSettings.AdminAutoPopulate {
		isLDAPAdmin, err := isLDAPAdmin(resp.Username, handler.LDAPService, ldapSettings)
		if err != nil {
			return resp,
				httperror.InternalServerError("Failed to search and match LDAP admin groups", err)
		}

		if isLDAPAdmin != (user.Role == portaineree.AdministratorRole) {
			userRole := portaineree.StandardUserRole
			if isLDAPAdmin {
				userRole = portaineree.AdministratorRole
			}

			if err := handler.updateUserRole(user, userRole); err != nil {
				return resp,
					&httperror.HandlerError{StatusCode: http.StatusUnprocessableEntity, Message: "Failed to update user role", Err: err}
			}
		}
	}

	err = handler.addUserIntoTeams(user, ldapSettings)
	if err != nil {
		log.Warn().Err(err).Msg("unable to automatically add user into teams")
	}

	err = handler.AuthorizationService.UpdateUserAuthorizations(user.ID)
	if err != nil {
		return resp, httperror.InternalServerError("Unable to update user authorizations", err)
	}

	info := handler.LicenseService.Info()

	if user.Role != portaineree.AdministratorRole {
		if !info.Valid {
			return resp, httperror.Forbidden("License is not valid", httperrors.ErrNoValidLicense)
		}

		if handler.LicenseService.ShouldEnforceOveruse() {
			// If the free subscription license is enforced, the LDAP standard user are not allowed to log in
			return resp, &httperror.HandlerError{StatusCode: http.StatusPaymentRequired, Message: "Node limit exceeds the 5 node free license, please contact your administrator", Err: httperrors.ErrLicenseOverused}
		}
	}

	return handler.writeToken(w, user, resp.Method, false)
}

func (handler *Handler) authenticateInternal(w http.ResponseWriter, user *portaineree.User, password string) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method:   portaineree.AuthenticationInternal,
		Username: user.Username,
	}

	err := handler.CryptoService.CompareHashAndData(user.Password, password)
	if err != nil {
		return resp,
			&httperror.HandlerError{StatusCode: http.StatusUnprocessableEntity, Message: "Invalid credentials", Err: httperrors.ErrUnauthorized}
	}

	info := handler.LicenseService.Info()

	if user.Role != portaineree.AdministratorRole {
		if !info.Valid {
			return resp, httperror.Forbidden("License is not valid", httperrors.ErrNoValidLicense)
		}

		if handler.LicenseService.ShouldEnforceOveruse() {
			// If the free subscription license is enforced, the standard user are not allowed to log in
			return resp, &httperror.HandlerError{StatusCode: http.StatusPaymentRequired, Message: "Node limit exceeds the 5 node free license, please contact your administrator", Err: httperrors.ErrLicenseOverused}
		}
	}

	forceChangePassword := !handler.passwordStrengthChecker.Check(password)

	return handler.writeToken(w, user, resp.Method, forceChangePassword)
}

func (handler *Handler) writeToken(
	w http.ResponseWriter,
	user *portaineree.User,
	method portaineree.AuthenticationMethod,
	forceChangePassword bool,
) (*authMiddlewareResponse, *httperror.HandlerError) {
	tokenData := composeTokenData(user, forceChangePassword)

	return handler.persistAndWriteToken(w, tokenData, nil, method)
}

func (handler *Handler) persistAndWriteToken(w http.ResponseWriter, tokenData *portaineree.TokenData, expiryTime *time.Time, method portaineree.AuthenticationMethod) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Username: tokenData.Username,
		Method:   method,
	}

	var token string
	var err error

	if method == portaineree.AuthenticationOAuth {
		token, err = handler.JWTService.GenerateTokenForOAuth(tokenData, expiryTime)
		if err != nil {
			return resp, httperror.InternalServerError("Unable to generate JWT token for OAuth", err)
		}
	} else {
		token, err = handler.JWTService.GenerateToken(tokenData)
		if err != nil {
			return resp, httperror.InternalServerError("Unable to generate JWT token", err)
		}
	}
	return resp, response.JSON(w, &authenticateResponse{JWT: token})

}

func (handler *Handler) addUserIntoTeams(user *portaineree.User, settings *portaineree.LDAPSettings) error {
	teams, err := handler.DataStore.Team().Teams()
	if err != nil {
		return err
	}

	userGroups, err := handler.LDAPService.GetUserGroups(user.Username, settings, false)
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

			membership := &portaineree.TeamMembership{
				UserID: user.ID,
				TeamID: team.ID,
				Role:   portaineree.TeamMember,
			}

			err := handler.DataStore.TeamMembership().Create(membership)
			if err != nil {

				return err
			}
		}
	}

	return nil
}

func isLDAPAdmin(username string, ldapService portaineree.LDAPService, ldapSettings *portaineree.LDAPSettings) (bool, error) {
	//get groups the user belongs to
	userGroups, err := ldapService.GetUserGroups(username, ldapSettings, true)
	if err != nil {
		return false, errors.Wrap(err, "failed to retrieve user groups from LDAP server")
	}

	//convert the AdminGroups recorded in LDAP Settings to a map
	adminGroupsMap := make(map[string]bool)
	for _, adminGroup := range ldapSettings.AdminGroups {
		adminGroupsMap[adminGroup] = true
	}

	//check if any of the user groups matches the admin group records
	for _, userGroup := range userGroups {
		if adminGroupsMap[userGroup] {
			return true, nil
		}
	}
	return false, nil
}

func (handler *Handler) updateUserRole(user *portaineree.User, role portaineree.UserRole) error {
	user.Role = role
	err := handler.DataStore.User().UpdateUser(user.ID, user)
	return errors.Wrap(err, "unable to update user role inside the database")
}

func teamExists(teamName string, ldapGroups []string) bool {
	for _, group := range ldapGroups {
		if strings.ToLower(group) == strings.ToLower(teamName) {
			return true
		}
	}

	return false
}

func teamMembershipExists(teamID portaineree.TeamID, memberships []portaineree.TeamMembership) bool {
	for _, membership := range memberships {
		if membership.TeamID == teamID {
			return true
		}
	}

	return false
}

func composeTokenData(user *portaineree.User, forceChangePassword bool) *portaineree.TokenData {
	return &portaineree.TokenData{
		ID:                  user.ID,
		Username:            user.Username,
		Role:                user.Role,
		ForceChangePassword: forceChangePassword,
	}
}

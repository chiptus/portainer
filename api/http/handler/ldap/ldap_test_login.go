package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
)

type testLoginPayload struct {
	LDAPSettings portaineree.LDAPSettings
	Username     string
	Password     string
}

type testLoginResponse struct {
	Valid bool `json:"valid"`
}

func (payload *testLoginPayload) Validate(r *http.Request) error {
	if len(payload.LDAPSettings.URLs) == 0 {
		return errors.New("Invalid LDAP URLs. At least one URL is required")
	}

	return nil
}

// @id LDAPTestLogin
// @summary Test Login to ldap server
// @description
// @description **Access policy**: administrator
// @tags ldap
// @security ApiKeyAuth
// @security jwt
// @accept json
// @param body body testLoginPayload true "details"
// @success 200 {object} testLoginResponse "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /ldap/test [post]
func (handler *Handler) ldapTestLogin(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload testLoginPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch default settings", err)
	}

	err = handler.LDAPService.AuthenticateUser(payload.Username, payload.Password, settings)
	if err != nil && err != httperrors.ErrUnauthorized {
		return httperror.InternalServerError("Unable to test user authorization", err)
	}

	return response.JSON(w, &testLoginResponse{Valid: err != httperrors.ErrUnauthorized})

}

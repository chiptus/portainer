package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

type usersPayload struct {
	LDAPSettings portainer.LDAPSettings
}

func (payload *usersPayload) Validate(r *http.Request) error {
	if len(payload.LDAPSettings.URLs) == 0 {
		return errors.New("Invalid LDAP URLs. At least one URL is required")
	}

	return nil
}

// @id LDAPUsers
// @summary Search LDAP Users
// @description
// @description **Access policy**: administrator
// @tags ldap
// @security jwt
// @accept json
// @param body body usersPayload true "details"
// @success 200 {array} portainer.LDAPUser "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /ldap/users [post]
func (handler *Handler) ldapUsers(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload usersPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to fetch default settings", err}
	}

	users, err := handler.LDAPService.SearchUsers(settings)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to search for users", err}
	}

	return response.JSON(w, users)
}

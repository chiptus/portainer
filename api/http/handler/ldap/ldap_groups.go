package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

type groupsPayload struct {
	LDAPSettings portainer.LDAPSettings
}

func (payload *groupsPayload) Validate(r *http.Request) error {
	if len(payload.LDAPSettings.URLs) == 0 {
		return errors.New("Invalid LDAP URLs. At least one URL is required")
	}
	return nil
}

// @id LDAPGroups
// @summary Search LDAP Groups
// @description
// @description **Access policy**: administrator
// @tags ldap
// @security ApiKeyAuth
// @security jwt
// @accept json
// @param body body groupsPayload true "details"
// @success 200 {array} portainer.LDAPUser "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /ldap/groups [post]
func (handler *Handler) ldapGroups(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload groupsPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch default settings", Err: err}
	}

	groups, err := handler.LDAPService.SearchGroups(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to search for groups", Err: err}
	}

	return response.JSON(w, groups)
}

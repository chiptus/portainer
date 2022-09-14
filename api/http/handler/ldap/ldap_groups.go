package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type groupsPayload struct {
	LDAPSettings portaineree.LDAPSettings
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
// @success 200 {array} portaineree.LDAPUser "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /ldap/groups [post]
func (handler *Handler) ldapGroups(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload groupsPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch default settings", err)
	}

	groups, err := handler.LDAPService.SearchGroups(settings)
	if err != nil {
		return httperror.InternalServerError("Unable to search for groups", err)
	}

	return response.JSON(w, groups)
}

package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

type adminGroupsPayload struct {
	LDAPSettings portainer.LDAPSettings
}

func (payload *adminGroupsPayload) Validate(r *http.Request) error {
	if len(payload.LDAPSettings.URLs) == 0 {
		return errors.New("Invalid LDAP URLs. At least one URL is required")
	}
	if len(payload.LDAPSettings.AdminGroupSearchSettings) == 0 {
		return errors.New("Invalid AdminGroupSearchSettings. When 'AdminAutoPopulate' is true, at least one search settings is required")
	}
	return nil
}

// @id LDAPAdminGroups
// @summary Fetch LDAP admin groups
// @description Fetch LDAP admin groups from LDAP server based on AdminGroupSearchSettings
// @description **Access policy**: administrator
// @tags ldap
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body adminGroupsPayload true "LDAPSettings"
// @success 200 {array} string "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /ldap/admin-groups [post]
func (handler *Handler) ldapAdminGroups(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload adminGroupsPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch default settings", Err: err}
	}

	groups, err := handler.LDAPService.SearchAdminGroups(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to search admin groups", Err: err}
	}

	return response.JSON(w, groups)
}

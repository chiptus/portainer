package registries

import (
	"net/http"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id RegistryList
// @summary List Registries
// @description List all registries based on the current user authorizations.
// @description Will return all registries if using an administrator account otherwise it
// @description will only return authorized registries.
// @description **Access policy**: restricted
// @tags registries
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} portaineree.Registry "Success"
// @failure 500 "Server error"
// @router /registries [get]
func (handler *Handler) registryList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registries, err := handler.DataStore.Registry().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve registries from the database", err)
	}

	return response.JSON(w, registries)
}

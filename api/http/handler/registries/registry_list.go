package registries

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id RegistryList
// @summary List Registries
// @description List all registries based on the current user authorizations.
// @description Will return all registries if using an administrator account otherwise it
// @description will only return authorized registries.
// @description **Access policy**: restricted
// @tags registries
// @security jwt
// @produce json
// @success 200 {array} portainer.Registry "Success"
// @failure 500 "Server error"
// @router /registries [get]
func (handler *Handler) registryList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve registries from the database", err}
	}

	return response.JSON(w, registries)
}

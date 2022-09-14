package registries

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id RegistryInspect
// @summary Inspect a registry
// @description Retrieve details about a registry.
// @description **Access policy**: restricted
// @tags registries
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Registry identifier"
// @success 200 {object} portaineree.Registry "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access registry"
// @failure 404 "Registry not found"
// @failure 500 "Server error"
// @router /registries/{id} [get]
func (handler *Handler) registryInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	hasAccess, isEndpointAdmin, err := handler.userHasRegistryAccess(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}
	if !hasAccess {
		return httperror.Forbidden("Access denied to resource", httperrors.ErrResourceAccessDenied)
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid registry identifier route variable", err)
	}

	registry, err := handler.DataStore.Registry().Registry(portaineree.RegistryID(registryID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a registry with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a registry with the specified identifier inside the database", err)
	}

	hideFields(registry, !isEndpointAdmin)
	return response.JSON(w, registry)
}

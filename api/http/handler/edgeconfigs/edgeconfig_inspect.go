package edgeconfigs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

// @id EdgeConfigInspect
// @summary Inspect an Edge configuration
// @description Retrieve details about an Edge configuration.
// @description **Access policy**: authenticated
// @tags edge_configs
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Edge configuration identifier"
// @success 200 {object} portaineree.EdgeConfig "Success"
// @failure 400 "Invalid request"
// @failure 404 "Edge configuration not found"
// @failure 500 "Server error"
// @router /edge_configurations/{id} [get]
func (h *Handler) edgeConfigInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeConfigID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge configuration identifier route variable", err)
	}

	edgeConfiguration, err := h.dataStore.EdgeConfig().Read(portaineree.EdgeConfigID(edgeConfigID))
	if h.dataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an edge configuration with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an edge configuration with the specified identifier inside the database", err)
	}

	return response.JSON(w, edgeConfiguration)
}

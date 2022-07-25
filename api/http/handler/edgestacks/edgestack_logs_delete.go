package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

// @id EdgeStackLogsDelete
// @summary Deletes the available logs for a given edge stack and endpoint
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @param id path string true "EdgeStack Id"
// @param endpoint_id path string true "Endpoint Id"
// @success 204
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/logs/{endpoint_id} [delete]
func (handler *Handler) edgeStackLogsDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "endpoint_id")
	if err != nil {
		return httperror.BadRequest("Invalid endpoint identifier route variable", err)
	}

	err = handler.DataStore.EdgeStackLog().Delete(portaineree.EdgeStackID(edgeStackID), portaineree.EndpointID(endpointID))
	if err != nil {
		return httperror.BadRequest("Could not delete the logs", err)
	}

	return response.Empty(w)
}

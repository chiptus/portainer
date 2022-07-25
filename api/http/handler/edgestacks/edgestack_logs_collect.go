package edgestacks

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

// @id EdgeStackLogsCollect
// @summary Schedule the collection of logs for a given endpoint and edge stack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @param id path string true "EdgeStack Id"
// @param endpoint_id path string true "Environment Id"
// @param tail query int false "Number of lines to request for the logs"
// @success 204
// @failure 400
// @failure 404
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/logs/{endpoint_id} [put]
func (handler *Handler) edgeStackLogsCollect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "endpoint_id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	tail := 50

	t, _ := request.RetrieveQueryParameter(r, "tail", true)
	if t != "" {
		tail, err = strconv.Atoi(t)
		if err != nil {
			return httperror.BadRequest("Invalid tail parameter", err)
		}
	}

	edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(edgeStackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("The edge stack was not found", err)
	} else if err != nil {
		return httperror.InternalServerError("Could not retrieve the edge stack from the database", err)
	}

	err = handler.edgeService.AddLogCommand(edgeStack, portaineree.EndpointID(endpointID), tail)
	if err != nil {
		return httperror.InternalServerError("Could not store the log collection request", err)
	}

	err = handler.DataStore.EdgeStackLog().Create(&portaineree.EdgeStackLog{
		EdgeStackID: edgeStack.ID,
		EndpointID:  portaineree.EndpointID(endpointID),
	})
	if err != nil {
		return httperror.InternalServerError("Could not store the log collection status", err)
	}

	return response.Empty(w)
}

package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type edgeStackLogsStatusResponse struct {
	Status string `json:"status"`
}

// @id EdgeStackLogsStatusGet
// @summary Gets the status of the log collection for a given edgestack and environment
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @param id path string true "EdgeStack Id"
// @param endpoint_id path string true "Environment Id"
// @success 200
// @failure 400
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/logs/{endpoint_id} [get]
func (handler *Handler) edgeStackLogsStatusGet(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "endpoint_id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	edgeStackLog, err := handler.DataStore.EdgeStackLog().EdgeStackLog(portaineree.EdgeStackID(edgeStackID), portaineree.EndpointID(endpointID))

	resp := edgeStackLogsStatusResponse{"collected"}

	if err != nil {
		if handler.DataStore.IsErrObjectNotFound(err) {
			resp.Status = "idle"
		} else {
			return httperror.InternalServerError("Could not retrieve the edge stack log from the database", err)
		}
	} else {
		if len(edgeStackLog.Logs) == 0 {
			resp.Status = "pending"
		}
	}

	return response.JSON(w, resp)
}

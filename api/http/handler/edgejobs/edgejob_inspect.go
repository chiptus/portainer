package edgejobs

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type edgeJobInspectResponse struct {
	*portaineree.EdgeJob
	Endpoints []portaineree.EndpointID
}

// @id EdgeJobInspect
// @summary Inspect an EdgeJob
// @description **Access policy**: administrator
// @tags edge_jobs
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "EdgeJob Id"
// @success 200 {object} portaineree.EdgeJob
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id} [get]
func (handler *Handler) edgeJobInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge job identifier route variable", err)
	}

	edgeJob, err := handler.DataStore.EdgeJob().Read(portaineree.EdgeJobID(edgeJobID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an Edge job with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an Edge job with the specified identifier inside the database", err)
	}

	responseObj := edgeJobInspectResponse{
		EdgeJob: edgeJob,
	}

	for endpointID := range edgeJob.Endpoints {
		responseObj.Endpoints = append(responseObj.Endpoints, endpointID)
	}

	return response.JSON(w, responseObj)
}

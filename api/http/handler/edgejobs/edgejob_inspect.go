package edgejobs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
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
// @param id path string true "EdgeJob Id"
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

	edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(portaineree.EdgeJobID(edgeJobID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an Edge job with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an Edge job with the specified identifier inside the database", err)
	}

	endpointIDs := []portaineree.EndpointID{}

	for endpointID := range edgeJob.Endpoints {
		endpointIDs = append(endpointIDs, endpointID)
	}

	responseObj := edgeJobInspectResponse{
		EdgeJob:   edgeJob,
		Endpoints: endpointIDs,
	}

	return response.JSON(w, responseObj)
}

package edgejobs

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/slices"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeJobTasksCollect
// @summary Collect the log for a specifc task on an EdgeJob
// @description **Access policy**: administrator
// @tags edge_jobs
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "EdgeJob Id"
// @param taskID path int true "Task Id"
// @success 204
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id}/tasks/{taskID}/logs [post]
func (handler *Handler) edgeJobTasksCollect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge job identifier route variable", err)
	}

	taskID, err := request.RetrieveNumericRouteVariableValue(r, "taskID")
	if err != nil {
		return httperror.BadRequest("Invalid Task identifier route variable", err)
	}

	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		edgeJob, err := tx.EdgeJob().Read(portaineree.EdgeJobID(edgeJobID))
		if tx.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find an Edge job with the specified identifier inside the database", err)
		} else if err != nil {
			return httperror.InternalServerError("Unable to find an Edge job with the specified identifier inside the database", err)
		}

		endpointID := portaineree.EndpointID(taskID)
		endpointsFromGroups, err := edge.GetEndpointsFromEdgeGroups(edgeJob.EdgeGroups, tx)
		if err != nil {
			return httperror.InternalServerError("Unable to get Endpoints from EdgeGroups", err)
		}

		if slices.Contains(endpointsFromGroups, endpointID) {
			edgeJob.GroupLogsCollection[endpointID] = portaineree.EdgeJobEndpointMeta{
				CollectLogs: true,
				LogsStatus:  portaineree.EdgeJobLogsStatusPending,
			}
		} else {
			meta := edgeJob.Endpoints[endpointID]
			meta.CollectLogs = true
			meta.LogsStatus = portaineree.EdgeJobLogsStatusPending
			edgeJob.Endpoints[endpointID] = meta
		}

		err = tx.EdgeJob().Update(edgeJob.ID, edgeJob)
		if err != nil {
			return httperror.InternalServerError("Unable to persist Edge job changes in the database", err)
		}

		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environment from the database", err)
		}

		if endpoint.Edge.AsyncMode {
			edgeJobFileContent, err := handler.FileService.GetFileContent(edgeJob.ScriptPath, "")
			if err != nil {
				return httperror.InternalServerError("Unable to retrieve Edge job script file from disk", err)
			}
			err = handler.edgeService.ReplaceJobCommandTx(tx, endpoint.ID, *edgeJob, edgeJobFileContent)
			if err != nil {
				return httperror.InternalServerError("Unable to persist edge job changes to the database", err)
			}
		} else {
			handler.ReverseTunnelService.AddEdgeJob(endpoint, edgeJob)
		}

		return nil
	})

	if err != nil {
		var handlerError *httperror.HandlerError
		if errors.As(err, &handlerError) {
			return handlerError
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

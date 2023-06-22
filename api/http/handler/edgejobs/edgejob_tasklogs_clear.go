package edgejobs

import (
	"errors"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/portainer/portainer/pkg/featureflags"
)

// @id EdgeJobTasksClear
// @summary Clear the log for a specifc task on an EdgeJob
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
// @router /edge_jobs/{id}/tasks/{taskID}/logs [delete]
func (handler *Handler) edgeJobTasksClear(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge job identifier route variable", err)
	}

	taskID, err := request.RetrieveNumericRouteVariableValue(r, "taskID")
	if err != nil {
		return httperror.BadRequest("Invalid Task identifier route variable", err)
	}

	mutationFn := func(edgeJob *portaineree.EdgeJob, endpointID portaineree.EndpointID, endpointsFromGroups []portaineree.EndpointID) {
		if slices.Contains(endpointsFromGroups, endpointID) {
			edgeJob.GroupLogsCollection[endpointID] = portaineree.EdgeJobEndpointMeta{
				CollectLogs: false,
				LogsStatus:  portaineree.EdgeJobLogsStatusIdle,
			}
		} else {
			meta := edgeJob.Endpoints[endpointID]
			meta.CollectLogs = false
			meta.LogsStatus = portaineree.EdgeJobLogsStatusIdle
			edgeJob.Endpoints[endpointID] = meta
		}
	}

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		updateEdgeJobFn := func(edgeJob *portaineree.EdgeJob, endpointID portaineree.EndpointID, endpointsFromGroups []portaineree.EndpointID) error {
			return handler.DataStore.EdgeJob().UpdateEdgeJobFunc(edgeJob.ID, func(j *portaineree.EdgeJob) {
				mutationFn(j, endpointID, endpointsFromGroups)
			})
		}

		err = handler.clearEdgeJobTaskLogs(handler.DataStore, portaineree.EdgeJobID(edgeJobID), portaineree.EndpointID(taskID), updateEdgeJobFn)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			updateEdgeJobFn := func(edgeJob *portaineree.EdgeJob, endpointID portaineree.EndpointID, endpointsFromGroups []portaineree.EndpointID) error {
				mutationFn(edgeJob, endpointID, endpointsFromGroups)

				return tx.EdgeJob().Update(edgeJob.ID, edgeJob)
			}

			return handler.clearEdgeJobTaskLogs(tx, portaineree.EdgeJobID(edgeJobID), portaineree.EndpointID(taskID), updateEdgeJobFn)
		})
	}

	if err != nil {
		var handlerError *httperror.HandlerError
		if errors.As(err, &handlerError) {
			return handlerError
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

func (handler *Handler) clearEdgeJobTaskLogs(tx dataservices.DataStoreTx, edgeJobID portaineree.EdgeJobID, endpointID portaineree.EndpointID, updateEdgeJob func(*portaineree.EdgeJob, portaineree.EndpointID, []portaineree.EndpointID) error) error {
	edgeJob, err := tx.EdgeJob().Read(edgeJobID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an Edge job with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an Edge job with the specified identifier inside the database", err)
	}

	err = handler.FileService.ClearEdgeJobTaskLogs(strconv.Itoa(int(edgeJobID)), strconv.Itoa(int(endpointID)))
	if err != nil {
		return httperror.InternalServerError("Unable to clear log file from disk", err)
	}

	endpointsFromGroups, err := edge.GetEndpointsFromEdgeGroups(edgeJob.EdgeGroups, tx)
	if err != nil {
		return httperror.InternalServerError("Unable to get Endpoints from EdgeGroups", err)
	}

	err = updateEdgeJob(edgeJob, endpointID, endpointsFromGroups)
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
}

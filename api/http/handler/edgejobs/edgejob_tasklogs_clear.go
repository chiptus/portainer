package edgejobs

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id EdgeJobTasksClear
// @summary Clear the log for a specifc task on an EdgeJob
// @description **Access policy**: administrator
// @tags edge_jobs
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path string true "EdgeJob Id"
// @param taskID path string true "Task Id"
// @success 204
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id}/tasks/{taskID}/logs [delete]
func (handler *Handler) edgeJobTasksClear(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge job identifier route variable", err}
	}

	taskID, err := request.RetrieveNumericRouteVariableValue(r, "taskID")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Task identifier route variable", err}
	}

	edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(portaineree.EdgeJobID(edgeJobID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an Edge job with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an Edge job with the specified identifier inside the database", err}
	}

	endpointID := portaineree.EndpointID(taskID)

	meta := edgeJob.Endpoints[endpointID]
	meta.CollectLogs = false
	meta.LogsStatus = portaineree.EdgeJobLogsStatusIdle
	edgeJob.Endpoints[endpointID] = meta

	err = handler.FileService.ClearEdgeJobTaskLogs(strconv.Itoa(edgeJobID), strconv.Itoa(taskID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to clear log file from disk", err}
	}

	err = handler.DataStore.EdgeJob().UpdateEdgeJob(edgeJob.ID, edgeJob)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Edge job changes in the database", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve environment from the database", err}
	}

	if endpoint.Edge.AsyncMode {
		edgeJobFileContent, err := handler.FileService.GetFileContent(edgeJob.ScriptPath, "")
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Edge job script file from disk", err}
		}
		err = handler.edgeService.ReplaceJobCommand(endpoint.ID, *edgeJob, edgeJobFileContent)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist edge job changes to the database", err}
		}
	} else {
		handler.ReverseTunnelService.AddEdgeJob(endpointID, edgeJob)
	}

	return response.Empty(w)
}

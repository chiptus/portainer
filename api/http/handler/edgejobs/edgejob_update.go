package edgejobs

import (
	"errors"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"

	"github.com/asaskevich/govalidator"
)

type edgeJobUpdatePayload struct {
	Name           *string
	CronExpression *string
	Recurring      *bool
	Endpoints      []portaineree.EndpointID
	FileContent    *string
}

func (payload *edgeJobUpdatePayload) Validate(r *http.Request) error {
	if payload.Name != nil && !govalidator.Matches(*payload.Name, `^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`) {
		return errors.New("Invalid Edge job name format. Allowed characters are: [a-zA-Z0-9_.-]")
	}
	return nil
}

// @id EdgeJobUpdate
// @summary Update an EdgeJob
// @description **Access policy**: administrator
// @tags edge_jobs
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeJob Id"
// @param body body edgeJobUpdatePayload true "EdgeGroup data"
// @success 200 {object} portaineree.EdgeJob
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id} [post]
func (handler *Handler) edgeJobUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge job identifier route variable", err)
	}

	var payload edgeJobUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(portaineree.EdgeJobID(edgeJobID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an Edge job with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an Edge job with the specified identifier inside the database", err)
	}

	err = handler.updateEdgeSchedule(edgeJob, &payload)
	if err != nil {
		return httperror.InternalServerError("Unable to update Edge job", err)
	}

	err = handler.DataStore.EdgeJob().UpdateEdgeJob(edgeJob.ID, edgeJob)
	if err != nil {
		return httperror.InternalServerError("Unable to persist Edge job changes inside the database", err)
	}

	return response.JSON(w, edgeJob)
}

func (handler *Handler) updateEdgeSchedule(edgeJob *portaineree.EdgeJob, payload *edgeJobUpdatePayload) error {
	if payload.Name != nil {
		edgeJob.Name = *payload.Name
	}

	endpointsToAdd := map[portaineree.EndpointID]bool{}
	endpointsToRemove := map[portaineree.EndpointID]bool{}

	if payload.Endpoints != nil {
		endpointsMap := map[portaineree.EndpointID]portaineree.EdgeJobEndpointMeta{}

		newEndpoints := endpointutils.EndpointSet(payload.Endpoints)
		for endpointID := range edgeJob.Endpoints {
			if !newEndpoints[endpointID] {
				endpointsToRemove[endpointID] = true
			}
		}

		for _, endpointID := range payload.Endpoints {
			endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
			if err != nil {
				return err
			}

			if !endpointutils.IsEdgeEndpoint(endpoint) {
				continue
			}

			if meta, exists := edgeJob.Endpoints[endpointID]; exists {
				endpointsMap[endpointID] = meta
			} else {
				endpointsMap[endpointID] = portaineree.EdgeJobEndpointMeta{}
				endpointsToAdd[endpointID] = true
			}
		}

		edgeJob.Endpoints = endpointsMap
	}

	updateVersion := false
	if payload.CronExpression != nil && *payload.CronExpression != edgeJob.CronExpression {
		edgeJob.CronExpression = *payload.CronExpression
		updateVersion = true
	}

	fileContent, err := handler.FileService.GetFileContent(edgeJob.ScriptPath, "")
	if err != nil {
		return err
	}

	if payload.FileContent != nil && *payload.FileContent != string(fileContent) {
		fileContent = []byte(*payload.FileContent)
		_, err := handler.FileService.StoreEdgeJobFileFromBytes(strconv.Itoa(int(edgeJob.ID)), fileContent)
		if err != nil {
			return err
		}

		updateVersion = true
	}

	if payload.Recurring != nil && *payload.Recurring != edgeJob.Recurring {
		edgeJob.Recurring = *payload.Recurring
		updateVersion = true
	}

	if updateVersion {
		edgeJob.Version++
	}

	for endpointID := range edgeJob.Endpoints {
		handler.ReverseTunnelService.AddEdgeJob(endpointID, edgeJob)

		needsAddOperation := endpointsToAdd[endpointID]
		if needsAddOperation || updateVersion {
			err := handler.storeEdgeAsyncCommand(edgeJob, endpointID, fileContent, needsAddOperation)
			if err != nil {
				return err
			}
		}

	}

	for endpointID := range endpointsToRemove {
		err := handler.edgeService.RemoveJobCommand(endpointID, edgeJob.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) storeEdgeAsyncCommand(edgeJob *portaineree.EdgeJob, endpointID portaineree.EndpointID, fileContent []byte, needsAddOperation bool) error {
	if needsAddOperation {
		return handler.edgeService.AddJobCommand(endpointID, *edgeJob, fileContent)
	}
	return handler.edgeService.ReplaceJobCommand(endpointID, *edgeJob, fileContent)
}

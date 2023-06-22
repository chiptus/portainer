package edgejobs

import (
	"errors"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/maps"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/portainer/portainer/pkg/featureflags"

	"github.com/asaskevich/govalidator"
)

type edgeJobUpdatePayload struct {
	Name           *string
	CronExpression *string
	Recurring      *bool
	Endpoints      []portaineree.EndpointID
	EdgeGroups     []portaineree.EdgeGroupID
	FileContent    *string
}

func (payload *edgeJobUpdatePayload) Validate(r *http.Request) error {
	if payload.Name != nil && !govalidator.Matches(*payload.Name, `^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`) {
		return errors.New("invalid Edge job name format. Allowed characters are: [a-zA-Z0-9_.-]")
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
// @param id path int true "EdgeJob Id"
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

	var edgeJob *portaineree.EdgeJob
	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		edgeJob, err = handler.updateEdgeJob(handler.DataStore, portaineree.EdgeJobID(edgeJobID), payload)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			edgeJob, err = handler.updateEdgeJob(tx, portaineree.EdgeJobID(edgeJobID), payload)
			return err
		})
	}

	return txResponse(w, edgeJob, err)
}

func (handler *Handler) updateEdgeJob(tx dataservices.DataStoreTx, edgeJobID portaineree.EdgeJobID, payload edgeJobUpdatePayload) (*portaineree.EdgeJob, error) {
	edgeJob, err := tx.EdgeJob().Read(portaineree.EdgeJobID(edgeJobID))
	if tx.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find an Edge job with the specified identifier inside the database", err)
	} else if err != nil {
		return nil, httperror.InternalServerError("Unable to find an Edge job with the specified identifier inside the database", err)
	}

	err = handler.updateEdgeSchedule(tx, edgeJob, &payload)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to update Edge job", err)
	}

	err = tx.EdgeJob().Update(edgeJob.ID, edgeJob)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to persist Edge job changes inside the database", err)
	}

	return edgeJob, nil
}

func (handler *Handler) updateEdgeSchedule(tx dataservices.DataStoreTx, edgeJob *portaineree.EdgeJob, payload *edgeJobUpdatePayload) error {
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
			endpoint, err := tx.Endpoint().Endpoint(endpointID)
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

	if len(payload.EdgeGroups) == 0 && len(edgeJob.EdgeGroups) > 0 {
		endpoints, err := edge.GetEndpointsFromEdgeGroups(edgeJob.EdgeGroups, tx)
		if err != nil {
			return errors.New("unable to get endpoints from edge groups")
		}

		for _, endpointID := range endpoints {
			endpointsToRemove[portaineree.EndpointID(endpointID)] = true
		}

		edgeJob.EdgeGroups = nil
	}

	edgeGroupsToAdd := []portaineree.EdgeGroupID{}
	edgeGroupsToRemove := []portaineree.EdgeGroupID{}
	endpointsFromGroupsToAddMap := map[portaineree.EndpointID]portaineree.EdgeJobEndpointMeta{}

	if len(payload.EdgeGroups) > 0 {
		for _, edgeGroupID := range payload.EdgeGroups {
			_, err := tx.EdgeGroup().Read(edgeGroupID)
			if err != nil {
				return err
			}

			if !slices.Contains(edgeJob.EdgeGroups, edgeGroupID) {
				edgeGroupsToAdd = append(edgeGroupsToAdd, edgeGroupID)
			}
		}

		endpointsFromGroupsToAdd, err := edge.GetEndpointsFromEdgeGroups(edgeGroupsToAdd, tx)
		if err != nil {
			return errors.New("unable to get endpoints from edge groups")
		}
		endpointsFromGroupsToAddMap = convertEndpointsToMetaObject(endpointsFromGroupsToAdd)

		for endpointID := range endpointsFromGroupsToAddMap {
			endpointsToAdd[endpointID] = true
		}

		newEdgeGroups := edge.EdgeGroupSet(payload.EdgeGroups)
		for _, edgeGroupID := range edgeJob.EdgeGroups {
			if !newEdgeGroups[edgeGroupID] {
				edgeGroupsToRemove = append(edgeGroupsToRemove, edgeGroupID)
			}
		}

		endpointsFromGroupsToRemove, err := edge.GetEndpointsFromEdgeGroups(edgeGroupsToRemove, tx)
		if err != nil {
			return errors.New("unable to get endpoints from edge groups")
		}

		endpointsToRemoveMap := convertEndpointsToMetaObject(endpointsFromGroupsToRemove)

		for endpointID := range endpointsToRemoveMap {
			endpointsToRemove[endpointID] = true
		}

		edgeJob.EdgeGroups = payload.EdgeGroups
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

	maps.Copy(endpointsFromGroupsToAddMap, edgeJob.Endpoints)

	for endpointID := range endpointsFromGroupsToAddMap {
		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return err
		}

		handler.ReverseTunnelService.AddEdgeJob(endpoint, edgeJob)

		needsAddOperation := endpointsToAdd[endpointID]
		if needsAddOperation || updateVersion {
			err := handler.storeEdgeAsyncCommand(tx, edgeJob, endpointID, fileContent, needsAddOperation)
			if err != nil {
				return err
			}
		}

	}

	for endpointID := range endpointsToRemove {
		handler.ReverseTunnelService.RemoveEdgeJobFromEndpoint(endpointID, edgeJob.ID)

		err := handler.edgeService.RemoveJobCommandTx(tx, endpointID, edgeJob.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) storeEdgeAsyncCommand(tx dataservices.DataStoreTx, edgeJob *portaineree.EdgeJob, endpointID portaineree.EndpointID, fileContent []byte, needsAddOperation bool) error {
	if needsAddOperation {
		return handler.edgeService.AddJobCommandTx(tx, endpointID, *edgeJob, fileContent)
	}
	return handler.edgeService.ReplaceJobCommandTx(tx, endpointID, *edgeJob, fileContent)
}

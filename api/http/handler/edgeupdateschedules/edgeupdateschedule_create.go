package edgeupdateschedules

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/set"
)

type createPayload struct {
	Name          string
	GroupIDs      []portaineree.EdgeGroupID
	Type          edgetypes.UpdateScheduleType
	Version       string
	ScheduledTime string
}

func (payload *createPayload) Validate(r *http.Request) error {
	if payload.Name == "" {
		return errors.New("invalid name")
	}

	if len(payload.GroupIDs) == 0 {
		return errors.New("required to choose at least one group")
	}

	if !slices.Contains([]edgetypes.UpdateScheduleType{edgetypes.UpdateScheduleRollback, edgetypes.UpdateScheduleUpdate}, payload.Type) {
		return errors.New("invalid schedule type")
	}

	if payload.Version == "" {
		return errors.New("Invalid version")
	}

	if payload.ScheduledTime == "" {
		return errors.New("Scheduled time is required")
	}

	return nil
}

// @id EdgeUpdateScheduleCreate
// @summary Creates a new Edge Update Schedule
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @accept json
// @param body body createPayload true "Schedule details"
// @produce json
// @success 200 {object} edgetypes.UpdateSchedule
// @failure 500
// @router /edge_update_schedules [post]
func (handler *Handler) create(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	var payload createPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	err = handler.validateUniqueName(payload.Name, 0)
	if err != nil {
		return httperror.NewError(http.StatusConflict, "Edge update schedule name already in use", err)

	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user information from token", err)
	}

	var edgeStackID portaineree.EdgeStackID
	var scheduleID edgetypes.UpdateScheduleID
	needCleanup := true
	defer func() {
		if !needCleanup {
			return
		}

		if scheduleID != 0 {
			err := handler.updateService.DeleteSchedule(scheduleID)
			if err != nil {
				log.Error().Err(err).Msg("Unable to cleanup edge update schedule")
			}
		}

		if edgeStackID != 0 {
			err = handler.edgeStacksService.DeleteEdgeStack(edgeStackID, payload.GroupIDs)

			if err != nil {
				log.Error().Err(err).Msg("Unable to cleanup edge stack")
			}
		}
	}()

	item := &edgetypes.UpdateSchedule{
		Name:    payload.Name,
		Version: payload.Version,

		Created:   time.Now().Unix(),
		CreatedBy: tokenData.ID,
		Type:      payload.Type,
	}

	relatedEnvironments, err := handler.fetchRelatedEnvironments(payload.GroupIDs)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch related environments", err)
	}

	err = handler.validateRelatedEnvironments(relatedEnvironments)
	if err != nil {
		return httperror.BadRequest("Environment is not supported for update", err)
	}

	previousVersions := handler.getPreviousVersions(relatedEnvironments)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch previous versions for related endpoints", err)
	}

	item.EnvironmentsPreviousVersions = previousVersions

	err = handler.updateService.CreateSchedule(item)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the edge update schedule", err)
	}

	scheduleID = item.ID

	edgeStackID, err = handler.createUpdateEdgeStack(item.ID, payload.GroupIDs, payload.Version, payload.ScheduledTime)
	if err != nil {
		return httperror.InternalServerError("Unable to create edge stack", err)
	}

	item.EdgeStackID = edgeStackID
	err = handler.updateService.UpdateSchedule(item.ID, item)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the edge update schedule", err)
	}

	needCleanup = false
	return response.JSON(w, item)
}

func (handler *Handler) validateRelatedEnvironments(relatedEnvironments []portaineree.Endpoint) error {
	if len(relatedEnvironments) == 0 {
		return errors.New("No related environments")
	}

	for _, environment := range relatedEnvironments {
		err := handler.isUpdateSupported(&environment)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) fetchRelatedEnvironments(edgeGroupIds []portaineree.EdgeGroupID) ([]portaineree.Endpoint, error) {
	relationConfig, err := edge.FetchEndpointRelationsConfig(handler.dataStore)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to fetch environment relations config")
	}

	relatedEnvironmentsIds, err := edge.EdgeStackRelatedEndpoints(edgeGroupIds, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to fetch related environments")
	}

	environments, err := handler.dataStore.Endpoint().Endpoints()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to fetch environments")
	}

	relatedEnvironmentIdsSet := set.ToSet(relatedEnvironmentsIds)

	relatedEnvironments := []portaineree.Endpoint{}

	for _, environment := range environments {
		if !relatedEnvironmentIdsSet.Contains(environment.ID) {
			continue
		}

		relatedEnvironments = append(relatedEnvironments, environment)
	}

	return relatedEnvironments, nil
}

func (handler *Handler) getPreviousVersions(relatedEnvironments []portaineree.Endpoint) map[portaineree.EndpointID]string {
	prevVersions := map[portaineree.EndpointID]string{}

	for _, environment := range relatedEnvironments {
		prevVersions[environment.ID] = environment.Agent.Version
	}

	return prevVersions
}

func (handler *Handler) isUpdateSupported(environment *portaineree.Endpoint) error {
	if !endpointutils.IsEdgeEndpoint(environment) {
		return errors.New("environment is not an edge endpoint, this feature is limited to edge endpoints")
	}

	if !endpointutils.IsDockerEndpoint(environment) {
		return errors.New("environment is not a docker endpoint, this feature is limited to docker endpoints")
	}

	snapshot, err := handler.dataStore.Snapshot().Snapshot(environment.ID)
	if err != nil {
		return errors.WithMessage(err, "unable to fetch snapshot")
	}

	if snapshot.Docker == nil {
		return errors.New("missing docker snapshot")
	}

	if snapshot.Docker.Swarm {
		return errors.New("swarm is not supported")
	}

	return nil
}

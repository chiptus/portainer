package edgeupdateschedules

import (
	"net/http"
	"slices"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/set"
	pslices "github.com/portainer/portainer-ee/api/internal/slices"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type createPayload struct {
	Name          string
	GroupIDs      []portaineree.EdgeGroupID
	Type          edgetypes.UpdateScheduleType
	Version       string
	ScheduledTime string
	RegistryID    portaineree.RegistryID
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
			err = handler.updateService.DeleteSchedule(scheduleID)
			if err != nil {
				log.Error().Err(err).Msg("Unable to cleanup edge update schedule")
			}
		}

		if edgeStackID != 0 {
			err = handler.edgeStacksService.DeleteEdgeStack(handler.dataStore, edgeStackID, payload.GroupIDs)
			if err != nil {
				log.Error().Err(err).Msg("Unable to cleanup edge stack")
			}
		}
	}()

	item := &edgetypes.UpdateSchedule{
		Name:         payload.Name,
		Version:      payload.Version,
		Created:      time.Now().Unix(),
		CreatedBy:    tokenData.ID,
		Type:         payload.Type,
		RegistryID:   payload.RegistryID,
		EdgeGroupIDs: payload.GroupIDs,
	}

	relatedEnvironments, err := handler.fetchRelatedEnvironments(payload.GroupIDs)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch related environments", err)
	}

	edgeEnvironmentType, err := handler.validateRelatedEnvironments(relatedEnvironments)
	if err != nil {
		return httperror.BadRequest("Fail to validate related environment", err)
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

	relatedEnvironmentsIDs := pslices.Map(relatedEnvironments, func(environment portaineree.Endpoint) portaineree.EndpointID {
		return environment.ID
	})

	edgeStackID, err = handler.createUpdateEdgeStack(
		item.ID,
		relatedEnvironmentsIDs,
		payload.RegistryID,
		payload.Version,
		payload.ScheduledTime,
		edgeEnvironmentType,
	)
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

func (handler *Handler) validateRelatedEnvironments(relatedEnvironments []portaineree.Endpoint) (portaineree.EndpointType, error) {
	if len(relatedEnvironments) == 0 {
		return 0, errors.New("No related environments")
	}

	first := relatedEnvironments[0].Type
	for _, environment := range relatedEnvironments {
		err := handler.isUpdateSupported(&environment)
		if err != nil {
			return 0, err
		}

		// Make sure that all environments in one edge group are same type
		if environment.Type != first {
			return 0, errors.New("Environment type is not unified")
		}
	}

	return first, nil
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

	if endpointutils.IsNomadEndpoint(environment) {
		// Nomad does not need to check snapshot
		return nil
	}

	if endpointutils.IsDockerEndpoint(environment) {
		snapshot, err := handler.dataStore.Snapshot().Read(environment.ID)
		if err != nil {
			handler.ReverseTunnelService.SetTunnelStatusToRequired(environment.ID)

			return errors.WithMessage(err, "unable to fetch snapshot, please try again later")
		}

		if snapshot.Docker == nil {
			handler.ReverseTunnelService.SetTunnelStatusToRequired(environment.ID)

			return errors.New("missing docker snapshot, please try again later")
		}

		if snapshot.Docker.Swarm {
			return errors.New("swarm is not supported")
		}

		return nil
	}

	return errors.New("environment is not a docker/nomad endpoint, this feature is limited to docker/nomad endpoints")
}

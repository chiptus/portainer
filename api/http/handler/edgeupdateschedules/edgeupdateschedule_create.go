package edgeupdateschedules

import (
	"net/http"
	"slices"
	"time"

	"github.com/portainer/portainer-ee/api/http/security"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type createPayload struct {
	Name          string
	GroupIDs      []portainer.EdgeGroupID
	Type          edgetypes.UpdateScheduleType
	Version       string
	ScheduledTime string
	RegistryID    portainer.RegistryID
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
		return httperror.Conflict("Edge update schedule name already in use", err)

	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user information from token", err)
	}

	var edgeStackID portainer.EdgeStackID
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

	relatedEnvironmentsIDs, previousVersions, envType, err := handler.filterEnvironments(payload.GroupIDs, payload.Version, payload.Type == edgetypes.UpdateScheduleRollback, 0)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch related environments", err)
	}

	item.EnvironmentsPreviousVersions = previousVersions

	err = handler.updateService.CreateSchedule(item)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the edge update schedule", err)
	}

	scheduleID = item.ID

	edgeStackID, err = handler.createUpdateEdgeStack(
		item.ID,
		relatedEnvironmentsIDs,
		payload.RegistryID,
		payload.Version,
		payload.ScheduledTime,
		envType,
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

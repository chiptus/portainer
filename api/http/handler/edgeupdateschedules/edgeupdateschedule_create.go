package edgeupdateschedules

import (
	"net/http"
	"slices"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/utils"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/pkg/errors"
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

	if payload.ScheduledTime != "" {
		scheduledTime, err := validateScheduleTime(payload.ScheduledTime)
		if err != nil {
			return err
		}

		payload.ScheduledTime = scheduledTime
	}

	return nil
}

func validateScheduleTime(scheduledTime string) (string, error) {
	_, err := time.Parse(portaineree.DateTimeFormat, scheduledTime)
	if err == nil {
		return scheduledTime, nil
	}

	timeWithSeconds := scheduledTime + ":00"
	_, err = time.Parse(portaineree.DateTimeFormat, timeWithSeconds)
	if err != nil {
		return "", errors.New("Invalid scheduled time")
	}

	return timeWithSeconds, nil

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

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user information from token", err)
	}

	item := &edgetypes.UpdateSchedule{
		Name:         payload.Name,
		Version:      payload.Version,
		Created:      time.Now().Unix(),
		CreatedBy:    tokenData.ID,
		Type:         payload.Type,
		RegistryID:   payload.RegistryID,
		EdgeGroupIDs: payload.GroupIDs,
	}

	err = handler.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {

		err = handler.validateUniqueName(tx, payload.Name, 0)
		if err != nil {
			return httperror.NewError(http.StatusConflict, "Edge update schedule name already in use", err)
		}

		relatedEnvironmentsIDs, environmentType, err := handler.filterEnvironments(tx, payload.GroupIDs, payload.Version, payload.Type == edgetypes.UpdateScheduleRollback)
		if err != nil {
			return httperror.InternalServerError("Unable to fetch related environments", err)
		}

		err = handler.updateService.CreateSchedule(tx, item, updateschedules.CreateMetadata{
			RelatedEnvironmentsIDs: relatedEnvironmentsIDs,
			ScheduledTime:          payload.ScheduledTime,
			EnvironmentType:        environmentType,
		})
		if err != nil {
			return httperror.InternalServerError("Unable to persist the edge update schedule", err)
		}

		return nil
	})

	return utils.TxResponse(w, item, err)
}

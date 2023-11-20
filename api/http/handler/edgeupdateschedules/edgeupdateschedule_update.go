package edgeupdateschedules

import (
	"errors"
	"net/http"
	"slices"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/utils"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
)

type updatePayload struct {
	Name          *string
	GroupIDs      []portainer.EdgeGroupID
	Type          *edgetypes.UpdateScheduleType
	Version       *string
	ScheduledTime *string
	RegistryID    *portainer.RegistryID
}

func (payload *updatePayload) Validate(r *http.Request) error {
	if payload.Name != nil && *payload.Name == "" {
		return errors.New("invalid name")
	}

	if payload.Type != nil && !slices.Contains([]edgetypes.UpdateScheduleType{edgetypes.UpdateScheduleRollback, edgetypes.UpdateScheduleUpdate}, *payload.Type) {
		return errors.New("invalid schedule type")
	}

	if payload.Version != nil && *payload.Version == "" {
		return errors.New("Invalid version")
	}

	if payload.ScheduledTime != nil {
		if *payload.ScheduledTime == "" {
			return errors.New("Scheduled time is required")
		}

		scheduledTime, err := validateScheduleTime(*payload.ScheduledTime)
		if err != nil {
			return err
		}

		payload.ScheduledTime = &scheduledTime
	}
	return nil
}

// @id EdgeUpdateScheduleUpdate
// @summary Updates an Edge Update Schedule
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @accept json
// @param body body updatePayload true "Schedule details"
// @produce json
// @success 204
// @failure 500
// @router /edge_update_schedules/{id} [post]
func (handler *Handler) update(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge Update identifier route variable", err)
	}

	payload, err := request.GetPayload[updatePayload](r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var item *edgetypes.UpdateSchedule
	err = handler.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		schedule, err := tx.EdgeUpdateSchedule().Read(edgetypes.UpdateScheduleID(id))
		if err != nil {
			status := http.StatusInternalServerError
			if dataservices.IsErrObjectNotFound(err) {
				status = http.StatusNotFound
			}
			return httperror.NewError(status, "Unable to find an Edge Update with the specified identifier inside the database", err)
		}

		item = schedule

		if payload.Name != nil && *payload.Name != schedule.Name {
			err = handler.validateUniqueName(tx, *payload.Name, schedule.ID)
			if err != nil {
				return httperror.NewError(http.StatusConflict, "Edge update schedule name already in use", err)
			}

			schedule.Name = *payload.Name
		}

		shouldUpdateRelations := payload.GroupIDs != nil || payload.Type != nil || payload.Version != nil || payload.ScheduledTime != nil || payload.RegistryID != nil
		if !shouldUpdateRelations {
			err = tx.EdgeUpdateSchedule().Update(schedule.ID, schedule)
			if err != nil {
				return httperror.InternalServerError("Unable to persist the edge update schedule", err)
			}

			return nil
		}

		stack, err := tx.EdgeStack().EdgeStack(schedule.EdgeStackID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Edge stack", err)
		}

		isActive := isUpdateActive(stack)

		if isActive {
			return httperror.BadRequest("Unable to update Edge update schedule", errors.New("edge stack is not in pending state"))
		}

		if payload.GroupIDs != nil {
			schedule.EdgeGroupIDs = payload.GroupIDs
		}

		if payload.Type != nil {
			schedule.Type = *payload.Type
		}

		if payload.Version != nil {
			schedule.Version = *payload.Version
		}

		scheduledTime := stack.ScheduledTime
		if payload.ScheduledTime != nil {
			scheduledTime = *payload.ScheduledTime
		}

		if payload.RegistryID != nil {
			schedule.RegistryID = *payload.RegistryID
		}

		relatedEnvironmentsIDs, environmentType, err := handler.filterEnvironments(tx, payload.GroupIDs, *payload.Version, schedule.Type == edgetypes.UpdateScheduleRollback)
		if err != nil {
			return httperror.InternalServerError("Unable to fetch related environments", err)
		}

		err = handler.updateService.UpdateSchedule(tx, schedule.ID, schedule, updateschedules.CreateMetadata{
			RelatedEnvironmentsIDs: relatedEnvironmentsIDs,
			ScheduledTime:          scheduledTime,
			EnvironmentType:        environmentType,
		})
		if err != nil {
			return httperror.InternalServerError("Unable to persist the edge update schedule", err)
		}

		return nil
	})

	return utils.TxResponse(w, item, err)
}

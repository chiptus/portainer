package edgeupdateschedules

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/internal/slices"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

type updatePayload struct {
	Name          *string
	GroupIDs      []portaineree.EdgeGroupID
	Type          *edgetypes.UpdateScheduleType
	Version       *string
	ScheduledTime *string
	RegistryID    portaineree.RegistryID
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

	if payload.ScheduledTime != nil && *payload.ScheduledTime == "" {
		return errors.New("Scheduled time is required")
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
// @router /edge_update_schedules [post]
func (handler *Handler) update(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	item, err := middlewares.FetchItem[edgetypes.UpdateSchedule](r, contextKey)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	var payload updatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if payload.Name != nil && *payload.Name != item.Name {
		err = handler.validateUniqueName(*payload.Name, item.ID)
		if err != nil {
			return httperror.NewError(http.StatusConflict, "Edge update schedule name already in use", err)
		}

		item.Name = *payload.Name
	}

	stack, err := handler.dataStore.EdgeStack().EdgeStack(item.EdgeStackID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge stack", err)
	}

	shouldUpdate := payload.GroupIDs != nil || payload.Type != nil || payload.Version != nil || payload.ScheduledTime != nil

	if shouldUpdate {
		canUpdate := true
		for _, environmentStatus := range stack.Status {
			if !environmentStatus.Details.Pending {
				canUpdate = false
			}
		}

		if !canUpdate {
			return httperror.BadRequest("Unable to update Edge update schedule", errors.New("edge stack is not in pending state"))
		}

		newGroupIds := payload.GroupIDs
		if newGroupIds == nil {
			newGroupIds = stack.EdgeGroups
		}

		if payload.Type != nil {
			item.Type = *payload.Type
		}

		if payload.Version != nil {
			item.Version = *payload.Version
		}

		scheduledTime := stack.ScheduledTime
		if payload.ScheduledTime != nil {
			scheduledTime = *payload.ScheduledTime
		}

		var registry *portaineree.Registry = nil
		if payload.RegistryID != 0 {
			registry, err = handler.dataStore.Registry().Registry(payload.RegistryID)
			if err == bolterrors.ErrObjectNotFound {
				return httperror.NotFound("Unable to find a registry with the specified identifier inside the database", err)
			} else if err != nil {
				return httperror.InternalServerError("Unable to find a registry with the specified identifier inside the database", err)
			}
			item.RegistryID = payload.RegistryID
		}

		err := handler.dataStore.EdgeStack().DeleteEdgeStack(item.EdgeStackID)
		if err != nil {
			return httperror.InternalServerError("Unable to delete Edge stack", err)
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

		stackID, err := handler.createUpdateEdgeStack(item.ID, newGroupIds, registry, item.Version, scheduledTime, edgeEnvironmentType)
		if err != nil {
			return httperror.InternalServerError("Unable to create Edge stack", err)
		}

		item.EdgeStackID = stackID
	}

	err = handler.updateService.UpdateSchedule(item.ID, item)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the edge update schedule", err)
	}

	return response.JSON(w, item)
}

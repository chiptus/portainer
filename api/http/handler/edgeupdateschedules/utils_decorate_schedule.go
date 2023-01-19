package edgeupdateschedules

import (
	"fmt"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"

	"github.com/portainer/portainer-ee/api/http/middlewares"
)

type decoratedUpdateSchedule struct {
	edgetypes.UpdateSchedule
	EdgeGroupIds  []portaineree.EdgeGroupID          `json:"edgeGroupIds"`
	Status        edgetypes.UpdateScheduleStatusType `json:"status"`
	StatusMessage string                             `json:"statusMessage"`
	ScheduledTime string                             `json:"scheduledTime"`
}

func decorateSchedule(schedule edgetypes.UpdateSchedule, edgeStackGetter middlewares.ItemGetter[portaineree.EdgeStackID, portaineree.EdgeStack], environmentGetter middlewares.ItemGetter[portaineree.EndpointID, portaineree.Endpoint]) (*decoratedUpdateSchedule, error) {
	edgeStack, err := edgeStackGetter(schedule.EdgeStackID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get edge stack")
	}

	status, statusMessage := aggregateStatus(schedule.EnvironmentsPreviousVersions, edgeStack, environmentGetter)

	decoratedItem := &decoratedUpdateSchedule{
		UpdateSchedule: schedule,
		EdgeGroupIds:   edgeStack.EdgeGroups,
		Status:         status,
		StatusMessage:  statusMessage,
		ScheduledTime:  edgeStack.ScheduledTime,
	}

	return decoratedItem, nil
}

func aggregateStatus(relatedEnvironmentsIDs map[portaineree.EndpointID]string, edgeStack *portaineree.EdgeStack, environmentGetter middlewares.ItemGetter[portaineree.EndpointID, portaineree.Endpoint]) (edgetypes.UpdateScheduleStatusType, string) {
	hasSent := false
	hasPending := false

	// if has no related environment
	if len(relatedEnvironmentsIDs) == 0 {
		return edgetypes.UpdateScheduleStatusSuccess, ""
	}

	for environmentID := range relatedEnvironmentsIDs {
		envStatus, ok := edgeStack.Status[environmentID]
		if !ok {
			hasPending = true
			continue
		}

		if envStatus.Details.RemoteUpdateSuccess {
			continue
		}

		// if edge stack reported ok, the update either failed (and we have no way to know) or it's still pending
		if !ok || envStatus.Details.Pending || envStatus.Details.Ok {
			hasPending = true
		}

		if envStatus.Details.Error {
			return edgetypes.UpdateScheduleStatusError, fmt.Sprintf("Error on environment %d: %s", environmentID, envStatus.Error)
		}

		if envStatus.Details.Acknowledged {
			hasSent = true
			break
		}

		// status is "success update"
	}

	if hasPending {
		return edgetypes.UpdateScheduleStatusPending, ""
	}

	if hasSent {
		return edgetypes.UpdateScheduleStatusSent, ""
	}

	return edgetypes.UpdateScheduleStatusSuccess, ""
}

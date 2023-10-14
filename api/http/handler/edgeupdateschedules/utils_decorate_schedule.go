package edgeupdateschedules

import (
	"fmt"
	"slices"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
)

type decoratedUpdateSchedule struct {
	edgetypes.UpdateSchedule
	EdgeGroupIds  []portainer.EdgeGroupID            `json:"edgeGroupIds"`
	Status        edgetypes.UpdateScheduleStatusType `json:"status"`
	StatusMessage string                             `json:"statusMessage"`
	ScheduledTime string                             `json:"scheduledTime"`
}

func decorateSchedule(schedule edgetypes.UpdateSchedule, edgeStackGetter middlewares.ItemGetter[portainer.EdgeStackID, portaineree.EdgeStack], environmentGetter middlewares.ItemGetter[portainer.EndpointID, portaineree.Endpoint]) (*decoratedUpdateSchedule, error) {
	edgeStack, err := edgeStackGetter(schedule.EdgeStackID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get edge stack")
	}

	status, statusMessage := aggregateStatus(schedule.EnvironmentsPreviousVersions, edgeStack, environmentGetter)

	decoratedItem := &decoratedUpdateSchedule{
		UpdateSchedule: schedule,
		EdgeGroupIds:   schedule.EdgeGroupIDs,
		Status:         status,
		StatusMessage:  statusMessage,
		ScheduledTime:  edgeStack.ScheduledTime,
	}

	return decoratedItem, nil
}

func aggregateStatus(relatedEnvironmentsIDs map[portainer.EndpointID]string, edgeStack *portaineree.EdgeStack, environmentGetter middlewares.ItemGetter[portainer.EndpointID, portaineree.Endpoint]) (edgetypes.UpdateScheduleStatusType, string) {
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

		// if a update schedule task is scheduled for future date, it will not have any status
		if len(envStatus.Status) == 0 {
			hasPending = true
			continue
		}

		if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
			return sts.Type == portainer.EdgeStackStatusRemoteUpdateSuccess
		}) {
			continue
		}

		if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
			return sts.Type == portainer.EdgeStackStatusPending ||
				sts.Type == portainer.EdgeStackStatusDeploymentReceived
		}) {
			hasPending = true
		}

		if idx := slices.IndexFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
			return sts.Type == portainer.EdgeStackStatusError
		}); idx >= 0 {
			return edgetypes.UpdateScheduleStatusError, fmt.Sprintf("Error on environment %d: %s", environmentID, envStatus.Status[idx].Error)
		}

		if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
			return sts.Type == portainer.EdgeStackStatusAcknowledged
		}) {
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

package edgeupdateschedules

import (
	"fmt"
	"slices"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
)

type decoratedUpdateSchedule struct {
	edgetypes.UpdateSchedule
	EdgeGroupIds  []portainer.EdgeGroupID            `json:"edgeGroupIds"`
	Status        edgetypes.UpdateScheduleStatusType `json:"status"`
	StatusMessage string                             `json:"statusMessage"`
	ScheduledTime string                             `json:"scheduledTime"`
}

func decorateSchedule(tx dataservices.DataStoreTx, schedule edgetypes.UpdateSchedule) (*decoratedUpdateSchedule, error) {
	edgeStack, err := tx.EdgeStack().EdgeStack(schedule.EdgeStackID)
	if err != nil {
		return nil, fmt.Errorf("unable to get edge stack: %w", err)
	}

	edgeGroup, err := tx.EdgeGroup().Read(edgeStack.EdgeGroups[0])
	if err != nil {
		return nil, fmt.Errorf("unable to get edge group: %w", err)
	}

	status, statusMessage := aggregateStatus(edgeGroup.Endpoints, edgeStack, tx.Endpoint().Endpoint)

	decoratedItem := &decoratedUpdateSchedule{
		UpdateSchedule: schedule,
		EdgeGroupIds:   schedule.EdgeGroupIDs,
		Status:         status,
		StatusMessage:  statusMessage,
		ScheduledTime:  edgeStack.ScheduledTime,
	}

	return decoratedItem, nil
}

func aggregateStatus(relatedEnvironmentsIDs []portainer.EndpointID, edgeStack *portaineree.EdgeStack, environmentGetter middlewares.ItemGetter[portainer.EndpointID, portaineree.Endpoint]) (edgetypes.UpdateScheduleStatusType, string) {
	hasSent := false
	hasPending := false

	// if has no related environment
	if len(relatedEnvironmentsIDs) == 0 {
		return edgetypes.UpdateScheduleStatusSuccess, ""
	}

	for _, environmentID := range relatedEnvironmentsIDs {
		envStatus, ok := edgeStack.Status[environmentID]
		if !ok || len(envStatus.Status) == 0 {
			hasPending = true
			continue
		}

		if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
			return sts.Type == portainer.EdgeStackStatusRemoteUpdateSuccess
		}) {
			continue
		}
		if idx := slices.IndexFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
			return sts.Type == portainer.EdgeStackStatusError
		}); idx != -1 {
			return edgetypes.UpdateScheduleStatusError, fmt.Sprintf("Error on environment %d: %s", environmentID, envStatus.Status[idx].Error)
		}

		lastStatus := envStatus.Status[len(envStatus.Status)-1]

		if lastStatus.Type == portainer.EdgeStackStatusPending {
			hasPending = true
		}

		if lastStatus.Type == portainer.EdgeStackStatusAcknowledged ||
			lastStatus.Type == portainer.EdgeStackStatusDeploymentReceived {
			hasSent = true
			break
		}

		// status is "success update"
	}

	if hasSent {
		return edgetypes.UpdateScheduleStatusSent, ""
	}

	if hasPending {
		return edgetypes.UpdateScheduleStatusPending, ""
	}

	return edgetypes.UpdateScheduleStatusSuccess, ""
}

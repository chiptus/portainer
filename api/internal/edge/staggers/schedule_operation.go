package staggers

import (
	"fmt"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

type StaggerScheduleOperation struct {
	// currentIndex is used to track the current index of the stagger queue
	currentIndex int
	// length is used to track the length of the stagger queue
	length int
	// staggerQueue is used to maintain a list of stagger queue for each endpoints
	staggerQueue [][]portaineree.EndpointID
	// endpointStatus is used to maintain a list of endpoint status for each endpoints
	endpointStatus map[portaineree.EndpointID]portainer.EdgeStackStatusType
	// paused is used to track if the stagger workflow is paused
	paused bool
	// rollback is used to track if the stagger workflow should be rolled back
	rollback bool

	// todo: support the below two fields
	timeout             time.Duration
	updateDelay         time.Duration
	updateFailureAction portaineree.EdgeUpdateFailureAction
}

func (sso *StaggerScheduleOperation) IsPaused() bool {
	return sso.paused
}

func (sso *StaggerScheduleOperation) SetPaused(paused bool) {
	log.Debug().Msg("=====> SetPaused")
	sso.paused = paused
}

func (sso *StaggerScheduleOperation) IsCompleted() bool {
	return (!sso.rollback && sso.currentIndex >= sso.length) || sso.isRolledback()
}

func (sso *StaggerScheduleOperation) isRolledback() bool {
	return sso.rollback && sso.currentIndex < 0
}

func (sso *StaggerScheduleOperation) ShouldRollback() bool {
	return sso.rollback
}

func (sso *StaggerScheduleOperation) SetRollback(rollback bool) {
	log.Debug().Msg("=====> SetRollback")
	if rollback {
		// we need to remove ETag cache here for all endpoints in the stagger queue
		for endpoint := range sso.endpointStatus {
			cache.Del(endpoint)
		}
	}
	sso.rollback = rollback
}

func (sso *StaggerScheduleOperation) MoveToNextQueue() {
	if sso.rollback {
		sso.currentIndex--
		return
	}
	sso.currentIndex++
}

func (sso *StaggerScheduleOperation) Info() string {
	return fmt.Sprintf("index: %d length: %d stagger queue: %v endpoint status: %v", sso.currentIndex, sso.length, sso.staggerQueue, sso.endpointStatus)
}

// UpdateStaggerQueue is used to check if the stagger queue should be moved to the next queue or set to other
// state based on the incoming endpoint status and the update failure action
func (sso *StaggerScheduleOperation) UpdateStaggerQueue(endpointID portaineree.EndpointID, status portainer.EdgeStackStatusType) {
	staggeredEndpoints := sso.staggerQueue[sso.currentIndex]

	allowToMoveToNextStaggerQueue := true

	isStaggeredEndpoint := false // it represents if the endpoint is in the current stagger queue
	for _, staggeredEndpoint := range staggeredEndpoints {
		// if we found a matched endpoint, we need to mark it
		// and update the endpoint status later as we need to
		// interate all current staggered endpoints to collect
		// the status
		if staggeredEndpoint == endpointID {
			isStaggeredEndpoint = true
			sso.endpointStatus[staggeredEndpoint] = status
		}

		endpointStatus := sso.endpointStatus[staggeredEndpoint]
		switch endpointStatus {
		case portainer.EdgeStackStatusRunning:

		case portainer.EdgeStackStatusError:
			allowToMoveToNextStaggerQueue = false

			if sso.updateFailureAction == portaineree.EdgeUpdateFailureActionContinue {
				// if the update failure action is continue, we can ignore the error
				allowToMoveToNextStaggerQueue = true
			} else if sso.updateFailureAction == portaineree.EdgeUpdateFailureActionPause {
				log.Debug().Msg("Stagger workflow is paused after updating stagger queue, skip")

				// if the update failure action is pause and we found an error, it
				// means we need to pause the entire stagger workflow
				sso.SetPaused(true)

				return

			} else if sso.updateFailureAction == portaineree.EdgeUpdateFailureActionRollback {
				log.Debug().Msg("An endpoint is failed to update, stagger workflow starts to rollback")

				// if the update failure action is rollback, we need to rollback the
				// entire stagger workflow
				sso.SetRollback(true)

				// with rolling back the current stagger queue, we need to overwrite the
				// current endpoint status to pending from Error
				sso.endpointStatus[staggeredEndpoint] = portainer.EdgeStackStatusPending

				return

			} else {
				log.Error().Msgf("Unsupported update failure action: %d", sso.updateFailureAction)
			}

		case portainer.EdgeStackStatusPending:
			// if one endpoint status is pending, it means that the current stagger queue is not completed
			allowToMoveToNextStaggerQueue = false

		}
	}

	if !isStaggeredEndpoint {
		// if the endpoint is not in the staggered queue, it means it is not in the staggered workflow
		// so we don't need to update its status in this round of updating
		return
	}

	if allowToMoveToNextStaggerQueue {
		// if all the endpoints in the current stagger queue are okay,
		// we can move to the next staggered queue
		sso.MoveToNextQueue()
		return
	}
}

// RollbackStaggerQueue is used to check if the stagger queue should be rolled back to the previous queue or
// set to other state based on the incoming endpoint status and the update failure action
func (sso *StaggerScheduleOperation) RollbackStaggerQueue(endpointID portaineree.EndpointID, status portainer.EdgeStackStatusType, stackFileVersion int, rollbackTo *int) {
	staggeredEndpoints := sso.staggerQueue[sso.currentIndex]

	allowToRollbackStaggerQueue := true
	isStaggeredEndpoint := false // it represents if the endpoint is in the current stagger queue
	for _, staggeredEndpoint := range staggeredEndpoints {
		if staggeredEndpoint == endpointID {
			isStaggeredEndpoint = true

			switch status {
			case portainer.EdgeStackStatusRunning:
				if rollbackTo == nil {
					// If the incoming status is Ok but the rollbackTo is nil, it means that the status
					// was generated before the stagger queue was set to Rollback. In such case, we need
					// to update the endpoint status to Running so that it can be processed in the next
					// round of stagger queue rollback
					sso.endpointStatus[staggeredEndpoint] = portainer.EdgeStackStatusRunning
				}
				// If the incoming status is Ok, it means that the endpoint is rolled back. We need to
				// update the endpoint status to Pending
				sso.endpointStatus[staggeredEndpoint] = portainer.EdgeStackStatusPending

				// if the endpoint is rolled back successfully, we should update the endpoint's edge
				// status's DeploymentInfo to the previous version. This db operation will be done in
				// API endpoint /edge_stacks/{id}/status

			case portainer.EdgeStackStatusError:
				// If the incoming status is Error, it means that the endpoint is failed to rollback.
				// In such case, we will ignore the error and move to the next stagger queue
				// This workflow has confirmed with Product team

				sso.endpointStatus[staggeredEndpoint] = portainer.EdgeStackStatusPending

			case portainer.EdgeStackStatusPending:
				// If the incoming status is Pending, it means that the endpoint is not rolled back.
				allowToRollbackStaggerQueue = false
			}

			continue
		}

		// the below block is used to check if the other endpoints in the current stagger
		// queue are rolled back or not
		endpointStatus := sso.endpointStatus[staggeredEndpoint]
		switch endpointStatus {
		case portainer.EdgeStackStatusRunning:
			// !! important
			// If the endpoint status is Ok, it means that the current queue is not rolled back.
			// The logic is opposite to update stagger queue
			allowToRollbackStaggerQueue = false

		case portainer.EdgeStackStatusError:
			// same as above. Ignore the error and move to the next stagger queue
			allowToRollbackStaggerQueue = true

		case portainer.EdgeStackStatusPending:
			// If the endpoint status is Pending, it means that the endpoint is already rolled back
			// or unnecessary to be rolled back because it was not updated, regardless of which
			// reason, it will not affect the flow of rolling back the stagger queue
		}
	}

	if !isStaggeredEndpoint {
		return
	}

	if allowToRollbackStaggerQueue {
		sso.MoveToNextQueue()
	}

}

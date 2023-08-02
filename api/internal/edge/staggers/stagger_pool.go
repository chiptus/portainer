package staggers

import (
	"fmt"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/unique"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

type StaggerPoolKey string

func GetStaggerPoolKey(edgeStackID portaineree.EdgeStackID, stackFileVersion int) StaggerPoolKey {
	return StaggerPoolKey(fmt.Sprintf("%d-%d", edgeStackID, stackFileVersion))
}

func (service *Service) startStaggerPool() {
	log.Debug().Msg("Starting stagger pool")

	ticker := time.NewTicker(7 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-service.shutdownCtx.Done():
			// todo: if Stagger pool is not empty and there are unfinished stagger queues
			// we need to save the current stagger queue state to the database
			log.Debug().Msg("Stopping stagger pool")
			close(service.staggerJobQueue)
			close(service.staggerStatusJobQueue)

			service.staggerPoolMtx.Lock()
			for _, scheduleOperation := range service.staggerPool {
				for _, timer := range scheduleOperation.timeoutTimerMap {
					timer.Stop()
				}
			}
			service.staggerPoolMtx.Unlock()
			return

		case newJob := <-service.staggerJobQueue:
			// Build the stagger schedule operation based on the stagger config
			if newJob.Config.StaggerOption == portaineree.EdgeStaggerOptionAllAtOnce {
				log.Debug().Msg("Stagger option is all at once, skip")
				break
			}

			log.Debug().
				Int("edgeStackID", int(newJob.EdgeStackID)).
				Msg("Received stagger job")

			timeoutDuration, err := time.ParseDuration(newJob.Config.Timeout)
			if err != nil {
				log.Error().Err(err).
					Msgf("Failed to parse timeout duration: %s", newJob.Config.Timeout)
				break
			}

			updateDelayDuration, err := time.ParseDuration(newJob.Config.UpdateDelay)
			if err != nil {
				log.Error().Err(err).
					Msgf("Failed to parse update delay duration: %s", newJob.Config.UpdateDelay)
				break
			}

			scheduleOperation := StaggerScheduleOperation{
				edgeStackID:         newJob.EdgeStackID,
				currentIndex:        0,
				length:              0,
				endpointStatus:      make(map[portaineree.EndpointID]portainer.EdgeStackStatusType, 0),
				timeoutTimerMap:     make(map[portaineree.EndpointID]*time.Timer, 0),
				timeout:             timeoutDuration,
				updateDelay:         updateDelayDuration,
				updateDelayMap:      make(map[int]time.Time, 0),
				updateFailureAction: newJob.Config.UpdateFailureAction,
			}

			// 1. collect all related endpoints
			endpointIDs := []portaineree.EndpointID{}
			err = service.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
				edgeStack, err := tx.EdgeStack().EdgeStack(newJob.EdgeStackID)
				if err != nil {
					return err
				}

				for _, edgeGroupID := range edgeStack.EdgeGroups {
					edgeGroup, err := tx.EdgeGroup().Read(edgeGroupID)
					if err != nil {
						return err
					}

					endpointIDs = append(endpointIDs, edgeGroup.Endpoints...)
				}
				return nil
			})
			if err != nil {
				log.Error().Err(err).
					Msgf("Failed to collect all related endpoints of edge stack: %d", newJob.EdgeStackID)
				break
			}

			endpointIDs = unique.Unique(endpointIDs)

			// 2. build stagger queue based on the stagger config
			config := newJob.Config
			if config.StaggerParallelOption == portaineree.EdgeStaggerParallelOptionFixed {
				scheduleOperation.staggerQueue = buildStaggerQueueWithFixedDeviceNumber(endpointIDs, config.DeviceNumber)

			} else if config.StaggerParallelOption == portaineree.EdgeStaggerParallelOptionIncremental {
				scheduleOperation.staggerQueue = buildStaggerQueueWithIncrementalDeviceNumber(endpointIDs, config.DeviceNumberStartFrom, config.DeviceNumberIncrementBy)

			} else {
				log.Error().Msgf("Unsupported stagger parallel option: %d", config.StaggerParallelOption)
				break
			}

			// 3. initialize the endpoint status
			scheduleOperation.length = len(scheduleOperation.staggerQueue)
			for _, endpointId := range endpointIDs {
				// Set default status to pending
				scheduleOperation.endpointStatus[endpointId] = portainer.EdgeStackStatusPending
			}

			poolKey := GetStaggerPoolKey(newJob.EdgeStackID, newJob.StackFileVersion)
			service.staggerPoolMtx.Lock()
			service.staggerPool[poolKey] = scheduleOperation
			service.staggerPoolMtx.Unlock()

			service.DisplayStaggerInfo()

		case newStatusJob := <-service.staggerStatusJobQueue:
			// Process the endpoints' edge stack status updates
			log.Debug().
				Int("edgeStackID", int(newStatusJob.EdgeStackID)).
				Int("stackFileVersion", newStatusJob.StackFileVersion).
				Int("endpointID", int(newStatusJob.EndpointID)).
				Msgf("Received stagger status job: %d", newStatusJob.Status)

			service.ProcessStatusJob(newStatusJob)

		case <-ticker.C:
			service.DisplayStaggerInfo()
		}
	}
}

func (service *Service) ProcessStatusJob(newStatusJob *StaggerStatusJob) {
	poolKey := GetStaggerPoolKey(newStatusJob.EdgeStackID, newStatusJob.StackFileVersion)

	service.staggerPoolMtx.Lock()
	defer service.staggerPoolMtx.Unlock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		log.Debug().Str("pool key", string(poolKey)).
			Msg("Failed to retrieve stagger schedule operation for edge stack ")
		return
	}

	if scheduleOperation.IsPaused() {
		log.Debug().Str("pool key", string(poolKey)).
			Msg("Stagger workflow is paused, skip")

		return
	}

	if scheduleOperation.IsCompleted() {
		log.Debug().Str("pool key", string(poolKey)).
			Msg("Stagger workflow is completed, skip")

		return
	}

	if scheduleOperation.ShouldRollback() {
		// Operation to rollback the edge stack of endpoints in the stagger queue one by one
		// This operation corresponds to the failure action of "rollback"
		log.Debug().Msg("Stagger workflow is rolling back")
		scheduleOperation.RollbackStaggerQueue(newStatusJob.EndpointID, newStatusJob.Status, newStatusJob.StackFileVersion, newStatusJob.RollbackTo)

	} else {
		// Operation to update the edge stack of endpoints in the stagger queue one by one
		// This operation corresponds to the failure action of "continue" and "pause"
		scheduleOperation.UpdateStaggerQueue(newStatusJob.EndpointID, newStatusJob.Status)

	}

	service.staggerPool[poolKey] = scheduleOperation
}

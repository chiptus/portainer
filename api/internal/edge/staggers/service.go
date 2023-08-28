package staggers

import (
	"context"
	"errors"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

type StaggerJob struct {
	EdgeStackID        portaineree.EdgeStackID
	StackFileVersion   int
	RelatedEndpointIDs []portaineree.EndpointID
	Config             *portaineree.EdgeStaggerConfig
}

type StaggerStatusJob struct {
	EdgeStackID      portaineree.EdgeStackID
	EndpointID       portaineree.EndpointID
	StackFileVersion int
	RollbackTo       *int
	Status           portainer.EdgeStackStatusType
}

// Service represents a service for managing edge stacks' stagger configs and workflow
type Service struct {
	shutdownCtx      context.Context
	dataStore        dataservices.DataStore
	edgeAsyncService *edgeasync.Service
	// staggerConfigsMtx is used to protect staggerConfigs
	staggerConfigsMtx sync.RWMutex
	// staggerConfigs is used to maintain a list of stagger configs for each stack
	// At a time, there is only one stagger config for each stack, the new stagger config for the same stack
	// cannot be added until the previous stagger config is completed
	// todo: Should be saved in the database in the future
	staggerConfigs map[portaineree.EdgeStackID]portaineree.EdgeStaggerConfig
	// staggerPoolMtx is used to protect staggerPool
	staggerPoolMtx sync.RWMutex
	// staggerPool is used to maintain a list of processed stagger schedule operation
	// based on stagger config for each stack on its stack file version.
	// StaggerPoolKey consists of edge stack id and stack file version, which allows to
	// handle multiple stagger workflows for the same edge stack
	staggerPool map[StaggerPoolKey]StaggerScheduleOperation
	// staggerJobQueue is used to maintain a list of stagger jobs to be processed.
	// stagger job means the operation of updating an edge stack in all related endpoints with stagger configuration.
	staggerJobQueue chan *StaggerJob
	// staggerStatusJobQueue is used to maintain a list of stagger status jobs to be processed
	// stagger status job means the operation of updating the endpoints' edge stack status
	staggerStatusJobQueue chan *StaggerStatusJob
	// asyncPoolTerminators is used to maintain a list of async pool terminators for each stack on
	// its stack file version
	asyncPoolTerminators map[StaggerPoolKey]context.CancelFunc
}

func NewService(ctx context.Context, dataStore dataservices.DataStore, edgeAsyncService *edgeasync.Service) *Service {

	s := &Service{
		shutdownCtx:       ctx,
		dataStore:         dataStore,
		edgeAsyncService:  edgeAsyncService,
		staggerConfigsMtx: sync.RWMutex{},
		staggerConfigs:    make(map[portaineree.EdgeStackID]portaineree.EdgeStaggerConfig, 0),
		staggerPoolMtx:    sync.RWMutex{},
		staggerPool:       make(map[StaggerPoolKey]StaggerScheduleOperation, 0),
		// todo: make the channel size configurable from UI
		// The buffer size specifies the maximum total number of edge stacks being set to update based on
		// stagger configuration, during a given time period.
		staggerJobQueue: make(chan *StaggerJob, 20),
		// The buffer size specifies the maximum total number of endpoints that can be queued in
		// all current stagger queues across all edge stacks being updated, during a given time period.
		staggerStatusJobQueue: make(chan *StaggerStatusJob, 100),
		asyncPoolTerminators:  make(map[StaggerPoolKey]context.CancelFunc, 0),
	}

	go s.startStaggerPool()
	return s
}

// AddStaggerConfig is used to add a new stagger config for specific edge stack. If the edge stack is still running
// under the existing stagger config, it will return an error
func (service *Service) AddStaggerConfig(id portaineree.EdgeStackID, stackFileVersion int, config *portaineree.EdgeStaggerConfig, endpointIDs []portaineree.EndpointID) error {
	if config.StaggerOption == portaineree.EdgeStaggerOptionAllAtOnce {
		return nil
	}

	service.staggerConfigsMtx.Lock()
	_, ok := service.staggerConfigs[id]
	if ok {
		service.staggerConfigsMtx.Unlock()
		return errors.New("the stack is still running under the existing stagger configuration")
	}
	service.staggerConfigs[id] = *config
	service.staggerConfigsMtx.Unlock()

	newJob := &StaggerJob{
		EdgeStackID:        id,
		StackFileVersion:   stackFileVersion,
		Config:             config,
		RelatedEndpointIDs: endpointIDs,
	}

	service.staggerJobQueue <- newJob
	return nil
}

func (service *Service) RemoveStaggerConfig(id portaineree.EdgeStackID) {
	// remove the config from stagger configuration

	service.staggerConfigsMtx.Lock()
	defer service.staggerConfigsMtx.Unlock()

	_, ok := service.staggerConfigs[id]
	if !ok {
		return
	}

	log.Debug().Int("edgeStackID", int(id)).
		Msg("[Stagger service] Remove stagger config")

	delete(service.staggerConfigs, id)
}

func (service *Service) IsEdgeStackUpdating(id portaineree.EdgeStackID) bool {
	service.staggerConfigsMtx.RLock()
	defer service.staggerConfigsMtx.RUnlock()

	_, ok := service.staggerConfigs[id]
	return ok
}

// IsStaggered is used to check if the edge stack is staggered for specific endpoint
func (service *Service) IsStaggeredEdgeStack(id portaineree.EdgeStackID, fileVersion int, endpointID portaineree.EndpointID) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	pool, ok := service.staggerPool[poolKey]
	if !ok {
		return false
	}

	if endpointID == 0 {
		return true
	}
	// If an endpoint is added into the edge group after the stagger workflow is started,
	// it will not be included in the stagger workflow
	_, ok = pool.endpointStatus[endpointID]

	return ok
}

// StopAndRemoveStaggerScheduleOperation is used to stop and remove the stagger schedule operation
// for specific edge stack. It is called when the edge stack is deleted.
func (service *Service) StopAndRemoveStaggerScheduleOperation(id portaineree.EdgeStackID) {
	service.staggerPoolMtx.Lock()
	for key, scheduleOperation := range service.staggerPool {
		if scheduleOperation.edgeStackID != id {
			continue
		}

		// stop the timeout timer
		for _, timer := range scheduleOperation.timeoutTimerMap {
			timer.Stop()
		}

		// stop the async pool
		cancelFunc, ok := service.asyncPoolTerminators[key]
		if ok {
			cancelFunc()
		}

		// remove all the stagger related to the edge stack
		delete(service.staggerPool, key)

		log.Debug().Str("poolKey", string(key)).
			Msg("[Stagger service] schedule operation is removed")
	}
	service.staggerPoolMtx.Unlock()

	service.RemoveStaggerConfig(id)
	log.Debug().Int("edgeStackID", int(id)).
		Msg("[Stagger service] configuration is removed")
}

// CanProceedAsStaggerJob is used to check if the edge stack can proceed as stagger job for specific endpoint
func (service *Service) CanProceedAsStaggerJob(id portaineree.EdgeStackID, fileVersion int, endpointID portaineree.EndpointID) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	scheduleOperation, ok := service.staggerPool[poolKey]
	service.staggerPoolMtx.RUnlock()
	if !ok {
		return false
	}

	// Check if the stagger workflow is paused
	if scheduleOperation.IsPaused() || scheduleOperation.IsCompleted() {
		log.Debug().Msg("[Stagger service] update is paused or completed, skip")

		return false
	}

	if !scheduleOperation.ShouldRollback() {
		if scheduleOperation.endpointStatus[endpointID] == portainer.EdgeStackStatusRunning {
			log.Debug().Msg("[Stagger service] Endpoint is already updated, skip")

			return true
		}
	} else {
		if scheduleOperation.endpointStatus[endpointID] == portainer.EdgeStackStatusPending {
			log.Debug().Msg("[Stagger service] Endpoint is already rolled back, skip")

			return true
		}
	}

	staggeringEndpoints := scheduleOperation.staggerQueue[scheduleOperation.currentIndex]
	for _, staggeringEndpoint := range staggeringEndpoints {
		if staggeringEndpoint != endpointID {
			continue
		}

		log.Debug().Int("endpointID", int(endpointID)).
			Msg("[Stagger service] Found endpoint in the stagger queue")

		// check if the update delay is set
		if scheduleOperation.updateDelay > 0 && !scheduleOperation.ShouldRollback() {
			// check if the update delay is reached

			delayTime, ok := scheduleOperation.updateDelayMap[scheduleOperation.currentIndex]
			if ok {
				// updateDelayMap starts to record delay time from index 1, so for the first
				// stagger queue, there is no delay time, or the delay time is already reached,
				// which has been removed from the map
				if time.Now().Before(delayTime) {
					log.Debug().Msg("Update delay is not reached, skip")
					return false
				}

				go service.removeUpdateDelay(poolKey)
			}
		}

		// check if the timeout is set
		if scheduleOperation.timeout > 0 && !scheduleOperation.ShouldRollback() {
			go service.setTimeout(id, fileVersion, endpointID)
		}

		return true
	}

	return false
}

// MarkAsStaggered is used to indicate the stagger workflow is set to rollback for specific edge stack
func (service *Service) MarkedAsRollback(id portaineree.EdgeStackID, fileVersion int) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		return false
	}

	return scheduleOperation.ShouldRollback()
}

// WasEndpointRolledBack is used to check if the endpoint was rolled back for specific edge stack
func (service *Service) WasEndpointRolledBack(id portaineree.EdgeStackID, fileVersion int, endpointId portaineree.EndpointID) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		return false
	}

	if !scheduleOperation.ShouldRollback() {
		return false
	}

	endpointStatus, ok := scheduleOperation.endpointStatus[endpointId]
	if !ok {
		return false
	}

	// if the endpoint status is Pending and the stagger queue rollback is enabled,
	// it means the endpoint is rolled back or not updated yet
	return endpointStatus == portainer.EdgeStackStatusPending
}

func (service *Service) MarkedAsCompleted(id portaineree.EdgeStackID, fileVersion int) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	scheduleOperation, ok := service.staggerPool[poolKey]
	service.staggerPoolMtx.RUnlock()
	if !ok {
		return false
	}

	// Must include IsPaused status, so that the updated edge stack on agent will not be removed
	if scheduleOperation.IsPaused() || scheduleOperation.IsCompleted() {
		log.Debug().Msg("[Stagger service] update is paused or completed, skip")

		go func() {
			// If the stagger workflow is paused or completed, we still need to maintain the stagger queue
			// for each stack. However, it is not necessary to keep its async pool, regardless of whether
			// endpoints are async edge agents or not.
			// The reason is that the way that the server notify async agents to update is by creating
			// stack command based on the stagger configuration, instead of being polled by the agent.
			// So terminating the async pool will not affect the stagger workflow.
			// This operation can release the resources of the server.
			service.terminateAsyncPool(poolKey)

			// unblock edge stack update with stagger configuration
			service.RemoveStaggerConfig(id)
		}()

		return true
	}

	return false
}

// UpdateStaggerEndpointStatusIfNeeds is used to update the stagger status for specific endpoint
// It is called when the agent updates the edge stack status. "fileVersion" is not the edge
// stack file version that each agent is using, it is the edge stack file version that is
// used to differentiate the stagger workflow for the same edge stack
func (service *Service) UpdateStaggerEndpointStatusIfNeeds(id portaineree.EdgeStackID, fileVersion int, rollbackTo *int, endpointID portaineree.EndpointID, status portainer.EdgeStackStatusType) {
	if status != portainer.EdgeStackStatusRunning &&
		status != portainer.EdgeStackStatusError {
		return
	}

	statusJob := &StaggerStatusJob{
		EdgeStackID:      id,
		EndpointID:       endpointID,
		StackFileVersion: fileVersion,
		RollbackTo:       rollbackTo,
		Status:           status,
	}

	service.staggerStatusJobQueue <- statusJob
}

func (service *Service) setAsyncPoolTerminator(edgeStackID portaineree.EdgeStackID, stackFileVersion int, cancelFunc context.CancelFunc) {
	poolKey := GetStaggerPoolKey(edgeStackID, stackFileVersion)

	service.staggerPoolMtx.Lock()
	defer service.staggerPoolMtx.Unlock()

	service.asyncPoolTerminators[poolKey] = cancelFunc
}

func (service *Service) terminateAsyncPool(poolKey StaggerPoolKey) {
	service.staggerPoolMtx.Lock()
	defer service.staggerPoolMtx.Unlock()

	cancelFunc, ok := service.asyncPoolTerminators[poolKey]
	if !ok {
		return
	}

	log.Debug().
		Str("poolKey", string(poolKey)).
		Msg("[Stagger Async] Stagger job completed")

	cancelFunc()
	delete(service.asyncPoolTerminators, poolKey)
}

func (service *Service) removeUpdateDelay(poolKey StaggerPoolKey) {
	log.Debug().Msg("[Stagger service] Update delay is expired, removed")

	// remove the update delay after it is reached, preventing it from being used
	// again in rollback workflow
	service.staggerPoolMtx.Lock()
	defer service.staggerPoolMtx.Unlock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		return
	}

	delete(scheduleOperation.updateDelayMap, scheduleOperation.currentIndex)
	service.staggerPool[poolKey] = scheduleOperation
}

func (service *Service) setTimeout(id portaineree.EdgeStackID, fileVersion int, endpointID portaineree.EndpointID) {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.Lock()
	defer service.staggerPoolMtx.Unlock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		return
	}

	endpoint := new(portaineree.Endpoint)
	err := service.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		var err error
		endpoint, err = tx.Endpoint().Endpoint(endpointID)
		return err
	})
	if err != nil {
		log.Error().Err(err).
			Int("endpointID", int(endpointID)).
			Msg("[Stagger service] Failed to get endpoint check interval for stagger job")
	}

	timeout := scheduleOperation.timeout
	if endpoint != nil {
		// If the stagger configuration timeout is shorter than the endpoint checkin
		// interval or snapshot interval, it's meaningless to set the timeout.
		// To avoid such case, we add the interval to the timeout
		if endpoint.Edge.AsyncMode {
			timeout += time.Duration(endpoint.Edge.SnapshotInterval) * time.Second
		} else {
			timeout += time.Duration(endpoint.EdgeCheckinInterval) * time.Second
		}
	}

	// start timeout timer
	timer := time.AfterFunc(timeout, func() {
		log.Warn().Int("endpointID", int(endpointID)).
			Int("edgeStackID", int(id)).
			Int("fileVersion", fileVersion).
			Msg("[Stagger service] job timeout, update endpoint status to Error")

		// If timeout is reached, explicitly update endpoint status to Error
		service.UpdateStaggerEndpointStatusIfNeeds(id, fileVersion, nil, endpointID, portainer.EdgeStackStatusError)
	})

	log.Debug().Int("endpointID", int(endpointID)).
		Str("poolKey", string(poolKey)).
		Int("timeout", int(timeout)).
		Msg("[Stagger service] Set timeout for stagger job")

	scheduleOperation.timeoutTimerMap[endpointID] = timer

	service.staggerPool[poolKey] = scheduleOperation
}

// DisplayStaggerInfo is used to display the stagger info for debugging purpose
func (service *Service) DisplayStaggerInfo() {
	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	for key, scheduleOperation := range service.staggerPool {
		if scheduleOperation.IsCompleted() || scheduleOperation.IsPaused() {
			continue
		}
		log.Debug().
			Bool("Rollback", scheduleOperation.ShouldRollback()).
			Str("edgeStackID-fileVersion", string(key)).
			Str("schedule operation", scheduleOperation.Info()).
			Msg("[Stagger service] pool info")
	}
}

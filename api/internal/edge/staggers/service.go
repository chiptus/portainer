package staggers

import (
	"context"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

type StaggerJob struct {
	EdgeStackID      portaineree.EdgeStackID
	StackFileVersion int
	Config           *portaineree.EdgeStaggerConfig
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

	// staggerConfigs is used to maintain a list of stagger configs for each stack
	staggerConfigs map[portaineree.EdgeStackID][]portaineree.EdgeStaggerConfig
	// staggerPoolMtx is used to protect staggerPool
	staggerPoolMtx sync.RWMutex
	// staggerPool is used to maintain a list of processed stagger schedule operation
	// based on stagger config for each stack.
	// StaggerPoolKey consists of edge stack id and stack file version, which allows to
	// handle multiple stagger workflows for the same edge stack
	staggerPool map[StaggerPoolKey]StaggerScheduleOperation
	// staggerJobQueue is used to maintain a list of stagger jobs to be processed
	staggerJobQueue chan *StaggerJob
	// staggerStatusJobQueue is used to maintain a list of stagger status jobs to be processed
	staggerStatusJobQueue chan *StaggerStatusJob
}

func NewService(ctx context.Context, dataStore dataservices.DataStore, edgeAsyncService *edgeasync.Service) *Service {

	s := &Service{
		shutdownCtx:      ctx,
		dataStore:        dataStore,
		edgeAsyncService: edgeAsyncService,
		staggerConfigs:   make(map[portaineree.EdgeStackID][]portaineree.EdgeStaggerConfig, 0),
		staggerPoolMtx:   sync.RWMutex{},
		staggerPool:      make(map[StaggerPoolKey]StaggerScheduleOperation, 0),
		// todo: make the channel size configurable
		staggerJobQueue:       make(chan *StaggerJob, 10),
		staggerStatusJobQueue: make(chan *StaggerStatusJob, 10),
	}

	go s.startStaggerPool()
	return s
}

func (service *Service) AddStaggerConfig(id portaineree.EdgeStackID, stackFileVersion int, config *portaineree.EdgeStaggerConfig) {
	if config.StaggerOption == portaineree.EdgeStaggerOptionAllAtOnce {
		return
	}

	newJob := &StaggerJob{
		EdgeStackID:      id,
		StackFileVersion: stackFileVersion,
		Config:           config,
	}

	service.staggerJobQueue <- newJob

	if _, ok := service.staggerConfigs[id]; !ok {
		service.staggerConfigs[id] = []portaineree.EdgeStaggerConfig{*config}
	}

	service.staggerConfigs[id] = append(service.staggerConfigs[id], *config)
}

// IsStaggered is used to check if the edge stack is staggered for specific endpoint
func (service *Service) IsStaggeredEdgeStack(id portaineree.EdgeStackID, fileVersion int) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	_, ok := service.staggerPool[poolKey]
	if !ok {
		// log.Debug().Msgf("Edge stack %d with file version %d is not staggered", id, fileVersion)
		return false
	}
	// log.Debug().Msgf("Edge stack %d with file version %d is staggered", id, fileVersion)
	return true
}

// CanProceedAsStaggerJob is used to check if the edge stack can proceed as stagger job for specific endpoint
func (service *Service) CanProceedAsStaggerJob(id portaineree.EdgeStackID, fileVersion int, endpointID portaineree.EndpointID) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		service.staggerPoolMtx.RUnlock()
		return false
	}
	service.staggerPoolMtx.RUnlock()

	// Check if the stagger workflow is paused
	if scheduleOperation.IsPaused() && scheduleOperation.IsCompleted() {
		log.Debug().Msg("Stagger workflow is paused or completed, skip")
		return false
	}

	staggeringEndpoints := scheduleOperation.staggerQueue[scheduleOperation.currentIndex]
	for _, staggeringEndpoint := range staggeringEndpoints {
		if staggeringEndpoint == endpointID {
			log.Debug().Int("endpointID", int(endpointID)).Msg("Found endpoint in the stagger queue")
			return true
		}
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

	if scheduleOperation.ShouldRollback() {
		endpointStatus, ok := scheduleOperation.endpointStatus[endpointId]
		if ok {
			if endpointStatus == portainer.EdgeStackStatusPending {
				// if the endpoint status is Pending and the stagger queue rollback is enabled,
				// it means the endpoint is rolled back
				return true
			}
		}
	}
	return false
}

func (service *Service) MarkedAsCompleted(id portaineree.EdgeStackID, fileVersion int) bool {
	poolKey := GetStaggerPoolKey(id, fileVersion)

	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	scheduleOperation, ok := service.staggerPool[poolKey]
	if !ok {
		return false
	}

	return scheduleOperation.IsCompleted()
}

// UpdateStaggerStatusIfNeeds is used to update the stagger status for specific endpoint
// It is called when the agent updates the edge stack status. "fileVersion" is not the edge
// stack file version that each agent is using, it is the edge stack file version that is
// used to differentiate the stagger workflow for the same edge stack
func (service *Service) UpdateStaggerStatusIfNeeds(id portaineree.EdgeStackID, fileVersion int, rollbackTo *int, endpointID portaineree.EndpointID, status portainer.EdgeStackStatusType) {
	if status != portainer.EdgeStackStatusRunning &&
		status != portainer.EdgeStackStatusError &&
		status != portainer.EdgeStackStatusPending {
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

// DisplayStaggerInfo is used to display the stagger info for debugging purpose
func (service *Service) DisplayStaggerInfo() {
	service.staggerPoolMtx.RLock()
	defer service.staggerPoolMtx.RUnlock()

	for key, scheduleOperation := range service.staggerPool {
		log.Debug().
			Str("edgeStackID-fileVersion", string(key)).
			Str("schedule operation", scheduleOperation.Info()).
			Msg("Stagger pool")
	}
}

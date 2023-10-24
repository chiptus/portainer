package testhelpers

import (
	"context"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/staggers"
	"github.com/portainer/portainer-ee/api/internal/set"
	portainer "github.com/portainer/portainer/api"
)

type testStaggerService struct {
}

func NewTestStaggerService() *testStaggerService {
	return &testStaggerService{}
}

func (testStaggerService) AddStaggerConfig(id portainer.EdgeStackID, stackFileVersion int, config *portaineree.EdgeStaggerConfig, endpointIDs []portainer.EndpointID) error {
	return nil
}

func (testStaggerService) RemoveStaggerConfig(id portainer.EdgeStackID) {

}

func (testStaggerService) IsEdgeStackUpdating(id portainer.EdgeStackID) bool {
	return false
}

func (testStaggerService) IsStaggeredEdgeStack(id portainer.EdgeStackID, fileVersion int, endpointID portainer.EndpointID) bool {
	return false
}

func (testStaggerService) StopAndRemoveStaggerScheduleOperation(id portainer.EdgeStackID) {

}

func (testStaggerService) CanProceedAsStaggerJob(id portainer.EdgeStackID, fileVersion int, endpointID portainer.EndpointID) bool {
	return false
}

func (testStaggerService) MarkedAsRollback(id portainer.EdgeStackID, fileVersion int) bool {
	return false
}

func (testStaggerService) WasEndpointRolledBack(id portainer.EdgeStackID, fileVersion int, endpointId portainer.EndpointID) bool {
	return false
}

func (testStaggerService) MarkedAsCompleted(id portainer.EdgeStackID, fileVersion int) bool {
	return false
}

func (testStaggerService) UpdateStaggerEndpointStatusIfNeeds(id portainer.EdgeStackID, fileVersion int, rollbackTo *int, endpointID portainer.EndpointID, status portainer.EdgeStackStatusType) {
}

func (testStaggerService) DisplayStaggerInfo() {

}

func (testStaggerService) ProcessStatusJob(newStatusJob *staggers.StaggerStatusJob) {

}

func (testStaggerService) StartStaggerJobForAsyncUpdate(edgeStackID portainer.EdgeStackID, relatedEndpointIds []portainer.EndpointID, endpointsToAdd set.Set[portainer.EndpointID], stackFileVersion int) {

}

func (testStaggerService) StartStaggerJobForAsyncRollback(ctx context.Context, wg *sync.WaitGroup, edgeStackID portainer.EdgeStackID, stackFileVersion int, endpoints map[portainer.EndpointID]*portaineree.Endpoint) {

}

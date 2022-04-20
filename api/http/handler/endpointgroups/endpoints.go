package endpointgroups

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

func (handler *Handler) updateEndpointRelations(endpoint *portaineree.Endpoint, endpointGroup *portaineree.EndpointGroup) error {
	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return nil
	}

	if endpointGroup == nil {
		unassignedGroup, err := handler.DataStore.EndpointGroup().EndpointGroup(portaineree.EndpointGroupID(1))
		if err != nil {
			return err
		}

		endpointGroup = unassignedGroup
	}

	endpointRelation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpoint.ID)
	if err != nil {
		return err
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return err
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return err
	}

	endpointStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
	stacksSet := map[portaineree.EdgeStackID]bool{}
	for _, edgeStackID := range endpointStacks {
		stacksSet[edgeStackID] = true
	}
	endpointRelation.EdgeStacks = stacksSet

	return handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpoint.ID, endpointRelation)
}

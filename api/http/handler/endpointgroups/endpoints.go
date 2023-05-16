package endpointgroups

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

func (handler *Handler) updateEndpointRelations(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, endpointGroup *portaineree.EndpointGroup) error {
	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return nil
	}

	if endpointGroup == nil {
		unassignedGroup, err := tx.EndpointGroup().EndpointGroup(portaineree.EndpointGroupID(1))
		if err != nil {
			return err
		}

		endpointGroup = unassignedGroup
	}

	endpointRelation, err := tx.EndpointRelation().EndpointRelation(endpoint.ID)
	if err != nil {
		return err
	}

	edgeGroups, err := tx.EdgeGroup().EdgeGroups()
	if err != nil {
		return err
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		return err
	}

	endpointStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
	stacksSet := map[portaineree.EdgeStackID]bool{}
	for _, edgeStackID := range endpointStacks {
		stacksSet[edgeStackID] = true
	}
	endpointRelation.EdgeStacks = stacksSet

	return tx.EndpointRelation().UpdateEndpointRelation(endpoint.ID, endpointRelation)
}

package endpointgroups

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/handler/endpoints"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
)

func (handler *Handler) updateEndpointRelations(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, endpointGroup *portainer.EndpointGroup) error {
	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return nil
	}

	if endpointGroup == nil {
		unassignedGroup, err := tx.EndpointGroup().Read(portainer.EndpointGroupID(1))
		if err != nil {
			return err
		}

		endpointGroup = unassignedGroup
	}

	endpointRelation, err := tx.EndpointRelation().EndpointRelation(endpoint.ID)
	if err != nil {
		return err
	}

	edgeGroups, err := tx.EdgeGroup().ReadAll()
	if err != nil {
		return err
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		return err
	}

	endpointStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
	stacksSet := map[portainer.EdgeStackID]bool{}
	for _, edgeStackID := range endpointStacks {
		stacksSet[edgeStackID] = true
	}
	endpointRelation.EdgeStacks = stacksSet

	if err := endpoints.UpdateEdgeConfigs(tx, handler.edgeAsyncService, endpoint); err != nil {
		return err
	}

	return tx.EndpointRelation().UpdateEndpointRelation(endpoint.ID, endpointRelation)
}

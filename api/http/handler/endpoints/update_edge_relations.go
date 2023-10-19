package endpoints

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/edge/edgeconfigtrigger"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/set"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
)

// updateEdgeRelations updates the edge stacks associated to an edge endpoint
func (handler *Handler) updateEdgeRelations(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) error {
	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return nil
	}

	relation, err := tx.EndpointRelation().EndpointRelation(endpoint.ID)
	if err != nil {
		return errors.WithMessage(err, "Unable to find environment relation inside the database")
	}

	endpointGroup, err := tx.EndpointGroup().Read(endpoint.GroupID)
	if err != nil {
		return errors.WithMessage(err, "Unable to find environment group inside the database")
	}

	edgeGroups, err := tx.EdgeGroup().ReadAll()
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve edge groups from the database")
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve edge stacks from the database")
	}

	existingEdgeStacks := relation.EdgeStacks

	currentEdgeStackSet := set.Set[portainer.EdgeStackID]{}
	currentEndpointEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
	for _, edgeStackID := range currentEndpointEdgeStacks {
		currentEdgeStackSet[edgeStackID] = true
		if !existingEdgeStacks[edgeStackID] {
			edgeStack, err := tx.EdgeStack().EdgeStack(edgeStackID)
			if err != nil {
				return errors.WithMessage(err, "Unable to retrieve edge stack from the database")
			}

			err = handler.edgeService.AddStackCommandTx(tx, endpoint, edgeStackID, edgeStack.ScheduledTime)
			if err != nil {
				return errors.WithMessage(err, "Unable to store edge async command into the database")
			}
		}
	}

	for existingEdgeStackID := range existingEdgeStacks {
		if !currentEdgeStackSet[existingEdgeStackID] {
			err = handler.edgeService.RemoveStackCommandTx(tx, endpoint.ID, existingEdgeStackID)
			if err != nil {
				return errors.WithMessage(err, "Unable to store edge async command into the database")
			}
		}
	}

	relation.EdgeStacks = currentEdgeStackSet

	err = tx.EndpointRelation().UpdateEndpointRelation(endpoint.ID, relation)
	if err != nil {
		return errors.WithMessage(err, "Unable to persist environment relation changes inside the database")
	}

	err = edgeconfigtrigger.UpdateForEndpoint(tx, handler.edgeService, endpoint)
	if err != nil {
		return errors.WithMessage(err, "Unable to update edge configurations")
	}

	return nil
}

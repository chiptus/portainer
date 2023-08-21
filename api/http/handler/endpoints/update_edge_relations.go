package endpoints

import (
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/set"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/portainer/portainer-ee/api/internal/unique"
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

	currentEdgeStackSet := set.Set[portaineree.EdgeStackID]{}
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

	err = handler.updateEdgeConfigs(tx, endpoint)
	if err != nil {
		return errors.WithMessage(err, "Unable to update edge configurations")
	}

	return nil
}

func (handler *Handler) updateEdgeConfigs(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) error {
	edgeConfigs, err := tx.EdgeConfig().ReadAll()
	if err != nil || len(edgeConfigs) == 0 {
		return err
	}

	endpointGroups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return err
	}

	edgeGroupsToEdgeConfigs := make(map[portaineree.EdgeGroupID][]portaineree.EdgeConfigID)
	for _, edgeConfig := range edgeConfigs {
		for _, edgeGroupID := range edgeConfig.EdgeGroupIDs {
			edgeGroupsToEdgeConfigs[edgeGroupID] = append(edgeGroupsToEdgeConfigs[edgeGroupID], edgeConfig.ID)
		}
	}

	endpoints := []portaineree.Endpoint{*endpoint}

	var edgeConfigsToCreate []portaineree.EdgeConfigID

	for edgeGroupID, edgeConfigIDs := range edgeGroupsToEdgeConfigs {
		edgeGroup, err := tx.EdgeGroup().Read(edgeGroupID)
		if err != nil {
			return err
		}

		relatedEndpointIDs := edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)

		for _, relEndpointID := range relatedEndpointIDs {
			if relEndpointID == endpoint.ID {
				edgeConfigsToCreate = append(edgeConfigsToCreate, edgeConfigIDs...)

				break
			}
		}
	}

	// Edge Configs to create
	edgeConfigsToCreate = unique.Unique(edgeConfigsToCreate)

	for _, edgeConfigID := range edgeConfigsToCreate {
		// Update the Edge Config
		edgeConfig, err := tx.EdgeConfig().Read(edgeConfigID)
		if err != nil {
			return err
		}

		switch edgeConfig.State {
		case portaineree.EdgeConfigFailureState, portaineree.EdgeConfigDeletingState:
			continue
		}

		edgeConfig.Progress.Total++

		if err = tx.EdgeConfig().Update(edgeConfigID, edgeConfig); err != nil {
			return err
		}

		// Update or create the Edge Config State
		edgeConfigState, err := tx.EdgeConfigState().Read(endpoint.ID)
		if err != nil {
			edgeConfigState = &portaineree.EdgeConfigState{
				EndpointID: endpoint.ID,
				States:     make(map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType),
			}

			if err = tx.EdgeConfigState().Create(edgeConfigState); err != nil {
				return err
			}
		}

		edgeConfigState.States[edgeConfigID] = portaineree.EdgeConfigSavingState

		if err = tx.EdgeConfigState().Update(edgeConfigState.EndpointID, edgeConfigState); err != nil {
			return err
		}

		if err = handler.edgeService.PushConfigCommand(tx, endpoint, edgeConfig, edgeConfigState); err != nil {
			return httperror.InternalServerError("Unable to persist the edge configuration command inside the database", err)
		}
	}

	// Edge Configs to remove
	edgeConfigState, err := tx.EdgeConfigState().Read(endpoint.ID)
	if dataservices.IsErrObjectNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, edgeConfig := range edgeConfigs {
		if _, ok := edgeConfigState.States[edgeConfig.ID]; !ok {
			continue
		}

		if slices.Contains(edgeConfigsToCreate, edgeConfig.ID) {
			continue
		}

		edgeConfig, err := tx.EdgeConfig().Read(edgeConfig.ID)
		if err != nil {
			return err
		}

		edgeConfigState.States[edgeConfig.ID] = portaineree.EdgeConfigDeletingState
	}

	if err := tx.EdgeConfigState().Update(edgeConfigState.EndpointID, edgeConfigState); err != nil {
		return err
	}

	cache.Del(endpoint.ID)

	return nil
}

package edgeconfigtrigger

import (
	"slices"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/unique"
	portainer "github.com/portainer/portainer/api"
)

func UpdateForEndpoint(tx dataservices.DataStoreTx, edgeAsyncService *edgeasync.Service, endpoint *portaineree.Endpoint) error {
	if endpoint.Type != portaineree.EdgeAgentOnDockerEnvironment {
		return nil
	}

	edgeConfigs, err := tx.EdgeConfig().ReadAll()
	if err != nil || len(edgeConfigs) == 0 {
		return err
	}

	endpointGroups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return err
	}

	edgeGroupsToEdgeConfigs := make(map[portainer.EdgeGroupID][]portaineree.EdgeConfigID)
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

		if _, ok := edgeConfigState.States[edgeConfigID]; ok {
			continue
		}

		edgeConfig.Progress.Total++

		if err = tx.EdgeConfig().Update(edgeConfigID, edgeConfig); err != nil {
			return err
		}

		edgeConfigState.States[edgeConfigID] = portaineree.EdgeConfigSavingState

		if err = tx.EdgeConfigState().Update(edgeConfigState.EndpointID, edgeConfigState); err != nil {
			return err
		}

		if err = edgeAsyncService.PushConfigCommand(tx, endpoint, edgeConfig, edgeConfigState); err != nil {
			return err
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

		if err = edgeAsyncService.PushConfigCommand(tx, endpoint, edgeConfig, edgeConfigState); err != nil {
			return err
		}
	}

	if err := tx.EdgeConfigState().Update(edgeConfigState.EndpointID, edgeConfigState); err != nil {
		return err
	}

	cache.Del(endpoint.ID)

	return nil
}

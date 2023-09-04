package utils

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/rs/zerolog/log"
)

func EndpointPendingActions(endpoint *portaineree.Endpoint) *portaineree.EndpointPendingActions {
	return endpoint.PendingActions
}

func GetUpdatedEndpointPendingActions(endpoint *portaineree.Endpoint, action string, value interface{}) *portaineree.EndpointPendingActions {
	if endpoint.PendingActions == nil {
		endpoint.PendingActions = &portaineree.EndpointPendingActions{}
	}

	switch action {
	case "CleanNAPWithOverridePolicies":
		endpoint.PendingActions.CleanNAPWithOverridePolicies.EndpointGroups = append(endpoint.PendingActions.CleanNAPWithOverridePolicies.EndpointGroups, value.(portaineree.EndpointGroupID))
	}

	return endpoint.PendingActions
}

func RunPendingActions(endpoint *portaineree.Endpoint, dataStore dataservices.DataStoreTx, authorizationService *authorization.Service) error {

	if endpoint.PendingActions == nil {
		return nil
	}

	log.Info().Msgf("Running pending actions for endpoint %d", endpoint.ID)

	if endpoint.PendingActions.CleanNAPWithOverridePolicies.EndpointGroups != nil {
		log.Info().Int("endpoint_id", int(endpoint.ID)).Msgf("Cleaning NAP with override policies for endpoint groups %v", endpoint.PendingActions.CleanNAPWithOverridePolicies.EndpointGroups)
		failedEndpointGroupIDs := make([]portaineree.EndpointGroupID, 0)
		for _, endpointGroupID := range endpoint.PendingActions.CleanNAPWithOverridePolicies.EndpointGroups {
			endpointGroup, err := dataStore.EndpointGroup().Read(portaineree.EndpointGroupID(endpointGroupID))
			if err != nil {
				log.Error().Err(err).Msgf("Error reading endpoint group to clean NAP with override policies for endpoint %d and endpoint group %d", endpoint.ID, endpointGroup.ID)
				failedEndpointGroupIDs = append(failedEndpointGroupIDs, endpointGroupID)
				continue
			}
			err = authorizationService.CleanNAPWithOverridePolicies(dataStore, endpoint, endpointGroup)
			if err != nil {
				failedEndpointGroupIDs = append(failedEndpointGroupIDs, endpointGroupID)
				log.Error().Err(err).Msgf("Error cleaning NAP with override policies for endpoint %d and endpoint group %d", endpoint.ID, endpointGroup.ID)
			}
		}

		endpoint.PendingActions.CleanNAPWithOverridePolicies.EndpointGroups = failedEndpointGroupIDs
		err := dataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("While running pending actions, error updating endpoint %d", endpoint.ID)
			return err
		}
	}

	return nil
}

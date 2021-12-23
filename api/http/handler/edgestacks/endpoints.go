package edgestacks

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

func hasKubeEndpoint(endpointService portaineree.EndpointService, endpointIDs []portaineree.EndpointID) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, endpointutils.IsKubernetesEndpoint)
}

func hasDockerEndpoint(endpointService portaineree.EndpointService, endpointIDs []portaineree.EndpointID) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, endpointutils.IsDockerEndpoint)
}

func hasEndpointPredicate(endpointService portaineree.EndpointService, endpointIDs []portaineree.EndpointID, predicate func(*portaineree.Endpoint) bool) (bool, error) {
	for _, endpointID := range endpointIDs {
		endpoint, err := endpointService.Endpoint(endpointID)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve environment from database: %w", err)
		}

		if predicate(endpoint) {
			return true, nil
		}
	}

	return false, nil
}

type endpointRelationsConfig struct {
	endpoints      []portaineree.Endpoint
	endpointGroups []portaineree.EndpointGroup
	edgeGroups     []portaineree.EdgeGroup
}

func fetchEndpointRelationsConfig(dataStore portaineree.DataStore) (*endpointRelationsConfig, error) {
	endpoints, err := dataStore.Endpoint().Endpoints()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve environments from database: %w", err)
	}

	endpointGroups, err := dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve environment groups from database: %w", err)
	}

	edgeGroups, err := dataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve edge groups from database: %w", err)
	}

	return &endpointRelationsConfig{
		endpoints:      endpoints,
		endpointGroups: endpointGroups,
		edgeGroups:     edgeGroups,
	}, nil
}

package edge

import (
	"errors"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

// EdgeStackRelatedEndpoints returns a list of environments(endpoints) related to this Edge stack
func EdgeStackRelatedEndpoints(edgeGroupIDs []portaineree.EdgeGroupID, endpoints []portaineree.Endpoint, endpointGroups []portaineree.EndpointGroup, edgeGroups []portaineree.EdgeGroup) ([]portaineree.EndpointID, error) {
	edgeStackEndpoints := []portaineree.EndpointID{}

	for _, edgeGroupID := range edgeGroupIDs {
		var edgeGroup *portaineree.EdgeGroup

		for _, group := range edgeGroups {
			if group.ID == edgeGroupID {
				edgeGroup = &group
				break
			}
		}

		if edgeGroup == nil {
			return nil, errors.New("Edge group was not found")
		}

		edgeStackEndpoints = append(edgeStackEndpoints, EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)...)
	}

	return edgeStackEndpoints, nil
}

type EndpointRelationsConfig struct {
	Endpoints      []portaineree.Endpoint
	EndpointGroups []portaineree.EndpointGroup
	EdgeGroups     []portaineree.EdgeGroup
}

// FetchEndpointRelationsConfig fetches config needed for Edge Stack related endpoints
func FetchEndpointRelationsConfig(dataStore dataservices.DataStore) (*EndpointRelationsConfig, error) {
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

	return &EndpointRelationsConfig{
		Endpoints:      endpoints,
		EndpointGroups: endpointGroups,
		EdgeGroups:     edgeGroups,
	}, nil
}

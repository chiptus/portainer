package edge

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
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

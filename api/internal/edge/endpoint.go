package edge

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// EndpointRelatedEdgeStacks returns a list of Edge stacks related to this Environment(Endpoint)
func EndpointRelatedEdgeStacks(endpoint *portaineree.Endpoint, endpointGroup *portaineree.EndpointGroup, edgeGroups []portaineree.EdgeGroup, edgeStacks []portaineree.EdgeStack) []portaineree.EdgeStackID {
	relatedEdgeGroupsSet := map[portaineree.EdgeGroupID]bool{}

	for _, edgeGroup := range edgeGroups {
		if edgeGroupRelatedToEndpoint(&edgeGroup, endpoint, endpointGroup) {
			relatedEdgeGroupsSet[edgeGroup.ID] = true
		}
	}

	relatedEdgeStacks := []portaineree.EdgeStackID{}
	for _, edgeStack := range edgeStacks {
		for _, edgeGroupID := range edgeStack.EdgeGroups {
			if relatedEdgeGroupsSet[edgeGroupID] {
				relatedEdgeStacks = append(relatedEdgeStacks, edgeStack.ID)
				break
			}
		}
	}

	return relatedEdgeStacks
}

func EdgeEndpoint(endpoints []portaineree.Endpoint, edgeIdentifier string) *portaineree.Endpoint {
	for _, endpoint := range endpoints {
		if endpoint.EdgeID == edgeIdentifier {
			return &endpoint
		}
	}
	return nil
}

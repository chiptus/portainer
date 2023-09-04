package edge

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/tag"
)

// EdgeGroupRelatedEndpoints returns a list of environments(endpoints) related to this Edge group
func EdgeGroupRelatedEndpoints(edgeGroup *portaineree.EdgeGroup, endpoints []portaineree.Endpoint, endpointGroups []portaineree.EndpointGroup) []portaineree.EndpointID {
	if !edgeGroup.Dynamic {
		return edgeGroup.Endpoints
	}

	endpointIDs := []portaineree.EndpointID{}
	for _, endpoint := range endpoints {
		if !endpointutils.IsEdgeEndpoint(&endpoint) {
			continue
		}

		var endpointGroup portaineree.EndpointGroup
		for _, group := range endpointGroups {
			if endpoint.GroupID == group.ID {
				endpointGroup = group

				break
			}
		}

		if edgeGroupRelatedToEndpoint(edgeGroup, &endpoint, &endpointGroup) {
			endpointIDs = append(endpointIDs, endpoint.ID)
		}
	}

	return endpointIDs
}

func EdgeGroupSet(edgeGroupIDs []portaineree.EdgeGroupID) map[portaineree.EdgeGroupID]bool {
	set := map[portaineree.EdgeGroupID]bool{}

	for _, edgeGroupID := range edgeGroupIDs {
		set[edgeGroupID] = true
	}

	return set
}

func GetEndpointsFromEdgeGroups(edgeGroupIDs []portaineree.EdgeGroupID, datastore dataservices.DataStoreTx) ([]portaineree.EndpointID, error) {
	endpoints, err := datastore.Endpoint().Endpoints()
	if err != nil {
		return nil, err
	}

	endpointGroups, err := datastore.EndpointGroup().ReadAll()
	if err != nil {
		return nil, err
	}

	var response []portaineree.EndpointID
	for _, edgeGroupID := range edgeGroupIDs {
		edgeGroup, err := datastore.EdgeGroup().Read(edgeGroupID)
		if err != nil {
			return nil, err
		}

		response = append(response, EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)...)
	}

	return response, nil
}

// edgeGroupRelatedToEndpoint returns true if edgeGroup is associated with environment(endpoint)
func edgeGroupRelatedToEndpoint(edgeGroup *portaineree.EdgeGroup, endpoint *portaineree.Endpoint, endpointGroup *portaineree.EndpointGroup) bool {
	if !edgeGroup.Dynamic {
		for _, endpointID := range edgeGroup.Endpoints {
			if endpoint.ID == endpointID {
				return true
			}
		}

		return false
	}

	endpointTags := tag.Set(endpoint.TagIDs)
	if endpointGroup.TagIDs != nil {
		endpointTags = tag.Union(endpointTags, tag.Set(endpointGroup.TagIDs))
	}

	edgeGroupTags := tag.Set(edgeGroup.TagIDs)

	if edgeGroup.PartialMatch {
		return tag.PartialMatch(edgeGroupTags, endpointTags)
	}

	return tag.FullMatch(edgeGroupTags, endpointTags)
}

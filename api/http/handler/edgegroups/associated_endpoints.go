package edgegroups

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
)

type endpointSetType map[portainer.EndpointID]bool

func GetEndpointsByTags(tx dataservices.DataStoreTx, tagIDs []portainer.TagID, partialMatch bool) ([]portainer.EndpointID, error) {
	if len(tagIDs) == 0 {
		return []portainer.EndpointID{}, nil
	}

	endpoints, err := tx.Endpoint().Endpoints()
	if err != nil {
		return nil, err
	}

	groupEndpoints := mapEndpointGroupToEndpoints(endpoints)

	tags := []portainer.Tag{}
	for _, tagID := range tagIDs {
		tag, err := tx.Tag().Read(tagID)
		if err != nil {
			return nil, err
		}

		tags = append(tags, *tag)
	}

	setsOfEndpoints := mapTagsToEndpoints(tags, groupEndpoints)

	var endpointSet endpointSetType
	if partialMatch {
		endpointSet = setsUnion(setsOfEndpoints)
	} else {
		endpointSet = setsIntersection(setsOfEndpoints)
	}

	results := []portainer.EndpointID{}
	for _, endpoint := range endpoints {
		if _, ok := endpointSet[endpoint.ID]; ok && endpointutils.IsEdgeEndpoint(&endpoint) && endpoint.UserTrusted {
			results = append(results, endpoint.ID)
		}
	}

	return results, nil
}

func getTrustedEndpoints(tx dataservices.DataStoreTx, endpointIDs []portainer.EndpointID) ([]portainer.EndpointID, error) {
	results := []portainer.EndpointID{}
	for _, endpointID := range endpointIDs {
		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return nil, err
		}

		if !endpoint.UserTrusted {
			continue
		}

		results = append(results, endpoint.ID)
	}

	return results, nil
}

func mapEndpointGroupToEndpoints(endpoints []portaineree.Endpoint) map[portainer.EndpointGroupID]endpointSetType {
	groupEndpoints := map[portainer.EndpointGroupID]endpointSetType{}

	for _, endpoint := range endpoints {
		groupID := endpoint.GroupID
		if groupEndpoints[groupID] == nil {
			groupEndpoints[groupID] = endpointSetType{}
		}

		groupEndpoints[groupID][endpoint.ID] = true
	}

	return groupEndpoints
}

func mapTagsToEndpoints(tags []portainer.Tag, groupEndpoints map[portainer.EndpointGroupID]endpointSetType) []endpointSetType {
	sets := []endpointSetType{}

	for _, tag := range tags {
		set := tag.Endpoints

		for groupID := range tag.EndpointGroups {
			for endpointID := range groupEndpoints[groupID] {
				set[endpointID] = true
			}
		}

		sets = append(sets, set)
	}

	return sets
}

func setsIntersection(sets []endpointSetType) endpointSetType {
	if len(sets) == 0 {
		return endpointSetType{}
	}

	intersectionSet := sets[0]

	for _, set := range sets {
		for endpointID := range intersectionSet {
			if !set[endpointID] {
				delete(intersectionSet, endpointID)
			}
		}
	}

	return intersectionSet
}

func setsUnion(sets []endpointSetType) endpointSetType {
	unionSet := endpointSetType{}

	for _, set := range sets {
		for endpointID := range set {
			unionSet[endpointID] = true
		}
	}

	return unionSet
}

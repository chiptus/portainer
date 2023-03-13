package edge

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/set"
	"github.com/rs/zerolog/log"
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

func AddEnvironmentToEdgeGroups(dataStore dataservices.DataStore, endpoint *portaineree.Endpoint, edgeGroupsIDs []portaineree.EdgeGroupID) error {
	for _, edgeGroupID := range edgeGroupsIDs {
		edgeGroup, err := dataStore.EdgeGroup().EdgeGroup(edgeGroupID)
		if err != nil {
			log.Warn().
				Err(err).
				Int("edgeGroupID", int(edgeGroupID)).
				Msg("Unable to retrieve edge group")
			continue
		}

		if edgeGroup.Dynamic {
			log.Warn().
				Int("edgeGroupID", int(edgeGroupID)).
				Msg("Unable to add endpoint to dynamic edge group")
			continue
		}

		edgeGroup.Endpoints = append(edgeGroup.Endpoints, endpoint.ID)
		err = dataStore.EdgeGroup().UpdateEdgeGroup(edgeGroupID, edgeGroup)
		if err != nil {
			log.Warn().
				Err(err).
				Int("edgeGroupID", int(edgeGroupID)).
				Msg("Unable to persist edge group changes inside the database")
		}
	}

	relation := &portaineree.EndpointRelation{
		EndpointID: portaineree.EndpointID(endpoint.ID),
		EdgeStacks: map[portaineree.EdgeStackID]bool{},
	}

	relationConfig, err := FetchEndpointRelationsConfig(dataStore)
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve environments relations config from database")
	}

	edgeStacks, err := dataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve edge stacks from database")
	}

	environmentGroup, err := dataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve environment group from database")
	}

	edgeStackSet := set.Set[portaineree.EdgeStackID]{}
	endpointEdgeStacks := EndpointRelatedEdgeStacks(endpoint, environmentGroup, relationConfig.EdgeGroups, edgeStacks)
	for _, edgeStackID := range endpointEdgeStacks {
		edgeStackSet[edgeStackID] = true
	}

	relation.EdgeStacks = edgeStackSet

	err = dataStore.EndpointRelation().Create(relation)
	if err != nil {
		return errors.WithMessage(err, "Unable to persist the relation object inside the database")
	}
	return nil
}

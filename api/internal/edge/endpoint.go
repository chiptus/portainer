package edge

import (
	"slices"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/set"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

// EndpointRelatedEdgeStacks returns a list of Edge stacks related to this Environment(Endpoint)
func EndpointRelatedEdgeStacks(endpoint *portaineree.Endpoint, endpointGroup *portainer.EndpointGroup, edgeGroups []portaineree.EdgeGroup, edgeStacks []portaineree.EdgeStack) []portainer.EdgeStackID {
	relatedEdgeGroupsSet := map[portainer.EdgeGroupID]bool{}

	for _, edgeGroup := range edgeGroups {
		if edgeGroupRelatedToEndpoint(&edgeGroup, endpoint, endpointGroup) {
			relatedEdgeGroupsSet[edgeGroup.ID] = true
		}
	}

	relatedEdgeStacks := []portainer.EdgeStackID{}
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

func AddEnvironmentToEdgeGroups(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeGroupsIDs []portainer.EdgeGroupID) error {
	for _, edgeGroupID := range edgeGroupsIDs {
		edgeGroup, err := tx.EdgeGroup().Read(edgeGroupID)
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
		err = tx.EdgeGroup().Update(edgeGroupID, edgeGroup)
		if err != nil {
			log.Warn().
				Err(err).
				Int("edgeGroupID", int(edgeGroupID)).
				Msg("Unable to persist edge group changes inside the database")
		}
	}

	relation := &portainer.EndpointRelation{
		EndpointID: portainer.EndpointID(endpoint.ID),
		EdgeStacks: map[portainer.EdgeStackID]bool{},
	}

	relationConfig, err := FetchEndpointRelationsConfig(tx)
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve environments relations config from database")
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve edge stacks from database")
	}

	environmentGroup, err := tx.EndpointGroup().Read(endpoint.GroupID)
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve environment group from database")
	}

	edgeStackSet := set.Set[portainer.EdgeStackID]{}
	endpointEdgeStacks := EndpointRelatedEdgeStacks(endpoint, environmentGroup, relationConfig.EdgeGroups, edgeStacks)
	for _, edgeStackID := range endpointEdgeStacks {
		edgeStackSet[edgeStackID] = true
	}

	relation.EdgeStacks = edgeStackSet

	err = tx.EndpointRelation().Create(relation)
	if err != nil {
		return errors.WithMessage(err, "Unable to persist the relation object inside the database")
	}

	return nil
}

// EndpointInEdgeGroup returns true and the edge group name if the endpoint in the edge group
func EndpointInEdgeGroup(
	tx dataservices.DataStoreTx,
	endpointID portainer.EndpointID,
	edgeGroupID portainer.EdgeGroupID,
) (bool, string, error) {
	endpointIDs, err := GetEndpointsFromEdgeGroups(
		[]portainer.EdgeGroupID{edgeGroupID},
		tx,
	)
	if err != nil {
		return false, "", err
	}

	if slices.Contains(endpointIDs, endpointID) {
		edgeGroup, err := tx.EdgeGroup().Read(edgeGroupID)
		if err != nil {
			log.Warn().
				Err(err).
				Int("edgeGroupID", int(edgeGroupID)).
				Msg("Unable to retrieve edge group")
			return false, "", err
		}

		return true, edgeGroup.Name, nil
	}

	return false, "", nil
}

// GetEndpointEdgeGroupNames returns edge group names where endpointID is in
func GetEndpointEdgeGroupNames(
	tx dataservices.DataStoreTx,
	endpointID portainer.EndpointID,
	edgeGroupIDs []portainer.EdgeGroupID,
) ([]string, error) {
	edgeGroupNames := []string{}
	for _, edgeGroupID := range edgeGroupIDs {
		in, edgeGroupName, err := EndpointInEdgeGroup(tx, endpointID, edgeGroupID)
		if err != nil {
			return edgeGroupNames, err
		}

		if in {
			edgeGroupNames = append(edgeGroupNames, edgeGroupName)
		}
	}

	return edgeGroupNames, nil
}

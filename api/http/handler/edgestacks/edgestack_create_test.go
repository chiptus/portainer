package edgestacks

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_updateEndpointRelation_successfulRuns(t *testing.T) {
	edgeStackID := portaineree.EdgeStackID(5)
	endpointRelations := []portaineree.EndpointRelation{
		{EndpointID: 1, EdgeStacks: map[portaineree.EdgeStackID]bool{}},
		{EndpointID: 2, EdgeStacks: map[portaineree.EdgeStackID]bool{}},
		{EndpointID: 3, EdgeStacks: map[portaineree.EdgeStackID]bool{}},
		{EndpointID: 4, EdgeStacks: map[portaineree.EdgeStackID]bool{}},
		{EndpointID: 5, EdgeStacks: map[portaineree.EdgeStackID]bool{}},
	}

	relatedIds := []portaineree.EndpointID{2, 3}

	dataStore := testhelpers.NewDatastore(testhelpers.WithEndpointRelations(endpointRelations))

	err := updateEndpointRelations(dataStore.EndpointRelation(), edgeStackID, relatedIds)

	assert.NoError(t, err, "updateEndpointRelations should not fail")

	relatedSet := map[portaineree.EndpointID]bool{}
	for _, relationID := range relatedIds {
		relatedSet[relationID] = true
	}

	for _, relation := range endpointRelations {
		shouldBeRelated := relatedSet[relation.EndpointID]
		assert.Equal(t, shouldBeRelated, relation.EdgeStacks[edgeStackID])
	}
}

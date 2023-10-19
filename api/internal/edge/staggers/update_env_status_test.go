package staggers

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func mockEdgeStack(dataStore dataservices.DataStore) (*portaineree.EdgeStack, error) {
	edgeStack := &portaineree.EdgeStack{
		ID: 1,
		Status: map[portainer.EndpointID]portainer.EdgeStackStatus{
			1: {},
			2: {},
			3: {
				Status: []portainer.EdgeStackDeploymentStatus{
					{
						Type: portainer.EdgeStackStatusAcknowledged,
						Time: 1697163145,
					},
					{
						Type: portainer.EdgeStackStatusDeploying,
						Time: 1697163148,
					},
					{
						Type: portainer.EdgeStackStatusError,
						Time: 1697163148,
					},
				},
			},
			4: {},
			5: {
				Status: []portainer.EdgeStackDeploymentStatus{
					{
						Type: portainer.EdgeStackStatusAcknowledged,
						Time: 1697163157,
					},
					{
						Type: portainer.EdgeStackStatusDeploying,
						Time: 1697163157,
					},
					{
						Type: portainer.EdgeStackStatusDeploymentReceived,
						Time: 1697163158,
					},
					{
						Type: portainer.EdgeStackStatusRunning,
						Time: 1697163160,
					},
				},
			},
			6: {
				Status: []portainer.EdgeStackDeploymentStatus{
					{
						Type: portainer.EdgeStackStatusAcknowledged,
						Time: 1697163144,
					},
					{
						Type: portainer.EdgeStackStatusDeploying,
						Time: 1697163147,
					},
					{
						Type: portainer.EdgeStackStatusDeploymentReceived,
						Time: 1697163148,
					},
					{
						Type: portainer.EdgeStackStatusRunning,
						Time: 1697163150,
					},
				},
			},
		},
	}

	err := dataStore.EdgeStack().Create(edgeStack.ID, edgeStack)
	return edgeStack, err
}

func TestUpdatePausedEnvironmentStatus(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)
	edgeStack, err := mockEdgeStack(store)
	if err != nil {
		t.Fatal(err)
	}

	expectedStatus := map[portainer.EndpointID]portainer.EdgeStackStatus{
		1: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusPausedDeploying,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
				},
			},
		},
		2: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusPausedDeploying,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
				},
			},
		},
		3: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: 1697163145,
				},
				{
					Type: portainer.EdgeStackStatusDeploying,
					Time: 1697163148,
				},
				{
					Type: portainer.EdgeStackStatusError,
					Time: 1697163148,
				},
			},
		},
		4: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusPausedDeploying,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
				},
			},
		},
		5: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: 1697163157,
				},
				{
					Type: portainer.EdgeStackStatusDeploying,
					Time: 1697163157,
				},
				{
					Type: portainer.EdgeStackStatusDeploymentReceived,
					Time: 1697163158,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
					Time: 1697163160,
				},
			},
		},
		6: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: 1697163144,
				},
				{
					Type: portainer.EdgeStackStatusDeploying,
					Time: 1697163147,
				},
				{
					Type: portainer.EdgeStackStatusDeploymentReceived,
					Time: 1697163148,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
					Time: 1697163150,
				},
			},
		},
	}

	updateEnvironmentStatus(store, edgeStack.ID, UpdatePausedEnvironmentStatus)

	updatedEdgeStack, err := store.EdgeStack().EdgeStack(edgeStack.ID)
	if err != nil {
		t.Fatal(err)
	}

	for i, status := range updatedEdgeStack.Status {
		for j, sts := range status.Status {
			is.Equal(expectedStatus[i].Status[j].Type, sts.Type)
		}
	}
}

func TestUpdateRollingbackEnvironmentStatus(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)
	edgeStack, err := mockEdgeStack(store)
	if err != nil {
		t.Fatal(err)
	}

	expectedStatus := map[portainer.EndpointID]portainer.EdgeStackStatus{
		1: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusRunning,
				},
			},
		},
		2: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusRunning,
				},
			},
		},
		3: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: 1697163145,
				},
				{
					Type: portainer.EdgeStackStatusDeploying,
					Time: 1697163148,
				},
				{
					Type: portainer.EdgeStackStatusError,
					Time: 1697163148,
				},
			},
		},
		4: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusRunning,
				},
			},
		},
		5: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: 1697163157,
				},
				{
					Type: portainer.EdgeStackStatusDeploying,
					Time: 1697163157,
				},
				{
					Type: portainer.EdgeStackStatusDeploymentReceived,
					Time: 1697163158,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
					Time: 1697163160,
				},
				{
					Type: portainer.EdgeStackStatusRollingBack,
				},
			},
		},
		6: {
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: 1697163144,
				},
				{
					Type: portainer.EdgeStackStatusDeploying,
					Time: 1697163147,
				},
				{
					Type: portainer.EdgeStackStatusDeploymentReceived,
					Time: 1697163148,
				},
				{
					Type: portainer.EdgeStackStatusRunning,
					Time: 1697163150,
				},
				{
					Type: portainer.EdgeStackStatusRollingBack,
				},
			},
		},
	}

	updateEnvironmentStatus(store, edgeStack.ID, UpdateRollingBackEnvironmentStatus)

	updatedEdgeStack, err := store.EdgeStack().EdgeStack(edgeStack.ID)
	if err != nil {
		t.Fatal(err)
	}

	for i, status := range updatedEdgeStack.Status {
		for j, sts := range status.Status {
			is.Equal(expectedStatus[i].Status[j].Type, sts.Type)
		}
	}
}

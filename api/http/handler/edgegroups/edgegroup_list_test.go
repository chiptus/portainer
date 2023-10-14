package edgegroups

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func Test_getEndpointTypes(t *testing.T) {
	endpoints := []portaineree.Endpoint{
		{ID: 1, Type: portaineree.DockerEnvironment},
		{ID: 2, Type: portaineree.AgentOnDockerEnvironment},
		{ID: 3, Type: portaineree.AzureEnvironment},
		{ID: 4, Type: portaineree.EdgeAgentOnDockerEnvironment},
		{ID: 5, Type: portaineree.KubernetesLocalEnvironment},
		{ID: 6, Type: portaineree.AgentOnKubernetesEnvironment},
		{ID: 7, Type: portaineree.EdgeAgentOnKubernetesEnvironment},
	}

	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints(endpoints))

	tests := []struct {
		endpointIds []portainer.EndpointID
		expected    []portainer.EndpointType
	}{
		{endpointIds: []portainer.EndpointID{1}, expected: []portainer.EndpointType{portaineree.DockerEnvironment}},
		{endpointIds: []portainer.EndpointID{2}, expected: []portainer.EndpointType{portaineree.AgentOnDockerEnvironment}},
		{endpointIds: []portainer.EndpointID{3}, expected: []portainer.EndpointType{portaineree.AzureEnvironment}},
		{endpointIds: []portainer.EndpointID{4}, expected: []portainer.EndpointType{portaineree.EdgeAgentOnDockerEnvironment}},
		{endpointIds: []portainer.EndpointID{5}, expected: []portainer.EndpointType{portaineree.KubernetesLocalEnvironment}},
		{endpointIds: []portainer.EndpointID{6}, expected: []portainer.EndpointType{portaineree.AgentOnKubernetesEnvironment}},
		{endpointIds: []portainer.EndpointID{7}, expected: []portainer.EndpointType{portaineree.EdgeAgentOnKubernetesEnvironment}},
		{endpointIds: []portainer.EndpointID{7, 2}, expected: []portainer.EndpointType{portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.AgentOnDockerEnvironment}},
		{endpointIds: []portainer.EndpointID{6, 4, 1}, expected: []portainer.EndpointType{portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnDockerEnvironment, portaineree.DockerEnvironment}},
		{endpointIds: []portainer.EndpointID{1, 2, 3}, expected: []portainer.EndpointType{portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.AzureEnvironment}},
	}

	for _, test := range tests {
		ans, err := getEndpointTypes(datastore, test.endpointIds)
		assert.NoError(t, err, "getEndpointTypes shouldn't fail")

		assert.ElementsMatch(t, test.expected, ans, "getEndpointTypes expected to return %b for %v, but returned %b", test.expected, test.endpointIds, ans)
	}
}

func Test_getEndpointTypes_failWhenEndpointDontExist(t *testing.T) {
	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portaineree.Endpoint{}))

	_, err := getEndpointTypes(datastore, []portainer.EndpointID{1})
	assert.Error(t, err, "getEndpointTypes should fail")
}

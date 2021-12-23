package edgegroups

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
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
		endpointIds []portaineree.EndpointID
		expected    []portaineree.EndpointType
	}{
		{endpointIds: []portaineree.EndpointID{1}, expected: []portaineree.EndpointType{portaineree.DockerEnvironment}},
		{endpointIds: []portaineree.EndpointID{2}, expected: []portaineree.EndpointType{portaineree.AgentOnDockerEnvironment}},
		{endpointIds: []portaineree.EndpointID{3}, expected: []portaineree.EndpointType{portaineree.AzureEnvironment}},
		{endpointIds: []portaineree.EndpointID{4}, expected: []portaineree.EndpointType{portaineree.EdgeAgentOnDockerEnvironment}},
		{endpointIds: []portaineree.EndpointID{5}, expected: []portaineree.EndpointType{portaineree.KubernetesLocalEnvironment}},
		{endpointIds: []portaineree.EndpointID{6}, expected: []portaineree.EndpointType{portaineree.AgentOnKubernetesEnvironment}},
		{endpointIds: []portaineree.EndpointID{7}, expected: []portaineree.EndpointType{portaineree.EdgeAgentOnKubernetesEnvironment}},
		{endpointIds: []portaineree.EndpointID{7, 2}, expected: []portaineree.EndpointType{portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.AgentOnDockerEnvironment}},
		{endpointIds: []portaineree.EndpointID{6, 4, 1}, expected: []portaineree.EndpointType{portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnDockerEnvironment, portaineree.DockerEnvironment}},
		{endpointIds: []portaineree.EndpointID{1, 2, 3}, expected: []portaineree.EndpointType{portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.AzureEnvironment}},
	}

	for _, test := range tests {
		ans, err := getEndpointTypes(datastore.Endpoint(), test.endpointIds)
		assert.NoError(t, err, "getEndpointTypes shouldn't fail")

		assert.ElementsMatch(t, test.expected, ans, "getEndpointTypes expected to return %b for %v, but returned %b", test.expected, test.endpointIds, ans)
	}
}

func Test_getEndpointTypes_failWhenEndpointDontExist(t *testing.T) {
	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portaineree.Endpoint{}))

	_, err := getEndpointTypes(datastore.Endpoint(), []portaineree.EndpointID{1})
	assert.Error(t, err, "getEndpointTypes should fail")
}

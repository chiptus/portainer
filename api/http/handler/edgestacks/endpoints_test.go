package edgestacks

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_hasKubeEndpoint(t *testing.T) {
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
		expected    bool
	}{
		{endpointIds: []portaineree.EndpointID{1}, expected: false},
		{endpointIds: []portaineree.EndpointID{2}, expected: false},
		{endpointIds: []portaineree.EndpointID{3}, expected: false},
		{endpointIds: []portaineree.EndpointID{4}, expected: false},
		{endpointIds: []portaineree.EndpointID{5}, expected: true},
		{endpointIds: []portaineree.EndpointID{6}, expected: true},
		{endpointIds: []portaineree.EndpointID{7}, expected: true},
		{endpointIds: []portaineree.EndpointID{7, 2}, expected: true},
		{endpointIds: []portaineree.EndpointID{6, 4, 1}, expected: true},
		{endpointIds: []portaineree.EndpointID{1, 2, 3}, expected: false},
	}

	for _, test := range tests {

		ans, err := hasKubeEndpoint(datastore.Endpoint(), test.endpointIds)
		assert.NoError(t, err, "hasKubeEndpoint shouldn't fail")

		assert.Equal(t, test.expected, ans, "hasKubeEndpoint expected to return %b for %v, but returned %b", test.expected, test.endpointIds, ans)
	}
}

func Test_hasKubeEndpoint_failWhenEndpointDontExist(t *testing.T) {
	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portaineree.Endpoint{}))

	_, err := hasKubeEndpoint(datastore.Endpoint(), []portaineree.EndpointID{1})
	assert.Error(t, err, "hasKubeEndpoint should fail")
}

func Test_hasDockerEndpoint(t *testing.T) {
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
		expected    bool
	}{
		{endpointIds: []portaineree.EndpointID{1}, expected: true},
		{endpointIds: []portaineree.EndpointID{2}, expected: true},
		{endpointIds: []portaineree.EndpointID{3}, expected: false},
		{endpointIds: []portaineree.EndpointID{4}, expected: true},
		{endpointIds: []portaineree.EndpointID{5}, expected: false},
		{endpointIds: []portaineree.EndpointID{6}, expected: false},
		{endpointIds: []portaineree.EndpointID{7}, expected: false},
		{endpointIds: []portaineree.EndpointID{7, 2}, expected: true},
		{endpointIds: []portaineree.EndpointID{6, 4, 1}, expected: true},
		{endpointIds: []portaineree.EndpointID{1, 2, 3}, expected: true},
	}

	for _, test := range tests {

		ans, err := hasDockerEndpoint(datastore.Endpoint(), test.endpointIds)
		assert.NoError(t, err, "hasDockerEndpoint shouldn't fail")

		assert.Equal(t, test.expected, ans, "hasDockerEndpoint expected to return %b for %v, but returned %b", test.expected, test.endpointIds, ans)
	}
}

func Test_hasDockerEndpoint_failWhenEndpointDontExist(t *testing.T) {
	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portaineree.Endpoint{}))

	_, err := hasDockerEndpoint(datastore.Endpoint(), []portaineree.EndpointID{1})
	assert.Error(t, err, "hasDockerEndpoint should fail")
}

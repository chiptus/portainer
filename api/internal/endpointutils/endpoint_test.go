package endpointutils

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/stretchr/testify/assert"
)

type isEndpointTypeTest struct {
	endpointType portaineree.EndpointType
	expected     bool
}

func Test_IsDockerEndpoint(t *testing.T) {
	tests := []isEndpointTypeTest{
		{endpointType: portaineree.DockerEnvironment, expected: true},
		{endpointType: portaineree.AgentOnDockerEnvironment, expected: true},
		{endpointType: portaineree.AzureEnvironment, expected: false},
		{endpointType: portaineree.EdgeAgentOnDockerEnvironment, expected: true},
		{endpointType: portaineree.KubernetesLocalEnvironment, expected: false},
		{endpointType: portaineree.AgentOnKubernetesEnvironment, expected: false},
		{endpointType: portaineree.EdgeAgentOnKubernetesEnvironment, expected: false},
	}

	for _, test := range tests {
		ans := IsDockerEndpoint(&portaineree.Endpoint{Type: test.endpointType})
		assert.Equal(t, test.expected, ans)
	}
}

func Test_IsKubernetesEndpoint(t *testing.T) {
	tests := []isEndpointTypeTest{
		{endpointType: portaineree.DockerEnvironment, expected: false},
		{endpointType: portaineree.AgentOnDockerEnvironment, expected: false},
		{endpointType: portaineree.AzureEnvironment, expected: false},
		{endpointType: portaineree.EdgeAgentOnDockerEnvironment, expected: false},
		{endpointType: portaineree.KubernetesLocalEnvironment, expected: true},
		{endpointType: portaineree.AgentOnKubernetesEnvironment, expected: true},
		{endpointType: portaineree.EdgeAgentOnKubernetesEnvironment, expected: true},
	}

	for _, test := range tests {
		ans := IsKubernetesEndpoint(&portaineree.Endpoint{Type: test.endpointType})
		assert.Equal(t, test.expected, ans)
	}
}

func Test_IsAgentEndpoint(t *testing.T) {
	tests := []isEndpointTypeTest{
		{endpointType: portaineree.DockerEnvironment, expected: false},
		{endpointType: portaineree.AgentOnDockerEnvironment, expected: true},
		{endpointType: portaineree.AzureEnvironment, expected: false},
		{endpointType: portaineree.EdgeAgentOnDockerEnvironment, expected: true},
		{endpointType: portaineree.KubernetesLocalEnvironment, expected: false},
		{endpointType: portaineree.AgentOnKubernetesEnvironment, expected: true},
		{endpointType: portaineree.EdgeAgentOnKubernetesEnvironment, expected: true},
	}

	for _, test := range tests {
		ans := IsAgentEndpoint(&portaineree.Endpoint{Type: test.endpointType})
		assert.Equal(t, test.expected, ans)
	}
}

func Test_FilterByExcludeIDs(t *testing.T) {
	tests := []struct {
		name            string
		inputArray      []portaineree.Endpoint
		inputExcludeIDs []portaineree.EndpointID
		asserts         func(*testing.T, []portaineree.Endpoint)
	}{
		{
			name: "filter endpoints",
			inputArray: []portaineree.Endpoint{
				{ID: portaineree.EndpointID(1)},
				{ID: portaineree.EndpointID(2)},
				{ID: portaineree.EndpointID(3)},
				{ID: portaineree.EndpointID(4)},
			},
			inputExcludeIDs: []portaineree.EndpointID{
				portaineree.EndpointID(2),
				portaineree.EndpointID(3),
			},
			asserts: func(t *testing.T, output []portaineree.Endpoint) {
				assert.Contains(t, output, portaineree.Endpoint{ID: portaineree.EndpointID(1)})
				assert.NotContains(t, output, portaineree.Endpoint{ID: portaineree.EndpointID(2)})
				assert.NotContains(t, output, portaineree.Endpoint{ID: portaineree.EndpointID(3)})
				assert.Contains(t, output, portaineree.Endpoint{ID: portaineree.EndpointID(4)})
			},
		},
		{
			name:       "empty input",
			inputArray: []portaineree.Endpoint{},
			inputExcludeIDs: []portaineree.EndpointID{
				portaineree.EndpointID(2),
			},
			asserts: func(t *testing.T, output []portaineree.Endpoint) {
				assert.Equal(t, 0, len(output))
			},
		},
		{
			name: "no filter",
			inputArray: []portaineree.Endpoint{
				{ID: portaineree.EndpointID(1)},
				{ID: portaineree.EndpointID(2)},
			},
			inputExcludeIDs: []portaineree.EndpointID{},
			asserts: func(t *testing.T, output []portaineree.Endpoint) {
				assert.Equal(t, 2, len(output))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FilterByExcludeIDs(tt.inputArray, tt.inputExcludeIDs)
			tt.asserts(t, output)
		})
	}
}

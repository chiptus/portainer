package endpointutils

import (
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
)

// IsLocalEndpoint returns true if this is a local environment(endpoint)
func IsLocalEndpoint(endpoint *portaineree.Endpoint) bool {
	return strings.HasPrefix(endpoint.URL, "unix://") || strings.HasPrefix(endpoint.URL, "npipe://") || endpoint.Type == 5
}

// IsKubernetesEndpoint returns true if this is a kubernetes environment(endpoint)
func IsKubernetesEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.KubernetesLocalEnvironment ||
		endpoint.Type == portaineree.AgentOnKubernetesEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment
}

// IsDockerEndpoint returns true if this is a docker environment(endpoint)
func IsDockerEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.DockerEnvironment ||
		endpoint.Type == portaineree.AgentOnDockerEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment
}

// IsEdgeEndpoint returns true if endpoint is an Edge Endpoint
func IsEdgeEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment
}

// IsAgentEndpoint returns true if this is an Agent endpoint
func IsAgentEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.AgentOnDockerEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment ||
		endpoint.Type == portaineree.AgentOnKubernetesEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment
}

// FilterByExcludeIDs receives an environment(endpoint) array and returns a filtered array using an excludeIds param
func FilterByExcludeIDs(endpoints []portaineree.Endpoint, excludeIds []portaineree.EndpointID) []portaineree.Endpoint {
	if len(excludeIds) == 0 {
		return endpoints
	}

	filteredEndpoints := make([]portaineree.Endpoint, 0)

	idsSet := make(map[portaineree.EndpointID]bool)
	for _, id := range excludeIds {
		idsSet[id] = true
	}

	for _, endpoint := range endpoints {
		if !idsSet[endpoint.ID] {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

// EndpointSet receives an environment(endpoint) array and returns a set
func EndpointSet(endpointIDs []portaineree.EndpointID) map[portaineree.EndpointID]bool {
	set := map[portaineree.EndpointID]bool{}

	for _, endpointID := range endpointIDs {
		set[endpointID] = true
	}

	return set
}

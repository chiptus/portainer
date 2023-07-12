package status

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

// NodesCount returns the total node number of all environments
func NodesCount(endpoints []portaineree.Endpoint) int {
	nodes := 0
	for _, env := range endpoints {
		nodes += countNodes(&env)
	}

	return nodes
}

func countNodes(endpoint *portaineree.Endpoint) int {

	// don't count edge endpoints that are not trusted
	if endpointutils.IsEdgeEndpoint(endpoint) && !endpoint.UserTrusted {
		return 0
	}

	// don't count endpoints that are not provisioned or in error
	if endpoint.CloudProvider != nil {
		if endpoint.Status == portaineree.EndpointStatusError {
			return 0
		}

		// Once provisioned we'll use the snapshot node count
		if endpoint.Status == portaineree.EndpointStatusProvisioning {
			return endpoint.CloudProvider.NodeCount
		}
	}

	// we count the number of nodes in the snapshot.  If there are no snapshots
	// we return a count of at least one node.
	if len(endpoint.Snapshots) == 1 {
		return max(endpoint.Snapshots[0].NodeCount, 1)
	}

	if len(endpoint.Kubernetes.Snapshots) == 1 {
		return max(endpoint.Kubernetes.Snapshots[0].NodeCount, 1)
	}

	if len(endpoint.Nomad.Snapshots) == 1 {
		return max(endpoint.Nomad.Snapshots[0].NodeCount, 1)
	}

	return 1
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

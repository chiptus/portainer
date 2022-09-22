package status

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// NodesCount returns the total node number of all environments
func NodesCount(snapshots []portaineree.Snapshot) int {
	nodes := 0
	for _, env := range snapshots {
		nodes += countNodes(&env)
	}

	return nodes
}

func countNodes(snapshot *portaineree.Snapshot) int {
	if snapshot.Docker != nil {
		return max(snapshot.Docker.NodeCount, 1)
	}

	if snapshot.Kubernetes != nil {
		return max(snapshot.Kubernetes.NodeCount, 1)
	}

	if snapshot.Nomad != nil {
		return max(snapshot.Nomad.NodeCount, 1)
	}

	return 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

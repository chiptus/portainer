package status

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// NodesCount returns the total node number of all environments
func NodesCount(environemnts []portaineree.Endpoint) int {
	nodes := 0
	for _, env := range environemnts {
		nodes += countNodes(&env)
	}

	return nodes
}

func countNodes(env *portaineree.Endpoint) int {
	switch env.Type {
	case portaineree.EdgeAgentOnDockerEnvironment, portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment:
		return countDockerNodes(env)
	case portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment:
		return countKubernetesNodes(env)
	case portaineree.EdgeAgentOnNomadEnvironment:
		return countNomadNodes(env)
	}
	return 1
}

func countDockerNodes(env *portaineree.Endpoint) int {
	snapshots := env.Snapshots
	if len(snapshots) == 0 {
		return 1
	}

	lastSnapshot := snapshots[len(snapshots)-1]
	return max(lastSnapshot.NodeCount, 1)
}

func countKubernetesNodes(env *portaineree.Endpoint) int {
	snapshots := env.Kubernetes.Snapshots
	if len(snapshots) == 0 {
		return 1
	}

	lastSnapshot := snapshots[len(snapshots)-1]
	return max(lastSnapshot.NodeCount, 1)
}

func countNomadNodes(env *portaineree.Endpoint) int {
	snapshots := env.Nomad.Snapshots
	if len(snapshots) == 0 {
		return 1
	}

	lastSnapshot := snapshots[len(snapshots)-1]
	return max(lastSnapshot.NodeCount, 1)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package status

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type nodesCountResponse struct {
	Nodes int `json:"nodes"`
}

// @id statusNodesCount
// @summary Retrieve the count of nodes
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags status
// @produce json
// @success 200 {object} nodesCountResponse "Success"
// @failure 500 "Server error"
// @router /status/nodes [get]
func (handler *Handler) statusNodesCount(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Failed to get environment list", err}
	}

	nodes := 0
	for _, endpoint := range endpoints {
		nodes += countNodes(&endpoint)
	}

	return response.JSON(w, &nodesCountResponse{Nodes: nodes})
}

func countNodes(endpoint *portaineree.Endpoint) int {
	switch endpoint.Type {
	case portaineree.EdgeAgentOnDockerEnvironment, portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment:
		return countDockerNodes(endpoint)
	case portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment:
		return countKubernetesNodes(endpoint)
	case portaineree.AzureEnvironment:
		return 1
	}
	return 1
}

func countDockerNodes(endpoint *portaineree.Endpoint) int {
	snapshots := endpoint.Snapshots
	if len(snapshots) == 0 {
		return 1
	}

	snapshot := snapshots[len(snapshots)-1]
	return max(snapshot.NodeCount, 1)
}

func countKubernetesNodes(endpoint *portaineree.Endpoint) int {
	snapshots := endpoint.Kubernetes.Snapshots
	if len(snapshots) == 0 {
		return 1
	}

	snapshot := snapshots[len(snapshots)-1]
	return max(snapshot.NodeCount, 1)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

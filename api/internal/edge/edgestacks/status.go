package edgestacks

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

// NewStatus returns a new status object for an Edge stack
func NewStatus(oldStatus map[portaineree.EndpointID]portainer.EdgeStackStatus, relatedEnvironmentIDs []portaineree.EndpointID) map[portaineree.EndpointID]portainer.EdgeStackStatus {
	status := map[portaineree.EndpointID]portainer.EdgeStackStatus{}

	for _, environmentID := range relatedEnvironmentIDs {

		newEnvStatus := portainer.EdgeStackStatus{
			Status:     []portainer.EdgeStackDeploymentStatus{},
			EndpointID: portainer.EndpointID(environmentID),
		}

		oldEnvStatus, ok := oldStatus[environmentID]
		if ok {
			newEnvStatus.DeploymentInfo = oldEnvStatus.DeploymentInfo
		}

		status[environmentID] = newEnvStatus
	}

	return status
}

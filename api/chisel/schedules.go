package chisel

import (
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
)

// AddEdgeJob register an EdgeJob inside the tunnel details associated to an environment(endpoint).
func (service *Service) AddEdgeJob(endpointID portaineree.EndpointID, edgeJob *portaineree.EdgeJob) {
	tunnel := service.GetTunnelDetails(endpointID)

	existingJobIndex := -1
	for idx, existingJob := range tunnel.Jobs {
		if existingJob.ID == edgeJob.ID {
			existingJobIndex = idx
			break
		}
	}

	if existingJobIndex == -1 {
		tunnel.Jobs = append(tunnel.Jobs, *edgeJob)
	} else {
		tunnel.Jobs[existingJobIndex] = *edgeJob
	}

	key := strconv.Itoa(int(endpointID))
	service.tunnelDetailsMap.Set(key, tunnel)
}

// RemoveEdgeJob will remove the specified Edge job from each tunnel it was registered with.
func (service *Service) RemoveEdgeJob(edgeJobID portaineree.EdgeJobID) {
	for item := range service.tunnelDetailsMap.IterBuffered() {
		tunnelDetails := item.Val.(*portaineree.TunnelDetails)

		updatedJobs := make([]portaineree.EdgeJob, 0)
		for _, edgeJob := range tunnelDetails.Jobs {
			if edgeJob.ID == edgeJobID {
				continue
			}
			updatedJobs = append(updatedJobs, edgeJob)
		}

		tunnelDetails.Jobs = updatedJobs
		service.tunnelDetailsMap.Set(item.Key, tunnelDetails)
	}
}

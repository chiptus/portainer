package chisel

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// AddEdgeJob register an EdgeJob inside the tunnel details associated to an environment(endpoint).
func (service *Service) AddEdgeJob(endpointID portaineree.EndpointID, edgeJob *portaineree.EdgeJob) {
	service.mu.Lock()
	tunnel := service.getTunnelDetails(endpointID)

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

	service.mu.Unlock()
}

// RemoveEdgeJob will remove the specified Edge job from each tunnel it was registered with.
func (service *Service) RemoveEdgeJob(edgeJobID portaineree.EdgeJobID) {
	service.mu.Lock()

	for _, tunnel := range service.tunnelDetailsMap {
		// Filter in-place
		n := 0
		for _, edgeJob := range tunnel.Jobs {
			if edgeJob.ID != edgeJobID {
				tunnel.Jobs[n] = edgeJob
				n++
			}
		}

		tunnel.Jobs = tunnel.Jobs[:n]
	}

	service.mu.Unlock()
}

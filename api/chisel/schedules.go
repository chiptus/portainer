package chisel

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
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

	cache.Del(endpointID)

	service.mu.Unlock()
}

// RemoveEdgeJob will remove the specified Edge job from each tunnel it was registered with.
func (service *Service) RemoveEdgeJob(edgeJobID portaineree.EdgeJobID) {
	service.mu.Lock()

	for endpointID, tunnel := range service.tunnelDetailsMap {
		n := 0
		for _, edgeJob := range tunnel.Jobs {
			if edgeJob.ID != edgeJobID {
				tunnel.Jobs[n] = edgeJob
				n++
			}
		}

		tunnel.Jobs = tunnel.Jobs[:n]

		cache.Del(endpointID)
	}

	service.mu.Unlock()
}

func (service *Service) RemoveEdgeJobFromEndpoint(endpointID portaineree.EndpointID, edgeJobID portaineree.EdgeJobID) {
	service.mu.Lock()
	tunnel := service.getTunnelDetails(endpointID)

	n := 0
	for _, edgeJob := range tunnel.Jobs {
		if edgeJob.ID != edgeJobID {
			tunnel.Jobs[n] = edgeJob
			n++
		}
	}

	tunnel.Jobs = tunnel.Jobs[:n]

	cache.Del(endpointID)

	service.mu.Unlock()
}

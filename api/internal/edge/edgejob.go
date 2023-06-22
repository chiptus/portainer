package edge

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

// LoadEdgeJobs registers all edge jobs inside corresponding environment(endpoint) tunnel
func LoadEdgeJobs(dataStore dataservices.DataStore, reverseTunnelService portaineree.ReverseTunnelService) error {
	edgeJobs, err := dataStore.EdgeJob().ReadAll()
	if err != nil {
		return err
	}

	for _, edgeJob := range edgeJobs {
		for endpointID := range edgeJob.Endpoints {
			endpoint, err := dataStore.Endpoint().Endpoint(endpointID)
			if err != nil {
				return err
			}

			reverseTunnelService.AddEdgeJob(endpoint, &edgeJob)
		}
	}

	return nil
}

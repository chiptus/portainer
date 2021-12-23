package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (m *Migrator) updateEndpointsToVersion10() error {
	legacyEndpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range legacyEndpoints {
		endpoint.Type = portaineree.DockerEnvironment
		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

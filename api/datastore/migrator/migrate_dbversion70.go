package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB70() error {
	log.Info().Msg("- add IngressAvailabilityPerNamespace field")
	if err := m.addIngressAvailabilityPerNamespaceFieldDB70(); err != nil {
		return err
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		// copy snapshots to new object
		log.Info().Msg("moving snapshots from endpoint to new object")
		snapshot := portaineree.Snapshot{EndpointID: endpoint.ID}

		if len(endpoint.Snapshots) > 0 {
			snapshot.Docker = &endpoint.Snapshots[len(endpoint.Snapshots)-1]
		}

		if len(endpoint.Kubernetes.Snapshots) > 0 {
			snapshot.Kubernetes = &endpoint.Kubernetes.Snapshots[len(endpoint.Kubernetes.Snapshots)-1]
		}

		if len(endpoint.Nomad.Snapshots) > 0 {
			snapshot.Nomad = &endpoint.Nomad.Snapshots[len(endpoint.Nomad.Snapshots)-1]
		}

		// save new object
		err = m.snapshotService.Create(&snapshot)
		if err != nil {
			return err
		}

		// set to nil old fields
		log.Info().Msg("deleting snapshot from endpoint")
		endpoint.Snapshots = []portainer.DockerSnapshot{}
		endpoint.Kubernetes.Snapshots = []portaineree.KubernetesSnapshot{}
		endpoint.Nomad.Snapshots = []portaineree.NomadSnapshot{}

		log.Info().Msg("- try to update the default value for Image notification toggle")
		if err = m.updateDefaultValueForImageNotificationToggleDb70(&endpoint); err != nil {
			return err
		}
		// update endpoint
		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateDefaultValueForImageNotificationToggleDb70(endpoint *portaineree.Endpoint) error {
	if m.Edition() == portaineree.PortainerCE {
		log.Info().Msg("skip image notification toggle migration for CE version")
		return nil
	}
	if m.Version() >= 50 && m.Version() < 70 {
		log.Info().Msgf("migrating from %d to 70, update the default value for image notification toggle", m.currentDBVersion)
		endpoint.EnableImageNotification = true
	}
	return nil
}

func (m *Migrator) addIngressAvailabilityPerNamespaceFieldDB70() error {
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		endpoint.Kubernetes.Configuration.IngressAvailabilityPerNamespace = true

		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

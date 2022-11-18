package migrator

import (
	"github.com/Masterminds/semver"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB70() error {
	log.Info().Msg("add IngressAvailabilityPerNamespace field")
	if err := m.updateIngressFieldsForEnvDB70(); err != nil {
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
		log.Info().Msgf("deleting snapshot from endpoint %d", endpoint.ID)
		endpoint.Snapshots = []portainer.DockerSnapshot{}
		endpoint.Kubernetes.Snapshots = []portaineree.KubernetesSnapshot{}
		endpoint.Nomad.Snapshots = []portaineree.NomadSnapshot{}

		log.Info().Msg("update default image notification toggle")
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
	if m.CurrentDBEdition() == portaineree.PortainerCE {
		log.Info().Msg("skip image notification toggle migration for CE version")
		return nil
	}

	constraint, err := semver.NewConstraint(">= 2.14, < 2.16")
	if err != nil {
		return err
	}

	if inRange, _ := constraint.Validate(m.CurrentSemanticDBVersion()); inRange {
		log.Info().Msgf("migrating from %s to 2.17, update the default value for image notification toggle", m.CurrentDBVersion())
		endpoint.EnableImageNotification = true
	}

	return nil
}

func (m *Migrator) updateIngressFieldsForEnvDB70() error {
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		endpoint.Kubernetes.Configuration.IngressAvailabilityPerNamespace = true
		endpoint.Kubernetes.Configuration.AllowNoneIngressClass = false
		endpoint.PostInitMigrations.MigrateIngresses = true

		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

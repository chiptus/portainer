package migrator

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	snapshotutils "github.com/portainer/portainer-ee/api/internal/snapshot"
	portainer "github.com/portainer/portainer/api"

	"github.com/docker/docker/api/types/volume"
	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB32() error {
	err := m.updateRegistriesToDB32()
	if err != nil {
		return err
	}

	err = m.updateDockerhubToDB32()
	if err != nil {
		return err
	}

	err = m.refreshRBACRoles()
	if err != nil {
		return err
	}

	err = m.refreshUserAuthorizations()
	if err != nil {
		return err
	}

	err = m.updateVolumeResourceControlToDB32()
	if err != nil {
		return err
	}

	err = m.updateAdminGroupSearchSettingsToDB32()
	if err != nil {
		return err
	}

	if err := m.migrateStackEntryPoint(); err != nil {
		return err
	}

	if err := m.kubeconfigExpiryToDB32(); err != nil {
		return err
	}

	if err := m.helmRepositoryURLToDB32(); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) updateRegistriesToDB32() error {
	log.Info().Msg("updating registries")

	registries, err := m.registryService.ReadAll()
	if err != nil {
		return err
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, registry := range registries {

		registry.RegistryAccesses = portainer.RegistryAccesses{}

		for _, endpoint := range endpoints {

			filteredUserAccessPolicies := portainer.UserAccessPolicies{}
			for userID, registryPolicy := range registry.UserAccessPolicies {
				if _, found := endpoint.UserAccessPolicies[userID]; found {
					filteredUserAccessPolicies[userID] = registryPolicy
				}
			}

			filteredTeamAccessPolicies := portainer.TeamAccessPolicies{}
			for teamID, registryPolicy := range registry.TeamAccessPolicies {
				if _, found := endpoint.TeamAccessPolicies[teamID]; found {
					filteredTeamAccessPolicies[teamID] = registryPolicy
				}
			}

			registry.RegistryAccesses[endpoint.ID] = portainer.RegistryAccessPolicies{
				UserAccessPolicies: filteredUserAccessPolicies,
				TeamAccessPolicies: filteredTeamAccessPolicies,
				Namespaces:         []string{},
			}
		}
		m.registryService.Update(registry.ID, &registry)
	}
	return nil
}

func (m *Migrator) updateDockerhubToDB32() error {
	log.Info().Msg("updating dockerhub")

	dockerhub, err := m.dockerhubService.DockerHub()
	if dataservices.IsErrObjectNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if !dockerhub.Authentication {
		return nil
	}

	registry := &portaineree.Registry{
		Type:             portaineree.DockerHubRegistry,
		Name:             "Dockerhub (authenticated - migrated)",
		URL:              "docker.io",
		Authentication:   true,
		Username:         dockerhub.Username,
		Password:         dockerhub.Password,
		RegistryAccesses: portainer.RegistryAccesses{},
	}

	// The following code will make this function idempotent.
	// i.e. if run again, it will not change the data.  It will ensure that
	// we only have one migrated registry entry. Duplicates will be removed
	// if they exist and which has been happening due to earlier migration bugs
	migrated := false
	registries, _ := m.registryService.ReadAll()
	for _, r := range registries {
		if r.Type == registry.Type &&
			r.Name == registry.Name &&
			r.URL == registry.URL &&
			r.Authentication == registry.Authentication {

			if !migrated {
				// keep this one entry
				migrated = true
			} else {
				// delete subsequent duplicates
				m.registryService.Delete(portainer.RegistryID(r.ID))
			}
		}
	}

	if migrated {
		return nil
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {

		if endpoint.Type != portaineree.KubernetesLocalEnvironment &&
			endpoint.Type != portaineree.AgentOnKubernetesEnvironment &&
			endpoint.Type != portaineree.EdgeAgentOnKubernetesEnvironment {

			userAccessPolicies := portainer.UserAccessPolicies{}
			for userID := range endpoint.UserAccessPolicies {
				if _, found := endpoint.UserAccessPolicies[userID]; found {
					userAccessPolicies[userID] = portainer.AccessPolicy{
						RoleID: 0,
					}
				}
			}

			teamAccessPolicies := portainer.TeamAccessPolicies{}
			for teamID := range endpoint.TeamAccessPolicies {
				if _, found := endpoint.TeamAccessPolicies[teamID]; found {
					teamAccessPolicies[teamID] = portainer.AccessPolicy{
						RoleID: 0,
					}
				}
			}

			registry.RegistryAccesses[endpoint.ID] = portainer.RegistryAccessPolicies{
				UserAccessPolicies: userAccessPolicies,
				TeamAccessPolicies: teamAccessPolicies,
				Namespaces:         []string{},
			}
		}
	}

	return m.registryService.Create(registry)
}

func (m *Migrator) migrateStackEntryPoint() error {
	log.Info().Msg("updating stack entry points")

	stacks, err := m.stackService.ReadAll()
	if err != nil {
		return err
	}
	for i := range stacks {
		stack := &stacks[i]
		if stack.GitConfig == nil {
			continue
		}
		stack.GitConfig.ConfigFilePath = stack.EntryPoint
		if err := m.stackService.Update(stack.ID, stack); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) updateVolumeResourceControlToDB32() error {
	log.Info().Msg("updating resource controls")

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return fmt.Errorf("failed fetching environments: %w", err)
	}

	resourceControls, err := m.resourceControlService.ReadAll()
	if err != nil {
		return fmt.Errorf("failed fetching resource controls: %w", err)
	}

	toUpdate := map[portainer.ResourceControlID]string{}
	volumeResourceControls := map[string]*portainer.ResourceControl{}

	for i := range resourceControls {
		resourceControl := resourceControls[i]
		if resourceControl.Type == portaineree.VolumeResourceControl {
			volumeResourceControls[resourceControl.ResourceID] = &resourceControl
		}
	}

	for _, endpoint := range endpoints {
		if !endpointutils.IsDockerEndpoint(&endpoint) {
			continue
		}

		totalSnapshots := len(endpoint.Snapshots)
		if totalSnapshots == 0 {
			log.Debug().Msg("no snapshot found")
			continue
		}

		snapshot := endpoint.Snapshots[totalSnapshots-1]

		endpointDockerID, err := snapshotutils.FetchDockerID(snapshot)
		if err != nil {
			log.Warn().Err(err).Msg("failed fetching environment docker id")
			continue
		}

		volumesData := snapshot.SnapshotRaw.Volumes
		if volumesData.Volumes == nil {
			log.Debug().Msg("no volume data found")
			continue
		}

		findResourcesToUpdateToDB32(endpointDockerID, volumesData, toUpdate, volumeResourceControls)

	}

	for _, resourceControl := range volumeResourceControls {
		if newResourceID, ok := toUpdate[resourceControl.ID]; ok {
			resourceControl.ResourceID = newResourceID

			err := m.resourceControlService.Update(resourceControl.ID, resourceControl)
			if err != nil {
				return fmt.Errorf("failed updating resource control %d: %w", resourceControl.ID, err)
			}
		} else {
			err := m.resourceControlService.Delete(resourceControl.ID)
			if err != nil {
				return fmt.Errorf("failed deleting resource control %d: %w", resourceControl.ID, err)
			}

			log.Debug().Str("resource_id", resourceControl.ResourceID).Msg("legacy resource control has been deleted")
		}
	}

	return nil
}

func findResourcesToUpdateToDB32(dockerID string, volumesData volume.ListResponse, toUpdate map[portainer.ResourceControlID]string, volumeResourceControls map[string]*portainer.ResourceControl) {
	volumes := volumesData.Volumes
	for _, volume := range volumes {
		volumeName := volume.Name
		createTime := volume.CreatedAt

		oldResourceID := fmt.Sprintf("%s%s", volumeName, createTime)
		resourceControl, ok := volumeResourceControls[oldResourceID]

		if ok {
			toUpdate[resourceControl.ID] = fmt.Sprintf("%s_%s", volumeName, dockerID)
		}
	}
}

func (m *Migrator) updateAdminGroupSearchSettingsToDB32() error {
	log.Info().Msg("updating admin group search settings")

	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	if legacySettings.LDAPSettings.AdminGroupSearchSettings == nil {
		legacySettings.LDAPSettings.AdminGroupSearchSettings = []portainer.LDAPGroupSearchSettings{
			{},
		}
	}
	return m.settingsService.UpdateSettings(legacySettings)
}

func (m *Migrator) kubeconfigExpiryToDB32() error {
	log.Info().Msg("updating kubeconfig expiry")

	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	settings.KubeconfigExpiry = portaineree.DefaultKubeconfigExpiry
	return m.settingsService.UpdateSettings(settings)
}

func (m *Migrator) helmRepositoryURLToDB32() error {
	log.Info().Msg("setting default helm repository URL")

	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	settings.HelmRepositoryURL = portaineree.DefaultHelmRepositoryURL
	return m.settingsService.UpdateSettings(settings)
}

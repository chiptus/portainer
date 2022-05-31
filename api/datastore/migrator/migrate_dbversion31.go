package migrator

import (
	"fmt"
	"log"

	"github.com/docker/docker/api/types/volume"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	snapshotutils "github.com/portainer/portainer-ee/api/internal/snapshot"
	"github.com/portainer/portainer/api/dataservices/errors"
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
	migrateLog.Info("- updating registries")
	registries, err := m.registryService.Registries()
	if err != nil {
		return err
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, registry := range registries {

		registry.RegistryAccesses = portaineree.RegistryAccesses{}

		for _, endpoint := range endpoints {

			filteredUserAccessPolicies := portaineree.UserAccessPolicies{}
			for userID, registryPolicy := range registry.UserAccessPolicies {
				if _, found := endpoint.UserAccessPolicies[userID]; found {
					filteredUserAccessPolicies[userID] = registryPolicy
				}
			}

			filteredTeamAccessPolicies := portaineree.TeamAccessPolicies{}
			for teamID, registryPolicy := range registry.TeamAccessPolicies {
				if _, found := endpoint.TeamAccessPolicies[teamID]; found {
					filteredTeamAccessPolicies[teamID] = registryPolicy
				}
			}

			registry.RegistryAccesses[endpoint.ID] = portaineree.RegistryAccessPolicies{
				UserAccessPolicies: filteredUserAccessPolicies,
				TeamAccessPolicies: filteredTeamAccessPolicies,
				Namespaces:         []string{},
			}
		}
		m.registryService.UpdateRegistry(registry.ID, &registry)
	}
	return nil
}

func (m *Migrator) updateDockerhubToDB32() error {
	migrateLog.Info("- updating dockerhub")
	dockerhub, err := m.dockerhubService.DockerHub()
	if err == errors.ErrObjectNotFound {
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
		RegistryAccesses: portaineree.RegistryAccesses{},
	}

	// The following code will make this function idempotent.
	// i.e. if run again, it will not change the data.  It will ensure that
	// we only have one migrated registry entry. Duplicates will be removed
	// if they exist and which has been happening due to earlier migration bugs
	migrated := false
	registries, _ := m.registryService.Registries()
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
				m.registryService.DeleteRegistry(portaineree.RegistryID(r.ID))
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

			userAccessPolicies := portaineree.UserAccessPolicies{}
			for userID := range endpoint.UserAccessPolicies {
				if _, found := endpoint.UserAccessPolicies[userID]; found {
					userAccessPolicies[userID] = portaineree.AccessPolicy{
						RoleID: 0,
					}
				}
			}

			teamAccessPolicies := portaineree.TeamAccessPolicies{}
			for teamID := range endpoint.TeamAccessPolicies {
				if _, found := endpoint.TeamAccessPolicies[teamID]; found {
					teamAccessPolicies[teamID] = portaineree.AccessPolicy{
						RoleID: 0,
					}
				}
			}

			registry.RegistryAccesses[endpoint.ID] = portaineree.RegistryAccessPolicies{
				UserAccessPolicies: userAccessPolicies,
				TeamAccessPolicies: teamAccessPolicies,
				Namespaces:         []string{},
			}
		}
	}

	return m.registryService.Create(registry)
}

func (m *Migrator) migrateStackEntryPoint() error {
	migrateLog.Info("- updating stack entry points")
	stacks, err := m.stackService.Stacks()
	if err != nil {
		return err
	}
	for i := range stacks {
		stack := &stacks[i]
		if stack.GitConfig == nil {
			continue
		}
		stack.GitConfig.ConfigFilePath = stack.EntryPoint
		if err := m.stackService.UpdateStack(stack.ID, stack); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) updateVolumeResourceControlToDB32() error {
	migrateLog.Info("- updating resource controls")
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return fmt.Errorf("failed fetching environments: %w", err)
	}

	resourceControls, err := m.resourceControlService.ResourceControls()
	if err != nil {
		return fmt.Errorf("failed fetching resource controls: %w", err)
	}

	toUpdate := map[portaineree.ResourceControlID]string{}
	volumeResourceControls := map[string]*portaineree.ResourceControl{}

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
			log.Println("[DEBUG] [volume migration] [message: no snapshot found]")
			continue
		}

		snapshot := endpoint.Snapshots[totalSnapshots-1]

		endpointDockerID, err := snapshotutils.FetchDockerID(snapshot)
		if err != nil {
			log.Printf("[WARN] [bolt,migrator,v31] [message: failed fetching environment docker id] [err: %s]", err)
			continue
		}

		volumesData := snapshot.SnapshotRaw.Volumes
		if volumesData.Volumes == nil {
			log.Println("[DEBUG] [volume migration] [message: no volume data found]")
			continue
		}

		findResourcesToUpdateToDB32(endpointDockerID, volumesData, toUpdate, volumeResourceControls)

	}

	for _, resourceControl := range volumeResourceControls {
		if newResourceID, ok := toUpdate[resourceControl.ID]; ok {
			resourceControl.ResourceID = newResourceID
			err := m.resourceControlService.UpdateResourceControl(resourceControl.ID, resourceControl)
			if err != nil {
				return fmt.Errorf("failed updating resource control %d: %w", resourceControl.ID, err)
			}

		} else {
			err := m.resourceControlService.DeleteResourceControl(resourceControl.ID)
			if err != nil {
				return fmt.Errorf("failed deleting resource control %d: %w", resourceControl.ID, err)
			}
			log.Printf("[DEBUG] [volume migration] [message: legacy resource control(%s) has been deleted]", resourceControl.ResourceID)
		}
	}

	return nil
}

func findResourcesToUpdateToDB32(dockerID string, volumesData volume.VolumeListOKBody, toUpdate map[portaineree.ResourceControlID]string, volumeResourceControls map[string]*portaineree.ResourceControl) {
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
	migrateLog.Info("- updating admin group search settings")
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	if legacySettings.LDAPSettings.AdminGroupSearchSettings == nil {
		legacySettings.LDAPSettings.AdminGroupSearchSettings = []portaineree.LDAPGroupSearchSettings{
			{},
		}
	}
	return m.settingsService.UpdateSettings(legacySettings)
}

func (m *Migrator) kubeconfigExpiryToDB32() error {
	migrateLog.Info("- updating kubeconfig expiry")
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	settings.KubeconfigExpiry = portaineree.DefaultKubeconfigExpiry
	return m.settingsService.UpdateSettings(settings)
}

func (m *Migrator) helmRepositoryURLToDB32() error {
	migrateLog.Info("- setting default helm repository URL")
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	settings.HelmRepositoryURL = portaineree.DefaultHelmRepositoryURL
	return m.settingsService.UpdateSettings(settings)
}

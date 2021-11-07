package migrator

import (
	"fmt"
	"log"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/internal/endpointutils"
	snapshotutils "github.com/portainer/portainer/api/internal/snapshot"
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

	if err := migrateStackEntryPoint(m.stackService); err != nil {
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
	registries, err := m.registryService.Registries()
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
		m.registryService.UpdateRegistry(registry.ID, &registry)
	}
	return nil
}

func (m *Migrator) updateDockerhubToDB32() error {
	dockerhub, err := m.dockerhubService.DockerHub()
	if err == errors.ErrObjectNotFound {
		return nil
	} else if err != nil {
		return err
	}

	if !dockerhub.Authentication {
		return nil
	}

	registry := &portainer.Registry{
		Type:             portainer.DockerHubRegistry,
		Name:             "Dockerhub (authenticated - migrated)",
		URL:              "docker.io",
		Authentication:   true,
		Username:         dockerhub.Username,
		Password:         dockerhub.Password,
		RegistryAccesses: portainer.RegistryAccesses{},
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {

		if endpoint.Type != portainer.KubernetesLocalEnvironment &&
			endpoint.Type != portainer.AgentOnKubernetesEnvironment &&
			endpoint.Type != portainer.EdgeAgentOnKubernetesEnvironment {

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

	return m.registryService.CreateRegistry(registry)
}

func migrateStackEntryPoint(stackService portainer.StackService) error {
	stacks, err := stackService.Stacks()
	if err != nil {
		return err
	}
	for i := range stacks {
		stack := &stacks[i]
		if stack.GitConfig == nil {
			continue
		}
		stack.GitConfig.ConfigFilePath = stack.EntryPoint
		if err := stackService.UpdateStack(stack.ID, stack); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) updateVolumeResourceControlToDB32() error {
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return fmt.Errorf("failed fetching environments: %w", err)
	}

	resourceControls, err := m.resourceControlService.ResourceControls()
	if err != nil {
		return fmt.Errorf("failed fetching resource controls: %w", err)
	}

	toUpdate := map[portainer.ResourceControlID]string{}
	volumeResourceControls := map[string]*portainer.ResourceControl{}

	for i := range resourceControls {
		resourceControl := resourceControls[i]
		if resourceControl.Type == portainer.VolumeResourceControl {
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

		if volumesData, done := snapshot.SnapshotRaw.Volumes.(map[string]interface{}); done {
			if volumesData["Volumes"] == nil {
				log.Println("[DEBUG] [volume migration] [message: no volume data found]")
				continue
			}

			findResourcesToUpdateToDB32(endpointDockerID, volumesData, toUpdate, volumeResourceControls)
		}
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

func findResourcesToUpdateToDB32(dockerID string, volumesData map[string]interface{}, toUpdate map[portainer.ResourceControlID]string, volumeResourceControls map[string]*portainer.ResourceControl) {
	volumes := volumesData["Volumes"].([]interface{})
	for _, volumeMeta := range volumes {
		volume := volumeMeta.(map[string]interface{})
		volumeName, nameExist := volume["Name"].(string)
		if !nameExist {
			continue
		}
		createTime, createTimeExist := volume["CreatedAt"].(string)
		if !createTimeExist {
			continue
		}

		oldResourceID := fmt.Sprintf("%s%s", volumeName, createTime)
		resourceControl, ok := volumeResourceControls[oldResourceID]

		if ok {
			toUpdate[resourceControl.ID] = fmt.Sprintf("%s_%s", volumeName, dockerID)
		}
	}
}

func (m *Migrator) updateAdminGroupSearchSettingsToDB32() error {
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
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	settings.KubeconfigExpiry = portainer.DefaultKubeconfigExpiry
	return m.settingsService.UpdateSettings(settings)
}

func (m *Migrator) helmRepositoryURLToDB32() error {
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	settings.HelmRepositoryURL = portainer.DefaultHelmRepositoryURL
	return m.settingsService.UpdateSettings(settings)
}

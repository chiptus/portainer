package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

func (m *Migrator) updateUsersToDBVersion18() error {
	migrateLog.Info("- updating users")

	legacyUsers, err := m.userService.Users()
	if err != nil {
		return err
	}

	for _, user := range legacyUsers {
		user.PortainerAuthorizations = map[portaineree.Authorization]bool{
			portaineree.OperationPortainerDockerHubInspect:        true,
			portaineree.OperationPortainerEndpointGroupList:       true,
			portaineree.OperationPortainerEndpointList:            true,
			portaineree.OperationPortainerEndpointInspect:         true,
			portaineree.OperationPortainerEndpointExtensionAdd:    true,
			portaineree.OperationPortainerEndpointExtensionRemove: true,
			portaineree.OperationPortainerExtensionList:           true,
			portaineree.OperationPortainerMOTD:                    true,
			portaineree.OperationPortainerRegistryList:            true,
			portaineree.OperationPortainerRegistryInspect:         true,
			portaineree.OperationPortainerTeamList:                true,
			portaineree.OperationPortainerTemplateList:            true,
			portaineree.OperationPortainerTemplateInspect:         true,
			portaineree.OperationPortainerUserList:                true,
			portaineree.OperationPortainerUserMemberships:         true,
		}

		err = m.userService.UpdateUser(user.ID, &user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateEndpointsToDBVersion18() error {
	migrateLog.Info("- updating endpoints")
	legacyEndpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range legacyEndpoints {
		endpoint.UserAccessPolicies = make(portaineree.UserAccessPolicies)
		for _, userID := range endpoint.AuthorizedUsers {
			endpoint.UserAccessPolicies[userID] = portaineree.AccessPolicy{
				RoleID: 4,
			}
		}

		endpoint.TeamAccessPolicies = make(portaineree.TeamAccessPolicies)
		for _, teamID := range endpoint.AuthorizedTeams {
			endpoint.TeamAccessPolicies[teamID] = portaineree.AccessPolicy{
				RoleID: 4,
			}
		}

		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateEndpointGroupsToDBVersion18() error {
	migrateLog.Info("- updating endpoint groups")
	legacyEndpointGroups, err := m.endpointGroupService.EndpointGroups()
	if err != nil {
		return err
	}

	for _, endpointGroup := range legacyEndpointGroups {
		endpointGroup.UserAccessPolicies = make(portaineree.UserAccessPolicies)
		for _, userID := range endpointGroup.AuthorizedUsers {
			endpointGroup.UserAccessPolicies[userID] = portaineree.AccessPolicy{
				RoleID: 4,
			}
		}

		endpointGroup.TeamAccessPolicies = make(portaineree.TeamAccessPolicies)
		for _, teamID := range endpointGroup.AuthorizedTeams {
			endpointGroup.TeamAccessPolicies[teamID] = portaineree.AccessPolicy{
				RoleID: 4,
			}
		}

		err = m.endpointGroupService.UpdateEndpointGroup(endpointGroup.ID, &endpointGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateRegistriesToDBVersion18() error {
	migrateLog.Info("- updating registries")
	legacyRegistries, err := m.registryService.Registries()
	if err != nil {
		return err
	}

	for _, registry := range legacyRegistries {
		registry.UserAccessPolicies = make(portaineree.UserAccessPolicies)
		for _, userID := range registry.AuthorizedUsers {
			registry.UserAccessPolicies[userID] = portaineree.AccessPolicy{}
		}

		registry.TeamAccessPolicies = make(portaineree.TeamAccessPolicies)
		for _, teamID := range registry.AuthorizedTeams {
			registry.TeamAccessPolicies[teamID] = portaineree.AccessPolicy{}
		}

		err = m.registryService.UpdateRegistry(registry.ID, &registry)
		if err != nil {
			return err
		}
	}

	return nil
}

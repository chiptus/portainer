package migrator

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

// UpgradeToEE will migrate the db from latest ce version to latest ee version
// Latest version is v25 on 06/11/2020
func (m *Migrator) UpgradeToEE() error {

	migrateLog.Info(fmt.Sprintf("Migrating CE database version %d to EE database version %d.", m.Version(), portaineree.DBVersion))

	migrateLog.Info("Updating LDAP settings to EE")
	err := m.updateSettingsToEE()
	if err != nil {
		return err
	}

	migrateLog.Info("Updating user roles to EE")
	err = m.updateUserRolesToEE()
	if err != nil {
		return err
	}
	migrateLog.Info("Updating role authorizations to EE")
	err = m.updateRoleAuthorizationsToEE()
	if err != nil {
		return err
	}
	migrateLog.Info("Updating user authorizations")
	err = m.authorizationService.UpdateUsersAuthorizations()
	if err != nil {
		return err
	}

	migrateLog.Info(fmt.Sprintf("Setting db version to %d", portaineree.DBVersionEE))
	err = m.versionService.StoreDBVersion(portaineree.DBVersionEE)
	if err != nil {
		return err
	}

	migrateLog.Info(fmt.Sprintf("Setting edition to %s", portaineree.PortainerEE.GetEditionLabel()))
	err = m.versionService.StoreEdition(portaineree.PortainerEE)
	if err != nil {
		return err
	}

	m.currentDBVersion = portaineree.DBVersionEE
	m.currentEdition = portaineree.PortainerEE

	return nil
}

func (m *Migrator) updateSettingsToEE() error {
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.LDAPSettings.URLs = []string{}
	url := legacySettings.LDAPSettings.URL
	if url != "" {
		legacySettings.LDAPSettings.URLs = append(legacySettings.LDAPSettings.URLs, url)
	}

	legacySettings.LDAPSettings.ServerType = portaineree.LDAPServerCustom

	return m.settingsService.UpdateSettings(legacySettings)
}

// Updating role authorizations because of the new policies in Kube RBAC
func (m *Migrator) updateRoleAuthorizationsToEE() error {
	migrateLog.Debug("Retriving settings")

	migrateLog.Debug("Updating Environment Admin Role")
	endpointAdministratorRole, err := m.roleService.Role(portaineree.RoleID(1))
	if err != nil {
		return err
	}
	endpointAdministratorRole.Priority = 1
	endpointAdministratorRole.Authorizations = authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole()

	err = m.roleService.UpdateRole(endpointAdministratorRole.ID, endpointAdministratorRole)

	migrateLog.Debug("Updating Help Desk Role")
	helpDeskRole, err := m.roleService.Role(portaineree.RoleID(2))
	if err != nil {
		return err
	}
	helpDeskRole.Priority = 2
	helpDeskRole.Authorizations = authorization.DefaultEndpointAuthorizationsForHelpDeskRole()

	err = m.roleService.UpdateRole(helpDeskRole.ID, helpDeskRole)

	migrateLog.Debug("Updating Standard User Role")
	standardUserRole, err := m.roleService.Role(portaineree.RoleID(3))
	if err != nil {
		return err
	}
	standardUserRole.Priority = 3
	standardUserRole.Authorizations = authorization.DefaultEndpointAuthorizationsForStandardUserRole()

	err = m.roleService.UpdateRole(standardUserRole.ID, standardUserRole)

	migrateLog.Debug("Updating Read Only User Role")
	readOnlyUserRole, err := m.roleService.Role(portaineree.RoleID(4))
	if err != nil {
		return err
	}
	readOnlyUserRole.Priority = 4
	readOnlyUserRole.Authorizations = authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole()

	err = m.roleService.UpdateRole(readOnlyUserRole.ID, readOnlyUserRole)
	if err != nil {
		return err
	}

	return nil
}

// If RBAC extension wasn't installed before, update all users in environments(endpoints) and
// environment(endpoint) groups to have read only access.
func (m *Migrator) updateUserRolesToEE() error {
	err := m.updateUserAuthorizationToEE()
	if err != nil {
		return err
	}

	migrateLog.Debug("Retriving extension info")
	extensions, err := m.extensionService.Extensions()
	for _, extension := range extensions {
		if extension.ID == 3 && extension.Enabled {
			migrateLog.Info("RBAC extensions were enabled before; Skip updating User Roles")
			return nil
		}
	}

	migrateLog.Debug("Retriving environment groups")
	endpointGroups, err := m.endpointGroupService.EndpointGroups()
	if err != nil {
		return err
	}

	for _, endpointGroup := range endpointGroups {
		migrateLog.Debug(fmt.Sprintf("Updating user policies for environment group %v", endpointGroup.ID))
		for key := range endpointGroup.UserAccessPolicies {
			updateUserAccessPolicyToReadOnlyRole(endpointGroup.UserAccessPolicies, key)
		}

		for key := range endpointGroup.TeamAccessPolicies {
			updateTeamAccessPolicyToReadOnlyRole(endpointGroup.TeamAccessPolicies, key)
		}

		err := m.endpointGroupService.UpdateEndpointGroup(endpointGroup.ID, &endpointGroup)
		if err != nil {
			return err
		}
	}

	migrateLog.Debug("Retriving environments")
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		migrateLog.Debug(fmt.Sprintf("Updating user policies for environment %v", endpoint.ID))
		for key := range endpoint.UserAccessPolicies {
			updateUserAccessPolicyToReadOnlyRole(endpoint.UserAccessPolicies, key)
		}

		for key := range endpoint.TeamAccessPolicies {
			updateTeamAccessPolicyToReadOnlyRole(endpoint.TeamAccessPolicies, key)
		}

		err := m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateUserAuthorizationToEE() error {
	legacyUsers, err := m.userService.Users()
	if err != nil {
		return err
	}

	for _, user := range legacyUsers {
		user.PortainerAuthorizations = authorization.DefaultPortainerAuthorizations()

		err = m.userService.UpdateUser(user.ID, &user)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateUserAccessPolicyToNoRole(policies portaineree.UserAccessPolicies, key portaineree.UserID) {
	tmp := policies[key]
	tmp.RoleID = 0
	policies[key] = tmp
}

func updateTeamAccessPolicyToNoRole(policies portaineree.TeamAccessPolicies, key portaineree.TeamID) {
	tmp := policies[key]
	tmp.RoleID = 0
	policies[key] = tmp
}

func updateUserAccessPolicyToReadOnlyRole(policies portaineree.UserAccessPolicies, key portaineree.UserID) {
	tmp := policies[key]
	tmp.RoleID = 4
	policies[key] = tmp
}

func updateTeamAccessPolicyToReadOnlyRole(policies portaineree.TeamAccessPolicies, key portaineree.TeamID) {
	tmp := policies[key]
	tmp.RoleID = 4
	policies[key] = tmp
}

package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore/validate"
	"github.com/portainer/portainer-ee/api/internal/authorization"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// UpgradeToEE will migrate the db from latest ce version to latest ee version
// Latest version is v25 on 06/11/2020
func (m *Migrator) UpgradeToEE() error {
	log.Info().Int("ce_version", m.Version()).Int("ee_version", portaineree.DBVersion).Msg("migrating CE database EE")

	log.Info().Msg("upgrading LDAP settings to EE")
	err := m.updateLdapSettingsToEE()
	if err != nil {
		return err
	}

	log.Info().Msg("upgrading user roles to EE")
	err = m.updateUserRolesToEE()
	if err != nil {
		return err
	}

	log.Info().Msg("upgrading role authorizations to EE")
	err = m.roleService.CreateOrUpdatePredefinedRoles()
	if err != nil {
		return err
	}

	log.Info().Msg("upgrading user authorizations")
	err = m.authorizationService.UpdateUsersAuthorizations()
	if err != nil {
		return err
	}

	// Validate
	roles, err := m.roleService.Roles()
	if err != nil {
		return errors.Wrap(err, "while getting roles")
	}

	err = validate.ValidatePredefinedRoles(roles)
	if err != nil {
		return errors.Wrap(err, "while validating roles")
	}

	log.Info().Int("version", portaineree.DBVersionEE).Msg("setting DB version")

	err = m.versionService.StoreDBVersion(portaineree.DBVersionEE)
	if err != nil {
		return err
	}

	log.Info().Str("edition", portaineree.PortainerEE.GetEditionLabel()).Msg("setting edition")

	err = m.versionService.StoreEdition(portaineree.PortainerEE)
	if err != nil {
		return err
	}

	m.currentDBVersion = portaineree.DBVersionEE
	m.currentEdition = portaineree.PortainerEE

	return nil
}

func (m *Migrator) updateLdapSettingsToEE() error {
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

	// The front end requires a slice with a single empty element to allow configuration
	legacySettings.LDAPSettings.AdminGroupSearchSettings = []portaineree.LDAPGroupSearchSettings{
		{},
	}

	err = m.settingsService.UpdateSettings(legacySettings)
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

	log.Debug().Msg("retrieving extension info")
	extensions, err := m.extensionService.Extensions()
	for _, extension := range extensions {
		if extension.ID == 3 && extension.Enabled {
			log.Info().Msg("RBAC extensions were enabled before; Skip updating User Roles")
			return nil
		}
	}

	log.Debug().Msg("retrieving environment groups")
	endpointGroups, err := m.endpointGroupService.EndpointGroups()
	if err != nil {
		return err
	}

	for _, endpointGroup := range endpointGroups {
		log.Debug().Int("group", int(endpointGroup.ID)).Msg("updating user policies for environment group")

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

	log.Debug().Msg("Retrieving environments")
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		log.Debug().Int("endpoint_id", int(endpoint.ID)).Msg("updating user policies for environment")

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

package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore/validate"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// UpgradeToEE will migrate the db from latest ce version to latest ee version
// Latest version is v25 on 06/11/2020
func (m *Migrator) UpgradeToEE() error {
	log.Info().Msgf("upgrading database from CE to EE")

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
	roles, err := m.roleService.ReadAll()
	if err != nil {
		return errors.Wrap(err, "while getting roles")
	}

	err = validate.ValidatePredefinedRoles(roles)
	if err != nil {
		return errors.Wrap(err, "while validating roles")
	}

	log.Info().Str("edition", portaineree.PortainerEE.GetEditionLabel()).Msg("set edition")

	v, err := m.versionService.Version()
	if err != nil {
		return err
	}

	v.Edition = int(portaineree.PortainerEE)

	err = m.versionService.UpdateVersion(v)
	if err != nil {
		return err
	}

	log.Info().Msg("setting default image up to date to true")
	if err := m.defaultImageUpToDateOn(); err != nil {
		return err
	}

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
	legacySettings.LDAPSettings.AdminGroupSearchSettings = []portainer.LDAPGroupSearchSettings{
		{},
	}

	return m.settingsService.UpdateSettings(legacySettings)
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
	endpointGroups, err := m.endpointGroupService.ReadAll()
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

		err := m.endpointGroupService.Update(endpointGroup.ID, &endpointGroup)
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
	legacyUsers, err := m.userService.ReadAll()
	if err != nil {
		return err
	}

	for _, user := range legacyUsers {
		user.PortainerAuthorizations = authorization.DefaultPortainerAuthorizations()

		err = m.userService.Update(user.ID, &user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) defaultImageUpToDateOn() error {
	// get all environments
	environments, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, environment := range environments {
		// update environment to enable image notification
		if environment.Type == portaineree.DockerEnvironment || environment.Type == portaineree.EdgeAgentOnDockerEnvironment {
			environment.EnableImageNotification = true
			if err := m.endpointService.UpdateEndpoint(environment.ID, &environment); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateUserAccessPolicyToReadOnlyRole(policies portainer.UserAccessPolicies, key portainer.UserID) {
	tmp := policies[key]
	tmp.RoleID = 4
	policies[key] = tmp
}

func updateTeamAccessPolicyToReadOnlyRole(policies portainer.TeamAccessPolicies, key portainer.TeamID) {
	tmp := policies[key]
	tmp.RoleID = 4
	policies[key] = tmp
}

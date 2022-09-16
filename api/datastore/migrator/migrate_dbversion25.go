package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) updateEndpointSettingsToDB26() error {
	log.Info().Msg("updating endpoint settings")

	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for i := range endpoints {
		endpoint := endpoints[i]

		securitySettings := portaineree.EndpointSecuritySettings{}

		if endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment ||
			endpoint.Type == portaineree.AgentOnDockerEnvironment ||
			endpoint.Type == portaineree.DockerEnvironment {

			securitySettings = portaineree.EndpointSecuritySettings{
				AllowBindMountsForRegularUsers:            settings.AllowBindMountsForRegularUsers,
				AllowContainerCapabilitiesForRegularUsers: settings.AllowContainerCapabilitiesForRegularUsers,
				AllowDeviceMappingForRegularUsers:         settings.AllowDeviceMappingForRegularUsers,
				AllowHostNamespaceForRegularUsers:         settings.AllowHostNamespaceForRegularUsers,
				AllowPrivilegedModeForRegularUsers:        settings.AllowPrivilegedModeForRegularUsers,
				AllowStackManagementForRegularUsers:       settings.AllowStackManagementForRegularUsers,
			}

			if endpoint.Type == portaineree.AgentOnDockerEnvironment || endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment {
				securitySettings.AllowVolumeBrowserForRegularUsers = settings.AllowVolumeBrowserForRegularUsers
				securitySettings.EnableHostManagementFeatures = settings.EnableHostManagementFeatures
			}
		}

		endpoint.SecuritySettings = securitySettings

		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateRbacRolesToDB26() error {
	log.Info().Msg("updating RBAC roles")

	return m.refreshRBACRoles()
}

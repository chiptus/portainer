package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB60() error {
	log.Info().Msg("add gpu input field")

	if err := m.addGpuInputFieldDB60(); err != nil {
		return err
	}

	log.Info().Msg("updating ldap settings")

	return m.updateLdapSettingsEE()
}

func (m *Migrator) addGpuInputFieldDB60() error {
	log.Info().Msg("add gpu input field")

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		if endpoint.Gpus == nil {
			endpoint.Gpus = []portainer.Pair{}
			err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Fixes issue when coming from an older version of EE that had a broken LDAP implementation EE-3910
// e.g. when starting with CE => BE 2.14.1 then to BE 2.14.2
func (m *Migrator) updateLdapSettingsEE() error {
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	if len(settings.LDAPSettings.URLs) == 0 {
		settings.LDAPSettings.URLs = []string{}
		if url := settings.LDAPSettings.URL; url != "" {
			settings.LDAPSettings.URLs = append(settings.LDAPSettings.URLs, url)
		}
		settings.LDAPSettings.ServerType = portaineree.LDAPServerCustom
	}

	if settings.LDAPSettings.AdminGroupSearchSettings == nil {
		// The front end requires a slice with a single empty element to allow configuration
		settings.LDAPSettings.AdminGroupSearchSettings = []portainer.LDAPGroupSearchSettings{
			{},
		}
	}

	return m.settingsService.UpdateSettings(settings)
}

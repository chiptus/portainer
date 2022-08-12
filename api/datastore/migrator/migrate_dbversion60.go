package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (m *Migrator) migrateDBVersionToDB60() error {
	migrateLog.Info("- add gpu input field")
	if err := m.addGpuInputFieldDB60(); err != nil {
		return err
	}

	migrateLog.Info("- updating ldap settings")
	if err := m.updateLdapSettingsEE(); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) addGpuInputFieldDB60() error {

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		endpoint.Gpus = []portaineree.Pair{}
		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
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

	if settings.LDAPSettings.URLs == nil || len(settings.LDAPSettings.URLs) == 0 {
		settings.LDAPSettings.URLs = []string{}
		if url := settings.LDAPSettings.URL; url != "" {
			settings.LDAPSettings.URLs = append(settings.LDAPSettings.URLs, url)
		}
		settings.LDAPSettings.ServerType = portaineree.LDAPServerCustom
	}

	if settings.LDAPSettings.AdminGroupSearchSettings == nil {
		// The front end requires a slice with a single empty element to allow configuration
		settings.LDAPSettings.AdminGroupSearchSettings = []portaineree.LDAPGroupSearchSettings{
			{},
		}
	}

	return m.settingsService.UpdateSettings(settings)
}

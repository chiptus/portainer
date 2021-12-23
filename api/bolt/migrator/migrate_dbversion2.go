package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (m *Migrator) updateSettingsToDBVersion3() error {
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.AuthenticationMethod = portaineree.AuthenticationInternal
	legacySettings.LDAPSettings = portaineree.LDAPSettings{
		TLSConfig: portaineree.TLSConfiguration{},
		SearchSettings: []portaineree.LDAPSearchSettings{
			portaineree.LDAPSearchSettings{},
		},
	}

	return m.settingsService.UpdateSettings(legacySettings)
}

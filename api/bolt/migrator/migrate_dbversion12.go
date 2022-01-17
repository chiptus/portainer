package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (m *Migrator) updateSettingsToVersion13() error {
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.LDAPSettings.AutoCreateUsers = false
	legacySettings.LDAPSettings.GroupSearchSettings = []portaineree.LDAPGroupSearchSettings{
		{},
	}

	return m.settingsService.UpdateSettings(legacySettings)
}

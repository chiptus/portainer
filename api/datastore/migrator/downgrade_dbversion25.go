package migrator

// DowngradeSettingsFrom25 downgrade template settings for portainer v1.2
func (m *Migrator) DowngradeSettingsFrom25() error {
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.TemplatesURL = "https://raw.githubusercontent.com/portainer/templates/master/templates-1.20.0.json"

	err = m.settingsService.UpdateSettings(legacySettings)

	return err
}

package migrator

func (m *Migrator) migrateSettingsToDB30() error {
	migrateLog.Info("- updating settings")
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.OAuthSettings.SSO = false
	legacySettings.OAuthSettings.HideInternalAuth = false
	legacySettings.OAuthSettings.LogoutURI = ""
	return m.settingsService.UpdateSettings(legacySettings)
}

package migrator

import "github.com/rs/zerolog/log"

func (m *Migrator) migrateSettingsToDB30() error {
	log.Info().Msg("updating legacy settings")

	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.OAuthSettings.SSO = false
	legacySettings.OAuthSettings.HideInternalAuth = false
	legacySettings.OAuthSettings.LogoutURI = ""

	return m.settingsService.UpdateSettings(legacySettings)
}

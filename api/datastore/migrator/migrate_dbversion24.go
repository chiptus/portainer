package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

func (m *Migrator) updateSettingsToDB25() error {
	migrateLog.Info("Updating settings")

	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	if legacySettings.TemplatesURL == "" {
		legacySettings.TemplatesURL = portaineree.DefaultTemplatesURL
	}

	legacySettings.UserSessionTimeout = portaineree.DefaultUserSessionTimeout
	legacySettings.EnableTelemetry = true

	legacySettings.AllowContainerCapabilitiesForRegularUsers = true

	return m.settingsService.UpdateSettings(legacySettings)
}

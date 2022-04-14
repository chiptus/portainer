package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (m *Migrator) updateSettingsToDBVersion19() error {
	migrateLog.Info("- updating settings")
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	if legacySettings.EdgeAgentCheckinInterval == 0 {
		legacySettings.EdgeAgentCheckinInterval = portaineree.DefaultEdgeAgentCheckinIntervalInSeconds
	}

	return m.settingsService.UpdateSettings(legacySettings)
}

package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB33() error {
	if err := m.migrateSettingsToDB33(); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) migrateSettingsToDB33() error {
	log.Info().Msg("setting default kubctl shell")
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	settings.KubectlShellImage = portaineree.DefaultKubectlShellImage
	return m.settingsService.UpdateSettings(settings)
}

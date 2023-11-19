package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

func (migrator *Migrator) setDefaultCategoryForEdgeConfigForDB110() error {
	edgeConfigurations, err := migrator.EdgeConfigService.ReadAll()
	if err != nil {
		return err
	}

	for _, config := range edgeConfigurations {
		if config.Category == "" {
			config.Category = portaineree.EdgeConfigCategoryConfig
		}

		if config.Prev != nil && config.Prev.Category == "" {
			config.Prev.Category = portaineree.EdgeConfigCategoryConfig
		}

		err = migrator.EdgeConfigService.Update(config.ID, &config)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateAppTemplatesVersionForDB110 changes the templates URL to be empty if it was never changed
// from the default value (version 2.0 URL)
func (migrator *Migrator) updateAppTemplatesVersionForDB110() error {
	log.Info().Msg("updating app templates url to v3.0")

	version2URL := "https://raw.githubusercontent.com/portainer/templates/master/templates-2.0.json"

	settings, err := migrator.settingsService.Settings()
	if err != nil {
		return err
	}

	if settings.TemplatesURL == version2URL || settings.TemplatesURL == portainer.DefaultTemplatesURL {
		settings.TemplatesURL = ""
	}

	return migrator.settingsService.UpdateSettings(settings)
}

// setUseCacheForDB110 sets the user cache to true for all users
func (migrator *Migrator) setUserCacheForDB110() error {
	users, err := migrator.userService.ReadAll()
	if err != nil {
		return err
	}

	for i := range users {
		user := &users[i]
		user.UseCache = true
		if err := migrator.userService.Update(user.ID, user); err != nil {
			return err
		}
	}

	return nil
}

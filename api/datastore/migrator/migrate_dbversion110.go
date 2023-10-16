package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
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

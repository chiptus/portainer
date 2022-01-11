package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (m *Migrator) updateSettingsToDB24() error {
	migrateLog.Info("Updating settings")
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.AllowHostNamespaceForRegularUsers = true
	legacySettings.AllowDeviceMappingForRegularUsers = true
	legacySettings.AllowStackManagementForRegularUsers = true

	return m.settingsService.UpdateSettings(legacySettings)
}

func (m *Migrator) updateStacksToDB24() error {
	migrateLog.Info("Updating stacks")
	stacks, err := m.stackService.Stacks()
	if err != nil {
		return err
	}

	for idx := range stacks {
		stack := &stacks[idx]
		stack.Status = portaineree.StackStatusActive
		err := m.stackService.UpdateStack(stack.ID, stack)
		if err != nil {
			return err
		}
	}

	return nil
}
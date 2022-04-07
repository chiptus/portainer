package migrator

func (m *Migrator) migrateDBVersionToDB37() error {

	migrateLog.Info("Refreshing RBAC roles")
	if err := m.refreshRBACRoles(); err != nil {
		return err
	}

	migrateLog.Info("Refreshing user authorisations")
	if err := m.refreshUserAuthorizations(); err != nil {
		return err
	}

	return nil
}

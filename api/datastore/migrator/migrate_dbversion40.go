package migrator

func (m *Migrator) migrateDBVersionToDB40() error {
	migrateLog.Info("- refreshing RBAC roles")
	if err := m.refreshRBACRoles(); err != nil {
		return err
	}

	migrateLog.Info("- refreshing user authorizations")
	if err := m.refreshUserAuthorizations(); err != nil {
		return err
	}

	return nil
}

package migrator

func (m *Migrator) migrateDBVersionToDB31() error {
	migrateLog.Info("Refresh RBAC roles")
	if err := m.refreshRBACRoles(); err != nil {
		return err
	}
	return nil
}

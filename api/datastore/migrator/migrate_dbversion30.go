package migrator

func (m *Migrator) migrateDBVersionToDB31() error {
	return m.refreshRBACRoles()
}

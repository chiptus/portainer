package migrator

func (m *Migrator) migrateDBVersionToDB31() error {
	if err := m.refreshRBACRoles(); err != nil {
		return err
	}
	return nil
}

package migrator

func (m *Migrator) migrateDBVersionTo31() error {
	if err := m.refreshRBACRoles(); err != nil {
		return err
	}
	return nil
}

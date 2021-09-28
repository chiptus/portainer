package migrator

func (m *Migrator) migrateDBVersionToDB34() error {

	if err := m.refreshRBACRoles(); err != nil {
		return err
	}

	if err := m.refreshUserAuthorizations(); err != nil {
		return err
	}

	return nil
}

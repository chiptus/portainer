package migrator

func (m *Migrator) updateRbacRolesToDB29() error {
	return m.refreshRBACRoles()
}

package migrator

import "github.com/rs/zerolog/log"

func (m *Migrator) migrateDBVersionToDB34() error {
	log.Info().Msg("refreshing RBAC roles")

	if err := m.refreshRBACRoles(); err != nil {
		return err
	}

	log.Info().Msg("refreshing user authorisations")

	return m.refreshUserAuthorizations()
}

package migrator

import "github.com/rs/zerolog/log"

func (m *Migrator) migrateDBVersionToDB31() error {
	log.Info().Msg("refresh RBAC roles")

	return m.refreshRBACRoles()
}

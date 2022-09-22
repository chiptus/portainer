package migrator

import (
	"reflect"
	"runtime"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type migration struct {
	dbversion int
	migrate   func() error
}

func migrationError(err error, context string) error {
	return errors.Wrap(err, "failed in "+context)
}

func newMigration(dbversion int, migrate func() error) migration {
	return migration{
		dbversion: dbversion,
		migrate:   migrate,
	}
}

func dbTooOldError() error {
	return errors.New("migrating from less than Portainer 1.21.0 is not supported, please contact Portainer support.")
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Migrate checks the database version and migrate the existing data to the most recent data model.
func (m *Migrator) MigrateCE() error {
	// set DB to updating status
	err := m.versionService.StoreIsUpdating(true)
	if err != nil {
		return migrationError(err, "StoreIsUpdating")
	}

	migrations := []migration{
		// Portainer 1.21.1
		newMigration(17, dbTooOldError),

		// Portainer 1.21.0
		newMigration(18, m.updateUsersToDBVersion18),
		newMigration(18, m.updateEndpointsToDBVersion18),
		newMigration(18, m.updateEndpointGroupsToDBVersion18),
		newMigration(18, m.updateRegistriesToDBVersion18),

		// 1.22.0
		newMigration(19, m.updateSettingsToDBVersion19),

		// 1.22.1
		newMigration(20, m.updateUsersToDBVersion20),
		newMigration(20, m.updateSettingsToDBVersion20),
		newMigration(20, m.updateSchedulesToDBVersion20),

		// Portainer 1.23.0
		// DBVersion 21 is missing as it was shipped as via hotfix 1.22.2
		newMigration(22, m.updateResourceControlsToDBVersion22),
		newMigration(22, m.updateUsersAndRolesToDBVersion22),

		// Portainer 1.24.0
		newMigration(23, m.updateTagsToDBVersion23),
		newMigration(23, m.updateEndpointsAndEndpointGroupsToDBVersion23),

		// Portainer 1.24.1
		newMigration(24, m.updateSettingsToDB24),

		// Portainer 2.0.0
		newMigration(25, m.updateSettingsToDB25),
		newMigration(25, m.updateStacksToDB24), // yes this looks odd. Don't be tempted to move it

		// Portainer 2.1.0
		newMigration(26, m.updateEndpointSettingsToDB26),
		newMigration(26, m.refreshRBACRoles),

		// Portainer 2.2.0
		newMigration(27, m.updateStackResourceControlToDB27),

		// Portainer EE-2.4.0
		newMigration(28, m.updateUsersAndRolesToDBVersion28),

		// Portainer EE-2.4.0
		newMigration(29, m.refreshRBACRoles),

		// Portainer EE-2.6.0
		newMigration(30, m.migrateSettingsToDB30),

		// Portainer EE-2.7.0
		newMigration(31, m.refreshRBACRoles),

		// Portainer 2.9.0
		newMigration(32, m.migrateDBVersionToDB32),

		// Portainer 2.9.1, 2.9.2
		newMigration(33, m.migrateDBVersionToDB33),

		// Portainer 2.10.0
		newMigration(34, m.refreshRBACRoles),
		newMigration(34, m.refreshUserAuthorizations),

		// Portainer 2.9.3 (yep out of order, but 2.10 is EE only)
		newMigration(35, m.migrateDBVersionToDB35),

		// Portainer 2.11.x
		newMigration(36, m.migrateUsersToDB36),

		// Portainer 2.13.0
		newMigration(40, m.migrateDBVersionToDB40),

		// Portainer 2.14.0
		newMigration(50, m.migrateDBVersionToDB50),

		// Portainer 2.15
		newMigration(60, m.migrateDBVersionToDB60),

		// Portainer 2.16
		newMigration(70, m.migrateDBVersionToDB70),
	}

	var lastDbVersion int
	for _, migration := range migrations {
		if m.currentDBVersion < migration.dbversion {

			// Print the next line only when the version changes
			if migration.dbversion > lastDbVersion {
				log.Info().Int("to_version", migration.dbversion).Msg("migrating DB")
			}

			err := migration.migrate()
			if err != nil {
				return migrationError(err, GetFunctionName(migration.migrate))
			}
		}
		lastDbVersion = migration.dbversion
	}

	log.Info().Int("version", portaineree.DBVersion).Msg("setting DB version")

	err = m.versionService.StoreDBVersion(portaineree.DBVersion)
	if err != nil {
		return migrationError(err, "StoreDBVersion")
	}

	log.Info().Int("version", portaineree.DBVersion).Msg("updated DB version")

	// reset DB updating status
	return m.versionService.StoreIsUpdating(false)
}

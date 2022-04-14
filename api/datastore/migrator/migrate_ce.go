package migrator

import (
	"reflect"
	"runtime"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
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
	}

	var lastDbVersion int
	for _, migration := range migrations {
		if m.currentDBVersion < migration.dbversion {

			// Print the next line only when the version changes
			if migration.dbversion > lastDbVersion {
				migrateLog.Infof("Migrating DB to version %d", migration.dbversion)
			}

			err := migration.migrate()
			if err != nil {
				return migrationError(err, GetFunctionName(migration.migrate))
			}
		}
		lastDbVersion = migration.dbversion
	}

	if m.currentDBVersion < 37 {
		migrateLog.Info("Migrating to DB 37")
		if err := m.migrateDBVersionToDB37(); err != nil {
			return migrationError(err, "migrateDBVersionToDB37")
		}
	}

	migrateLog.Infof("Set DB version to %d", portaineree.DBVersion)
	err = m.versionService.StoreDBVersion(portaineree.DBVersion)
	if err != nil {
		return migrationError(err, "StoreDBVersion")
	}
	migrateLog.Infof("Updated DB version to %d", portaineree.DBVersion)

	// reset DB updating status
	return m.versionService.StoreIsUpdating(false)
}

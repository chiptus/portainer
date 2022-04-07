package migrator

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
)

func migrationError(err error, context string) error {
	return errors.Wrap(err, "failed in "+context)
}

// Migrate checks the database version and migrate the existing data to the most recent data model.
func (m *Migrator) MigrateCE() error {
	// set DB to updating status
	err := m.versionService.StoreIsUpdating(true)
	if err != nil {
		return migrationError(err, "StoreIsUpdating")
	}

	// Portainer 1.21.1
	if m.currentDBVersion < 17 {
		return migrationError(err, "migrating from less than Portainer 1.21.0 is not supported, please contact Portainer support.")
	}

	// Portainer 1.21.0
	if m.currentDBVersion < 18 {
		err := m.updateUsersToDBVersion18()
		if err != nil {
			return migrationError(err, "updateUsersToDBVersion18")
		}

		err = m.updateEndpointsToDBVersion18()
		if err != nil {
			return migrationError(err, "updateEndpointsToDBVersion18")
		}

		err = m.updateEndpointGroupsToDBVersion18()
		if err != nil {
			return migrationError(err, "updateEndpointGroupsToDBVersion18")
		}

		err = m.updateRegistriesToDBVersion18()
		if err != nil {
			return migrationError(err, "updateRegistriesToDBVersion18")
		}
	}

	// Portainer 1.22.0
	if m.currentDBVersion < 19 {
		err := m.updateSettingsToDBVersion19()
		if err != nil {
			return migrationError(err, "updateSettingsToDBVersion19")
		}
	}

	// Portainer 1.22.1
	if m.currentDBVersion < 20 {
		err := m.updateUsersToDBVersion20()
		if err != nil {
			return migrationError(err, "updateUsersToDBVersion20")
		}

		err = m.updateSettingsToDBVersion20()
		if err != nil {
			return migrationError(err, "updateSettingsToDBVersion20")
		}

		err = m.updateSchedulesToDBVersion20()
		if err != nil {
			return migrationError(err, "updateSchedulesToDBVersion20")
		}
	}

	// Portainer 1.23.0
	// DBVersion 21 is missing as it was shipped as via hotfix 1.22.2
	if m.currentDBVersion < 22 {
		err := m.updateResourceControlsToDBVersion22()
		if err != nil {
			return migrationError(err, "updateResourceControlsToDBVersion22")
		}

		err = m.updateUsersAndRolesToDBVersion22()
		if err != nil {
			return migrationError(err, "updateUsersAndRolesToDBVersion22")
		}
	}

	// Portainer 1.24.0
	if m.currentDBVersion < 23 {
		migrateLog.Info("Migrating to DB 23")
		err := m.updateTagsToDBVersion23()
		if err != nil {
			return migrationError(err, "updateTagsToDBVersion23")
		}

		err = m.updateEndpointsAndEndpointGroupsToDBVersion23()
		if err != nil {
			return migrationError(err, "updateEndpointsAndEndpointGroupsToDBVersion23")
		}
	}

	// Portainer 1.24.1
	if m.currentDBVersion < 24 {
		migrateLog.Info("Migrating to DB 24")
		err := m.updateSettingsToDB24()
		if err != nil {
			return migrationError(err, "updateSettingsToDB24")
		}
	}

	// Portainer 2.0.0
	if m.currentDBVersion < 25 {
		migrateLog.Info("Migrating to DB 25")
		err := m.updateSettingsToDB25()
		if err != nil {
			return migrationError(err, "updateSettingsToDB25")
		}

		err = m.updateStacksToDB24()
		if err != nil {
			return migrationError(err, "updateStacksToDB24")
		}
	}

	// Portainer 2.1.0
	if m.currentDBVersion < 26 {
		migrateLog.Info("Migrating to DB 26")
		err := m.updateEndpointSettingsToDB26()
		if err != nil {
			return migrationError(err, "updateEndpointSettingsToDB26")
		}
		migrateLog.Info("Successful migration of environments settings to DB version 26")

		err = m.updateRbacRolesToDB26()
		if err != nil {
			return migrationError(err, "updateRbacRolesToDB26")
		}
		migrateLog.Info("Successful migration of RBAC roles to DB version 26")
	}

	// Portainer 2.2.0
	if m.currentDBVersion < 27 {
		migrateLog.Info("Migrating to DB 27")
		err := m.updateStackResourceControlToDB27()
		if err != nil {
			return migrationError(err, "updateStackResourceControlToDB27")
		}
		migrateLog.Info("Successful migration to DB version 27")
	}

	// Portainer EE-2.4.0
	if m.currentDBVersion < 28 {
		migrateLog.Info("Migrating to DB 28")
		err := m.updateUsersAndRolesToDBVersion28()
		if err != nil {
			return migrationError(err, "updateUsersAndRolesToDBVersion28")
		}
		migrateLog.Info("Successful migration to DB version 28")
	}

	// Portainer EE-2.4.0
	if m.currentDBVersion < 29 {
		migrateLog.Info("Migrating to DB 29")
		err := m.refreshRBACRoles()
		if err != nil {
			return migrationError(err, "refreshRBACRoles")
		}
		migrateLog.Info("Successful migration to DB version 29")
	}

	// Portainer EE-2.6.0
	if m.currentDBVersion < 30 {
		migrateLog.Info("Migrating to DB 30")
		err := m.migrateDBVersionToDB30()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB30")
		}
		migrateLog.Info("Successful migration to DB version 30")
	}

	// Portainer EE-2.7.0
	if m.currentDBVersion < 31 {
		migrateLog.Info("Migrating to DB 31")
		err := m.migrateDBVersionToDB31()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB31")
		}
	}

	// Portainer 2.9.0
	if m.currentDBVersion < 32 {
		migrateLog.Info("Migrating to DB 32")
		if err := m.migrateDBVersionToDB32(); err != nil {
			return migrationError(err, "migrateDBVersionToDB32")
		}
	}

	// Portainer 2.9.1, 2.9.2
	if m.currentDBVersion < 33 {
		migrateLog.Info("Migrating to DB 33")
		err := m.migrateDBVersionToDB33()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB33")
		}
	}

	// Portainer 2.10.0
	if m.currentDBVersion < 34 {
		migrateLog.Info("Migrating to DB 34")
		err := m.migrateDBVersionToDB34()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB34")
		}
	}

	// Portainer 2.9.3 (yep out of order, but 2.10 is EE only)
	if m.currentDBVersion < 35 {
		migrateLog.Info("Migrating to DB 35")
		if err := m.migrateDBVersionToDB35(); err != nil {
			return migrationError(err, "migrateDBVersionToDB35")
		}
	}

	if m.currentDBVersion < 36 {
		migrateLog.Info("Migrating to DB 36")
		if err := m.migrateDBVersionToDB36(); err != nil {
			return migrationError(err, "migrateDBVersionToDB36")
		}
	}

	if m.currentDBVersion < 37 {
		migrateLog.Info("Migrating to DB 37")
		if err := m.migrateDBVersionToDB37(); err != nil {
			return migrationError(err, "migrateDBVersionToDB37")
		}
	}

	log.Println("Update DB version to ", portaineree.DBVersion)
	err = m.versionService.StoreDBVersion(portaineree.DBVersion)
	if err != nil {
		return migrationError(err, "StoreDBVersion")
	}
	migrateLog.Info(fmt.Sprintf("Updated DB version to %d", portaineree.DBVersion))

	// reset DB updating status
	return m.versionService.StoreIsUpdating(false)
}

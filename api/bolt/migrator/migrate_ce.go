package migrator

import (
	"log"

	werrors "github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
)

func migrationError(err error, context string) error {
	return werrors.Wrap(err, "failed in "+context)
}

// MigrateCE checks the database version and migrate the existing data to the most recent data model.
func (m *Migrator) MigrateCE() error {
	// set DB to updating status
	err := m.versionService.StoreIsUpdating(true)
	if err != nil {
		return migrationError(err, "StoreIsUpdating")
	}

	// Portainer < 1.12
	if m.currentDBVersion < 1 {
		err := m.updateAdminUserToDBVersion1()
		if err != nil {
			return migrationError(err, "updateAdminUserToDBVersion1")
		}
	}

	// Portainer 1.12.x
	if m.currentDBVersion < 2 {
		err := m.updateResourceControlsToDBVersion2()
		if err != nil {
			return migrationError(err, "updateResourceControlsToDBVersion2")
		}
		err = m.updateEndpointsToDBVersion2()
		if err != nil {
			return migrationError(err, "updateEndpointsToDBVersion2")
		}
	}

	// Portainer 1.13.x
	if m.currentDBVersion < 3 {
		err := m.updateSettingsToDBVersion3()
		if err != nil {
			return migrationError(err, "updateSettingsToDBVersion3")
		}
	}

	// Portainer 1.14.0
	if m.currentDBVersion < 4 {
		err := m.updateEndpointsToDBVersion4()
		if err != nil {
			return migrationError(err, "updateEndpointsToDBVersion4")
		}
	}

	// https://github.com/portainer/portainer/issues/1235
	if m.currentDBVersion < 5 {
		err := m.updateSettingsToVersion5()
		if err != nil {
			return migrationError(err, "updateSettingsToVersion5")
		}
	}

	// https://github.com/portainer/portainer/issues/1236
	if m.currentDBVersion < 6 {
		err := m.updateSettingsToVersion6()
		if err != nil {
			return migrationError(err, "updateSettingsToVersion6")
		}
	}

	// https://github.com/portainer/portainer/issues/1449
	if m.currentDBVersion < 7 {
		err := m.updateSettingsToVersion7()
		if err != nil {
			return migrationError(err, "updateSettingsToVersion7")
		}
	}

	if m.currentDBVersion < 8 {
		err := m.updateEndpointsToVersion8()
		if err != nil {
			return migrationError(err, "updateEndpointsToVersion8")
		}
	}

	// https: //github.com/portainer/portainer/issues/1396
	if m.currentDBVersion < 9 {
		err := m.updateEndpointsToVersion9()
		if err != nil {
			return migrationError(err, "updateEndpointsToVersion9")
		}
	}

	// https://github.com/portainer/portainer/issues/461
	if m.currentDBVersion < 10 {
		err := m.updateEndpointsToVersion10()
		if err != nil {
			return migrationError(err, "updateEndpointsToVersion10")
		}
	}

	// https://github.com/portainer/portainer/issues/1906
	if m.currentDBVersion < 11 {
		err := m.updateEndpointsToVersion11()
		if err != nil {
			return migrationError(err, "updateEndpointsToVersion11")
		}
	}

	// Portainer 1.18.0
	if m.currentDBVersion < 12 {
		err := m.updateEndpointsToVersion12()
		if err != nil {
			return migrationError(err, "updateEndpointsToVersion12")
		}

		err = m.updateEndpointGroupsToVersion12()
		if err != nil {
			return migrationError(err, "updateEndpointGroupsToVersion12")
		}

		err = m.updateStacksToVersion12()
		if err != nil {
			return migrationError(err, "updateStacksToVersion12")
		}
	}

	// Portainer 1.19.0
	if m.currentDBVersion < 13 {
		err := m.updateSettingsToVersion13()
		if err != nil {
			return migrationError(err, "updateSettingsToVersion13")
		}
	}

	// Portainer 1.19.2
	if m.currentDBVersion < 14 {
		err := m.updateResourceControlsToDBVersion14()
		if err != nil {
			return migrationError(err, "updateResourceControlsToDBVersion14")
		}
	}

	// Portainer 1.20.0
	if m.currentDBVersion < 15 {
		err := m.updateSettingsToDBVersion15()
		if err != nil {
			return migrationError(err, "updateSettingsToDBVersion15")
		}

		err = m.updateTemplatesToVersion15()
		if err != nil {
			return migrationError(err, "updateTemplatesToVersion15")
		}
	}

	if m.currentDBVersion < 16 {
		err := m.updateSettingsToDBVersion16()
		if err != nil {
			return migrationError(err, "updateSettingsToDBVersion16")
		}
	}

	// Portainer 1.20.1
	if m.currentDBVersion < 17 {
		err := m.updateExtensionsToDBVersion17()
		if err != nil {
			return migrationError(err, "updateExtensionsToDBVersion17")
		}
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
		err := m.updateSettingsToDB24()
		if err != nil {
			return migrationError(err, "updateSettingsToDB24")
		}
	}

	// Portainer 2.0.0
	if m.currentDBVersion < 25 {
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
		err := m.updateStackResourceControlToDB27()
		if err != nil {
			return migrationError(err, "updateStackResourceControlToDB27")
		}
		migrateLog.Info("Successful migration to DB version 27")
	}

	// Portainer EE-2.4.0
	if m.currentDBVersion < 28 {
		err := m.updateUsersAndRolesToDBVersion28()
		if err != nil {
			return migrationError(err, "updateUsersAndRolesToDBVersion28")
		}
		migrateLog.Info("Successful migration to DB version 28")
	}

	// Portainer EE-2.4.0
	if m.currentDBVersion < 29 {
		err := m.refreshRBACRoles()
		if err != nil {
			return migrationError(err, "refreshRBACRoles")
		}
		migrateLog.Info("Successful migration to DB version 29")
	}

	// Portainer EE-2.6.0
	if m.currentDBVersion < 30 {
		err := m.migrateDBVersionToDB30()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB30")
		}
		migrateLog.Info("Successful migration to DB version 30")
	}

	// Portainer EE-2.7.0
	if m.currentDBVersion < 31 {
		err := m.migrateDBVersionToDB31()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB31")
		}
	}

	// Portainer 2.9.0
	if m.currentDBVersion < 32 {
		if err := m.migrateDBVersionToDB32(); err != nil {
			return migrationError(err, "migrateDBVersionToDB32")
		}
	}

	// Portainer 2.9.1, 2.9.2
	if m.currentDBVersion < 33 {
		err := m.migrateDBVersionToDB33()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB33")
		}
	}

	// Portainer 2.10.0
	if m.currentDBVersion < 34 {
		err := m.migrateDBVersionToDB34()
		if err != nil {
			return migrationError(err, "migrateDBVersionToDB34")
		}
	}

	// Portainer 2.9.3 (yep out of order, but 2.10 is EE only)
	if m.currentDBVersion < 35 {
		if err := m.migrateDBVersionToDB35(); err != nil {
			return migrationError(err, "migrateDBVersionToDB35")
		}
	}

	log.Println("Update DB version to ", portainer.DBVersion)
	err = m.versionService.StoreDBVersion(portainer.DBVersion)
	if err != nil {
		return migrationError(err, "StoreDBVersion")
	}

	// reset DB updating status
	return m.versionService.StoreIsUpdating(false)
}

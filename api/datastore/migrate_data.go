package datastore

import (
	"fmt"
	"runtime/debug"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cli"
	dserrors "github.com/portainer/portainer-ee/api/dataservices/errors"
	plog "github.com/portainer/portainer-ee/api/datastore/log"
	"github.com/portainer/portainer-ee/api/datastore/migrator"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
	"github.com/sirupsen/logrus"

	werrors "github.com/pkg/errors"
)

const (
	beforePortainerVersionUpgradeBackup = "portainer.db.bak"
	beforePortainerUpgradeToEEBackup    = "portainer.db.before-EE-upgrade"
)

var migrateLog = plog.NewScopedLog("database, migrate")

func (store *Store) MigrateData() error {
	version, err := store.version()
	if err != nil {
		return err
	}

	edition := store.edition()

	migratorParams := &migrator.MigratorParameters{
		CurrentEdition:          edition,
		DatabaseVersion:         version,
		EndpointGroupService:    store.EndpointGroupService,
		EndpointService:         store.EndpointService,
		EndpointRelationService: store.EndpointRelationService,
		ExtensionService:        store.ExtensionService,
		RegistryService:         store.RegistryService,
		ResourceControlService:  store.ResourceControlService,
		RoleService:             store.RoleService,
		ScheduleService:         store.ScheduleService,
		SettingsService:         store.SettingsService,
		StackService:            store.StackService,
		TagService:              store.TagService,
		TeamMembershipService:   store.TeamMembershipService,
		UserService:             store.UserService,
		VersionService:          store.VersionService,
		FileService:             store.fileService,
		DockerhubService:        store.DockerHubService,
		AuthorizationService:    authorization.NewService(store),
	}

	return store.connectionMigrateData(migratorParams)
}

// FailSafeMigrate backup and restore DB if migration fail
func (store *Store) FailSafeMigrate(migrator *migrator.Migrator, version int) (err error) {
	defer func() {
		if e := recover(); e != nil {
			store.Rollback(true)
			// return error with cause and stacktrace (recover() doesn't include a stacktrace)
			err = fmt.Errorf("%v %s", e, string(debug.Stack()))
		}
	}()

	// !Important: we must use a named return value in the function definition and not a local
	// !variable referenced from the closure or else the return value will be incorrectly set
	return migrator.Migrate(version)
}

// MigrateData automatically migrate the data based on the DBVersion.
// This process is only triggered on an existing database, not if the database was just created.
// if force is true, then migrate regardless.
func (store *Store) connectionMigrateData(migratorParams *migrator.MigratorParameters) error {
	migrator := migrator.NewMigrator(migratorParams)

	// backup db file before upgrading DB to support rollback
	isUpdating, err := migratorParams.VersionService.IsUpdating()
	if err != nil && err != portainerDsErrors.ErrObjectNotFound {
		return err
	}

	ver := migrator.Version()
	if !isUpdating && ver != portaineree.DBVersion {
		err = store.backupVersion(migrator)
		if err != nil {
			return werrors.Wrapf(err, "failed to backup database")
		}
	}

	logrus.Infof("migrator.Version() = %v, portaineree.DBVersion = %v", migrator.Version(), portaineree.DBVersion)
	// Migrate to the latest CE version
	if migrator.Version() < portaineree.DBVersion {
		migrateLog.Info(fmt.Sprintf("Migrating database from version %v to %v", migrator.Version(), portaineree.DBVersion))
		err = store.FailSafeMigrate(migrator, portaineree.DBVersion)
		if err != nil {
			migrateLog.Error("An error occurred during database migration", err)
			return err
		}
	}

	logrus.Infof("portaineree.Edition = %v, portaineree.PortainerEE = %v, migrator.Edition() = %v", portaineree.Edition, portaineree.PortainerEE, migrator.Edition())
	if portaineree.Edition == portaineree.PortainerEE {
		// 2 – if DB is CE Edition we need to upgrade settings to EE
		if migrator.Edition() < portaineree.PortainerEE {
			err = migrator.UpgradeToEE()
			if err != nil {
				migrateLog.Error("An error occurred while upgrading database to EE", err)
				store.RollbackFailedUpgradeToEE()
				return err
			}
		}

		// 3 – if DB is EE Edition we need to migrate to latest version of EE
		if migrator.Edition() == portaineree.PortainerEE && migrator.Version() < portaineree.DBVersionEE {
			store.Backup()
			err = store.FailSafeMigrate(migrator, portaineree.DBVersionEE)
			if err != nil {
				migrateLog.Error("An error occurred while migrating EE database to latest version", err)
				store.Restore()
				return err
			}
		}
	}

	return nil
}

// RollbackFailedUpgradeToEE down migrate to previous version
func (store *Store) RollbackFailedUpgradeToEE() error {
	return store.restoreWithOptions(&BackupOptions{
		BackupFileName: beforePortainerUpgradeToEEBackup,
		Edition:        portaineree.PortainerCE,
	})
}

// backupVersion will backup the database or panic if any errors occur
func (store *Store) backupVersion(migrator *migrator.Migrator) error {
	migrateLog.Info("Backing up database prior to version upgrade...")

	options := getBackupRestoreOptions(store.commonBackupDir())

	_, err := store.backupWithOptions(options)
	if err != nil {
		migrateLog.Error("An error occurred during database backup", err)
		removalErr := store.removeWithOptions(options)
		if removalErr != nil {
			migrateLog.Error("An error occurred during store removal prior to backup", err)
		}
		return err
	}

	return nil
}

func (store *Store) rollbackToCE(forceUpdate bool) error {
	version, err := store.version()
	if err != nil {
		return err
	}

	edition := store.edition()

	migratorParams := &migrator.MigratorParameters{
		CurrentEdition:          edition,
		DatabaseVersion:         version,
		EndpointGroupService:    store.EndpointGroupService,
		EndpointService:         store.EndpointService,
		EndpointRelationService: store.EndpointRelationService,
		ExtensionService:        store.ExtensionService,
		RegistryService:         store.RegistryService,
		ResourceControlService:  store.ResourceControlService,
		RoleService:             store.RoleService,
		ScheduleService:         store.ScheduleService,
		SettingsService:         store.SettingsService,
		StackService:            store.StackService,
		TagService:              store.TagService,
		TeamMembershipService:   store.TeamMembershipService,
		UserService:             store.UserService,
		VersionService:          store.VersionService,
		FileService:             store.fileService,
		DockerhubService:        store.DockerHubService,
		AuthorizationService:    authorization.NewService(store),
	}
	migrator := migrator.NewMigrator(migratorParams)
	if err != nil {
		return err
	}

	migrateLog.Info(fmt.Sprintf("Current Software Edition: %s", migrator.Edition().GetEditionLabel()))
	migrateLog.Info(fmt.Sprintf("Current DB Version: %d", migrator.Version()))

	if migrator.Edition() == portaineree.PortainerCE {
		return dserrors.ErrMigrationToCE
	}

	rollbackConfirmMessage := "Are you sure you want to rollback Portainer to the community edition?"
	if !forceUpdate {
		confirmed, err := cli.Confirm(rollbackConfirmMessage)
		if err != nil || !confirmed {
			return err
		}
	}

	err = store.VersionService.StoreEdition(portaineree.PortainerCE)
	if err != nil {
		migrateLog.Error("An Error occurred with rolling back to the community edition", err)
		return err
	}

	err = store.downgradeLDAPSettings()
	if err != nil {
		migrateLog.Error("An Error occurred with rolling back LDAP URL setting", err)
		return err
	}
	migrateLog.Info("Rolled back to CE Edition.")
	return nil
}

func (store *Store) downgradeLDAPSettings() error {
	legacySettings, err := store.SettingsService.Settings()
	if err != nil {
		return err
	}

	urls := legacySettings.LDAPSettings.URLs
	if len(urls) > 0 {
		legacySettings.LDAPSettings.URL = urls[0] // use the first URL
		return store.SettingsService.UpdateSettings(legacySettings)
	}

	return nil
}

// Rollback to a pre-upgrade backup copy/snapshot of portainer.db
func (store *Store) connectionRollback(force bool) error {

	if !force {
		confirmed, err := cli.Confirm("Are you sure you want to rollback your database to the previous backup?")
		if err != nil || !confirmed {
			return err
		}
	}

	options := getBackupRestoreOptions(store.commonBackupDir())

	err := store.restoreWithOptions(options)
	if err != nil {
		return err
	}

	return store.connection.Close()
}

package bolt

import (
	"fmt"
	"runtime/debug"

	"github.com/portainer/portainer/api/cli"

	werrors "github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	errors "github.com/portainer/portainer/api/bolt/errors"
	plog "github.com/portainer/portainer/api/bolt/log"
	"github.com/portainer/portainer/api/bolt/migrator"
	"github.com/portainer/portainer/api/internal/authorization"
)

const (
	beforePortainerUpgradeToEEBackup    = "portainer.db.before-EE-upgrade"
	beforePortainerVersionUpgradeBackup = "portainer.db.bak"
)

var migrateLog = plog.NewScopedLog("bolt, migrate")

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
func (store *Store) MigrateData(force bool) error {
	// 0 – if DB is new then we don't need to migrate any data and just set version and edition to latest EE
	if store.isNew && !force {
		err := store.VersionService.StoreDBVersion(portainer.DBVersionEE)
		if err != nil {
			return err
		}

		err = store.VersionService.StoreEdition(portainer.PortainerEE)
		if err != nil {
			return err
		}

		return nil
	}

	migrator, err := store.newMigrator()
	if err != nil {
		return err
	}

	// backup db file before upgrading DB to support rollback
	isUpdating, err := store.VersionService.IsUpdating()
	if err != nil && err != errors.ErrObjectNotFound {
		return err
	}

	if !isUpdating && migrator.Version() != portainer.DBVersion {
		err = store.backupVersion(migrator)
		if err != nil {
			return werrors.Wrapf(err, "failed to backup database")
		}
	}

	if migrator.Edition() == portainer.PortainerCE {
		// backup before migrating
		store.BackupWithOptions(&BackupOptions{
			BackupFileName: beforePortainerUpgradeToEEBackup,
			Edition:        portainer.PortainerCE,
		})

		store.VersionService.StorePreviousDBVersion(migrator.Version())

		// 1 – We need to migrate DB to latest CE version

		if migrator.Version() < portainer.DBVersion {
			store.Backup()
			err = store.FailSafeMigrate(migrator, portainer.DBVersion)
			if err != nil {
				store.Restore()
				migrateLog.Error("An error occurred while migrating CE database to latest version", err)
				return err
			}
		}

	}

	if portainer.Edition == portainer.PortainerEE {
		// 2 – if DB is CE Edition we need to upgrade settings to EE
		if migrator.Edition() < portainer.PortainerEE {
			err = migrator.UpgradeToEE()
			if err != nil {
				migrateLog.Error("An error occurred while upgrading database to EE", err)
				store.RollbackFailedUpgradeToEE()
				return err
			}
		}

		// 3 – if DB is EE Edition we need to migrate to latest version of EE
		if migrator.Edition() == portainer.PortainerEE && migrator.Version() < portainer.DBVersionEE {
			store.Backup()
			err = store.FailSafeMigrate(migrator, portainer.DBVersionEE)
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
	return store.RestoreWithOptions(&BackupOptions{
		BackupFileName: beforePortainerUpgradeToEEBackup,
		Edition:        portainer.PortainerCE,
	})
}

// RollbackToCE rollbacks the store to the current ce version
func (store *Store) RollbackToCE() error {
	// RollbackToCE rollbacks the store to the current ce version
	return store.rollbackToCE(false)
}

func (store *Store) rollbackToCE(forceUpdate bool) error {
	migrator, err := store.newMigrator()
	if err != nil {
		return err
	}

	migrateLog.Info(fmt.Sprintf("Current Software Edition: %s", migrator.Edition().GetEditionLabel()))
	migrateLog.Info(fmt.Sprintf("Current DB Version: %d", migrator.Version()))

	if migrator.Edition() == portainer.PortainerCE {
		return errors.ErrMigrationToCE
	}

	previousVersion, err := store.VersionService.PreviousDBVersion()
	if err != nil {
		migrateLog.Error("An Error occurred with retrieving previous DB version", err)
		return err
	}

	if !forceUpdate {
		confirmed, err := cli.Confirm(fmt.Sprintf("Are you sure you want to rollback your database to %d?", previousVersion))
		if err != nil || !confirmed {
			return err
		}
	}

	if previousVersion < 25 {
		migrator.DowngradeSettingsFrom25()
	}

	err = store.VersionService.StoreDBVersion(previousVersion)
	if err != nil {
		migrateLog.Error(fmt.Sprintf("An Error occurred with rolling back to CE Edition, DB Version %d", previousVersion), err)
		return err
	}

	err = store.VersionService.StoreEdition(portainer.PortainerCE)
	if err != nil {
		migrateLog.Error(fmt.Sprintf("An Error occurred with rolling back to CE Edition, DB Version %d", previousVersion), err)
		return err
	}

	migrateLog.Info(fmt.Sprintf("Rolled back to CE Edition, DB Version %d", previousVersion))
	return nil
}

func (store *Store) newMigrator() (*migrator.Migrator, error) {

	version, err := store.version()
	if err != nil {
		return nil, err
	}

	edition := store.edition()

	params := &migrator.Parameters{
		DB:              store.connection.DB,
		DatabaseVersion: version,
		CurrentEdition:  edition,

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

	return migrator.NewMigrator(params), nil
}

// getBackupRestoreOptions returns options to store db at common backup dir location; used by:
// - db backup prior to version upgrade
// - db rollback
func getBackupRestoreOptions(store *Store) *BackupOptions {
	return &BackupOptions{
		BackupDir:      store.commonBackupDir(),
		BackupFileName: beforePortainerVersionUpgradeBackup,
	}
}

// backupVersion will backup the database or panic if any errors occur
func (store *Store) backupVersion(migrator *migrator.Migrator) error {
	migrateLog.Info("Backing up database prior to version upgrade...")

	options := getBackupRestoreOptions(store)

	_, err := store.BackupWithOptions(options)
	if err != nil {
		migrateLog.Error("An error occurred during database backup", err)
		store.RemoveWithOptions(options)
		return err
	}

	return nil
}

// Rollback to a pre-upgrade backup copy/snapshot of portainer.db
func (store *Store) Rollback(force bool) error {

	if !force {
		confirmed, err := cli.Confirm("Are you sure you want to rollback your database to the previous backup?")
		if err != nil || !confirmed {
			return err
		}
	}

	options := getBackupRestoreOptions(store)

	err := store.RestoreWithOptions(options)
	if err != nil {
		return err
	}

	return store.Close()
}

package datastore

import (
	"fmt"
	"runtime/debug"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cli"
	dserrors "github.com/portainer/portainer-ee/api/dataservices/errors"
	"github.com/portainer/portainer-ee/api/datastore/migrator"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	beforePortainerVersionUpgradeBackup = "portainer.db.bak"
	beforePortainerUpgradeToEEBackup    = "portainer.db.before-EE-upgrade"
)

func (store *Store) MigrateData() error {
	version, err := store.version()
	if err != nil {
		return err
	}

	// Backup Database
	backupPath, err := store.Backup()
	if err != nil {
		return errors.Wrap(err, "while backing up db before migration")
	}

	edition := store.edition()

	migratorParams := &migrator.MigratorParameters{
		CurrentEdition:           edition,
		DatabaseVersion:          version,
		CloudProvisioningService: store.CloudProvisioningService,
		EndpointGroupService:     store.EndpointGroupService,
		EndpointService:          store.EndpointService,
		EndpointRelationService:  store.EndpointRelationService,
		ExtensionService:         store.ExtensionService,
		RegistryService:          store.RegistryService,
		ResourceControlService:   store.ResourceControlService,
		RoleService:              store.RoleService,
		ScheduleService:          store.ScheduleService,
		SettingsService:          store.SettingsService,
		StackService:             store.StackService,
		TagService:               store.TagService,
		TeamMembershipService:    store.TeamMembershipService,
		UserService:              store.UserService,
		VersionService:           store.VersionService,
		FileService:              store.fileService,
		DockerhubService:         store.DockerHubService,
		CloudCredentialService:   store.CloudCredentialService,
		AuthorizationService:     authorization.NewService(store),
	}

	// restore on error
	err = store.connectionMigrateData(migratorParams)
	if err != nil {
		log.Error().Err(err).Msg("while DB migration, restoring DB")

		options := BackupOptions{
			BackupPath: backupPath,
		}

		err := store.restoreWithOptions(&options)
		if err != nil {
			log.Fatal().
				Str("database_file", store.databasePath()).
				Str("backup", options.BackupPath).Err(err).
				Msg("failed restoring the backup, Portainer database file needs to restored manually by replacing the database file with a recent backup")
		}
	}

	return err
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
			return errors.Wrapf(err, "failed to backup database")
		}
	}

	log.Info().
		Int("migrator_version", migrator.Version()).
		Int("portaineree_db_version", portaineree.DBVersion).
		Msg("migrating database")

	// Migrate to the latest CE version
	if migrator.Version() < portaineree.DBVersion {
		log.Info().Int("from", migrator.Version()).Int("to", portaineree.DBVersion).Msg("migrating database version")

		err = store.FailSafeMigrate(migrator, portaineree.DBVersion)
		if err != nil {
			log.Error().Err(err).Msg("an error occurred during database migration")

			return err
		}
	}

	log.Info().
		Str("portainer_edition", portaineree.Edition.GetEditionLabel()).
		Str("migrator_edition", migrator.Edition().GetEditionLabel()).
		Str("store_edition", store.edition().GetEditionLabel()).
		Msg("")

	// If DB is CE Edition we need to upgrade settings to EE
	if migrator.Edition() < portaineree.PortainerEE {
		err = migrator.UpgradeToEE()
		if err != nil {
			log.Error().Err(err).Msg("an error occurred while upgrading database to EE")

			store.RollbackFailedUpgradeToEE()
			return err
		}
	}

	// If DB is EE Edition we need to migrate to latest version of EE
	if migrator.Edition() == portaineree.PortainerEE && migrator.Version() < portaineree.DBVersionEE {
		err = store.FailSafeMigrate(migrator, portaineree.DBVersionEE)
		if err != nil {
			log.Error().Err(err).Msg("an error occurred while migrating EE database to latest version")

			return err
		}
	}

	// Always calling CreateOrUpdatePredefinedRoles as sometimes, it was noticed that
	// this method was not run and some roles were missing from the database
	err = store.RoleService.CreateOrUpdatePredefinedRoles()
	if err != nil {
		return errors.Wrap(err, "failed refreshing predefined roles")
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
	log.Info().Msg("backing up database prior to version upgrade")

	options := getBackupRestoreOptions(store.commonBackupDir())

	_, err := store.backupWithOptions(options)
	if err != nil {
		log.Error().Err(err).Msg("an error occurred during database backup")

		removalErr := store.removeWithOptions(options)
		if removalErr != nil {
			log.Error().Err(err).Msg("an error occurred during store removal prior to backup")
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
		CurrentEdition:           edition,
		DatabaseVersion:          version,
		CloudProvisioningService: store.CloudProvisioningService,
		CloudCredentialService:   store.CloudCredentialService,
		EndpointGroupService:     store.EndpointGroupService,
		EndpointService:          store.EndpointService,
		EndpointRelationService:  store.EndpointRelationService,
		ExtensionService:         store.ExtensionService,
		RegistryService:          store.RegistryService,
		ResourceControlService:   store.ResourceControlService,
		RoleService:              store.RoleService,
		ScheduleService:          store.ScheduleService,
		SettingsService:          store.SettingsService,
		StackService:             store.StackService,
		TagService:               store.TagService,
		TeamMembershipService:    store.TeamMembershipService,
		UserService:              store.UserService,
		VersionService:           store.VersionService,
		FileService:              store.fileService,
		DockerhubService:         store.DockerHubService,
		AuthorizationService:     authorization.NewService(store),
		PodSecurityService:       store.PodSecurityService,
	}
	migrator := migrator.NewMigrator(migratorParams)
	if err != nil {
		return err
	}

	log.Info().
		Str("current_software_edition", migrator.Edition().GetEditionLabel()).
		Int("current_DB_version", migrator.Version()).
		Msg("")

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
		log.Error().Err(err).Msg("an error occurred with rolling back to the community edition")

		return err
	}

	err = store.downgradeLDAPSettings()
	if err != nil {
		log.Error().Err(err).Msg("an error occurred with rolling back LDAP URL setting")

		return err
	}

	log.Info().Msg("rolled back to CE Edition")

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

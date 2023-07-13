package datastore

import (
	"fmt"
	"runtime/debug"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cli"
	"github.com/portainer/portainer-ee/api/database/models"
	dserrors "github.com/portainer/portainer-ee/api/dataservices/errors"
	"github.com/portainer/portainer-ee/api/datastore/migrator"
	"github.com/portainer/portainer-ee/api/internal/authorization"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	beforePortainerVersionUpgradeBackup = "portainer.db.bak"
	beforePortainerUpgradeToEEBackup    = "portainer.db.before-EE-upgrade"
)

func (store *Store) MigrateData() error {
	updating, err := store.VersionService.IsUpdating()
	if err != nil {
		return errors.Wrap(err, "while checking if the store is updating")
	}

	if updating {
		return dserrors.ErrDatabaseIsUpdating
	}

	// migrate new version bucket if required (doesn't write anything to db yet)
	version, err := store.getOrMigrateLegacyVersion()
	if err != nil {
		return errors.Wrap(err, "while migrating legacy version")
	}

	migratorParams := store.newMigratorParameters(version, store.flags)
	migrator := migrator.NewMigrator(migratorParams)

	if !migrator.NeedsMigration() {
		return nil
	}

	// before we alter anything in the DB, create a backup
	backupPath, err := store.Backup(version)
	if err != nil {
		return errors.Wrap(err, "while backing up database")
	}

	err = store.FailSafeMigrate(migrator, version)
	if err != nil {
		err = errors.Wrap(err, "failed to migrate database")

		log.Warn().Msg("migration failed, restoring database to previous version")
		err = store.restoreWithOptions(&BackupOptions{BackupPath: backupPath})
		if err != nil {
			return errors.Wrap(err, "failed to restore database")
		}

		log.Info().Msg("database restored to previous version")
		return err
	}

	return nil
}

// FailSafeMigrate backup and restore DB if migration fail
func (store *Store) FailSafeMigrate(migrator *migrator.Migrator, version *models.Version) (err error) {
	defer func() {
		if e := recover(); e != nil {
			// return error with cause and stacktrace (recover() doesn't include a stacktrace)
			err = fmt.Errorf("%v %s", e, string(debug.Stack()))
		}
	}()

	err = store.VersionService.StoreIsUpdating(true)
	if err != nil {
		return err
	}

	// now update the version to the new struct (if required)
	err = store.finishMigrateLegacyVersion(version)
	if err != nil {
		return errors.Wrap(err, "while updating version")
	}

	log.Info().Msg("migrating database from version " + version.SchemaVersion + " to " + portaineree.APIVersion)

	err = migrator.Migrate()
	if err != nil {
		return err
	}

	// If DB is CE Edition we need to upgrade settings to EE
	if portaineree.SoftwareEdition(version.Edition) < portaineree.PortainerEE {
		err = migrator.UpgradeToEE()
		if err != nil {
			return errors.Wrap(err, "while upgrading to EE")
		}
	}

	err = store.VersionService.StoreIsUpdating(false)
	if err != nil {
		return errors.Wrap(err, "failed to update the store")
	}

	return nil
}

func (store *Store) newMigratorParameters(version *models.Version, flags *portaineree.CLIFlags) *migrator.MigratorParameters {
	return &migrator.MigratorParameters{
		Flags:                    flags,
		CurrentDBVersion:         version,
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
		SnapshotService:          store.SnapshotService,
		StackService:             store.StackService,
		TagService:               store.TagService,
		TeamMembershipService:    store.TeamMembershipService,
		UserService:              store.UserService,
		VersionService:           store.VersionService,
		FileService:              store.fileService,
		DockerhubService:         store.DockerHubService,
		CloudCredentialService:   store.CloudCredentialService,
		AuthorizationService:     authorization.NewService(store),
		EdgeStackService:         store.EdgeStackService,
		EdgeJobService:           store.EdgeJobService,
		PodSecurityService:       store.PodSecurityService,
		EdgeUpdateService:        store.EdgeUpdateScheduleService,
		EdgeGroupService:         store.EdgeGroupService,
		FDOProfilesService:       store.FDOProfilesService,
		TunnelServerService:      store.TunnelServerService,
	}
}

// RollbackFailedUpgradeToEE down migrate to previous version
func (store *Store) RollbackFailedUpgradeToEE() error {
	return store.restoreWithOptions(&BackupOptions{
		BackupFileName: beforePortainerUpgradeToEEBackup,
		Edition:        portaineree.PortainerCE,
	})
}

func (store *Store) rollbackToCE(forceUpdate bool) error {
	v, err := store.VersionService.Version()
	if err != nil {
		return err
	}

	edition := portaineree.SoftwareEdition(v.Edition)

	log.Info().
		Str("current_software_edition", edition.GetEditionLabel()).
		Str("current_DB_version", v.SchemaVersion).Msg("")

	if edition == portaineree.PortainerCE {
		return dserrors.ErrMigrationToCE
	}

	rollbackConfirmMessage := "Are you sure you want to rollback Portainer to the community edition?"
	if !forceUpdate {
		confirmed, err := cli.Confirm(rollbackConfirmMessage)
		if err != nil || !confirmed {
			return err
		}
	}

	v.Edition = int(portaineree.PortainerCE)

	err = store.VersionService.UpdateVersion(v)
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

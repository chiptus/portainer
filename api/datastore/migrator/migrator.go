package migrator

import (
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices/cloudcredential"
	"github.com/portainer/portainer-ee/api/dataservices/cloudprovisioning"
	"github.com/portainer/portainer-ee/api/dataservices/dockerhub"
	"github.com/portainer/portainer-ee/api/dataservices/endpoint"
	"github.com/portainer/portainer-ee/api/dataservices/endpointgroup"
	"github.com/portainer/portainer-ee/api/dataservices/endpointrelation"
	"github.com/portainer/portainer-ee/api/dataservices/extension"
	"github.com/portainer/portainer-ee/api/dataservices/fdoprofile"
	"github.com/portainer/portainer-ee/api/dataservices/podsecurity"
	"github.com/portainer/portainer-ee/api/dataservices/registry"
	"github.com/portainer/portainer-ee/api/dataservices/resourcecontrol"
	"github.com/portainer/portainer-ee/api/dataservices/role"
	"github.com/portainer/portainer-ee/api/dataservices/schedule"
	"github.com/portainer/portainer-ee/api/dataservices/settings"
	"github.com/portainer/portainer-ee/api/dataservices/snapshot"
	"github.com/portainer/portainer-ee/api/dataservices/stack"
	"github.com/portainer/portainer-ee/api/dataservices/tag"
	"github.com/portainer/portainer-ee/api/dataservices/teammembership"
	"github.com/portainer/portainer-ee/api/dataservices/user"
	"github.com/portainer/portainer-ee/api/dataservices/version"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
)

type (
	// Migrator defines a service to migrate data after a Portainer version update.
	Migrator struct {
		currentDBVersion *models.Version
		migrations       []Migrations

		cloudProvisionService   *cloudprovisioning.Service
		endpointGroupService    *endpointgroup.Service
		endpointService         *endpoint.Service
		endpointRelationService *endpointrelation.Service
		extensionService        *extension.Service
		fdoProfilesService      *fdoprofile.Service
		registryService         *registry.Service
		resourceControlService  *resourcecontrol.Service
		roleService             *role.Service
		scheduleService         *schedule.Service
		settingsService         *settings.Service
		snapshotService         *snapshot.Service
		stackService            *stack.Service
		tagService              *tag.Service
		teamMembershipService   *teammembership.Service
		userService             *user.Service
		versionService          *version.Service
		fileService             portainer.FileService
		authorizationService    *authorization.Service
		dockerhubService        *dockerhub.Service
		cloudCredentialService  *cloudcredential.Service
	}

	// MigratorParameters represents the required parameters to create a new Migrator instance.
	MigratorParameters struct {
		CurrentDBVersion         *models.Version
		CloudProvisioningService *cloudprovisioning.Service
		EndpointGroupService     *endpointgroup.Service
		EndpointService          *endpoint.Service
		EndpointRelationService  *endpointrelation.Service
		ExtensionService         *extension.Service
		FDOProfilesService       *fdoprofile.Service
		RegistryService          *registry.Service
		ResourceControlService   *resourcecontrol.Service
		RoleService              *role.Service
		ScheduleService          *schedule.Service
		SettingsService          *settings.Service
		SnapshotService          *snapshot.Service
		StackService             *stack.Service
		TagService               *tag.Service
		TeamMembershipService    *teammembership.Service
		UserService              *user.Service
		VersionService           *version.Service
		FileService              portainer.FileService
		AuthorizationService     *authorization.Service
		DockerhubService         *dockerhub.Service
		PodSecurityService       *podsecurity.Service
		CloudCredentialService   *cloudcredential.Service
	}

	Migrations struct {
		version        *semver.Version
		migrationFuncs MigrationFuncs
	}

	MigrationFuncs []func() error
)

// NewMigrator creates a new Migrator.
func NewMigrator(parameters *MigratorParameters) *Migrator {
	migrator := &Migrator{
		currentDBVersion:        parameters.CurrentDBVersion,
		endpointGroupService:    parameters.EndpointGroupService,
		endpointService:         parameters.EndpointService,
		endpointRelationService: parameters.EndpointRelationService,
		extensionService:        parameters.ExtensionService,
		fdoProfilesService:      parameters.FDOProfilesService,
		registryService:         parameters.RegistryService,
		resourceControlService:  parameters.ResourceControlService,
		roleService:             parameters.RoleService,
		scheduleService:         parameters.ScheduleService,
		settingsService:         parameters.SettingsService,
		snapshotService:         parameters.SnapshotService,
		tagService:              parameters.TagService,
		teamMembershipService:   parameters.TeamMembershipService,
		stackService:            parameters.StackService,
		userService:             parameters.UserService,
		versionService:          parameters.VersionService,
		fileService:             parameters.FileService,
		authorizationService:    parameters.AuthorizationService,
		dockerhubService:        parameters.DockerhubService,
		cloudCredentialService:  parameters.CloudCredentialService,
	}

	migrator.initMigrations()
	return migrator
}

func (m *Migrator) CurrentDBVersion() string {
	return m.currentDBVersion.SchemaVersion
}

func (m *Migrator) CurrentDBEdition() portaineree.SoftwareEdition {
	return portaineree.SoftwareEdition(m.currentDBVersion.Edition)
}

func (m *Migrator) CurrentSemanticDBVersion() *semver.Version {
	v, err := semver.NewVersion(m.currentDBVersion.SchemaVersion)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to parse current version")
	}

	return v
}

func (m *Migrator) addMigrations(v string, funcs ...func() error) {
	m.migrations = append(m.migrations, Migrations{
		version:        semver.MustParse(v),
		migrationFuncs: funcs,
	})
}

func (m *Migrator) latestMigrations() Migrations {
	return m.migrations[len(m.migrations)-1]
}

// !NOTE: Migration funtions should ideally be idempotent.
// !      Which simply means the function can run over the same data many times but only transform it once.
// !      In practice this really just means an extra check or two to ensure we're not destroying valid data.
// This is not a hard rule though.  Understand the limitations.  A migration function may only run over
// the data more than once if a new migration function is added and the version of your database schema is
// the same.  e.g. two developers working on the same version add two different functions for different things.
// This changes the migration funcs count and so the migrations for that version will be run again.
// Where you're changing from an older revision e.g. 2.15 to a newer revision 2.16 only migrations to 2.16 will be
// run - just like previous versions before this refactor.

func (m *Migrator) initMigrations() {
	// !IMPORTANT: Do not be tempted to alter the order of these migrations.
	// !           Even though one of them looks out of order. Caused by history related
	// !           to maintaining two versions and releasing at different times

	m.addMigrations("1.0.0", dbTooOldError) // default version found after migration

	m.addMigrations("1.21",
		m.updateUsersToDBVersion18,
		m.updateEndpointsToDBVersion18,
		m.updateEndpointGroupsToDBVersion18,
		m.updateRegistriesToDBVersion18)

	m.addMigrations("1.22", m.updateSettingsToDBVersion19)

	m.addMigrations("1.22.1",
		m.updateUsersToDBVersion20,
		m.updateSettingsToDBVersion20,
		m.updateSchedulesToDBVersion20)

	m.addMigrations("1.23",
		m.updateResourceControlsToDBVersion22,
		m.updateUsersAndRolesToDBVersion22)

	m.addMigrations("1.24",
		m.updateTagsToDBVersion23,
		m.updateEndpointsAndEndpointGroupsToDBVersion23)

	m.addMigrations("1.24.1", m.updateSettingsToDB24)

	m.addMigrations("2.0",
		m.updateSettingsToDB25,
		m.updateStacksToDB24)

	m.addMigrations("2.1",
		m.updateEndpointSettingsToDB26,
		m.refreshRBACRoles)

	m.addMigrations("2.2",
		m.updateStackResourceControlToDB27)

	m.addMigrations("2.4",
		m.updateUsersAndRolesToDBVersion28,
		m.refreshRBACRoles)

	m.addMigrations("2.6", m.migrateSettingsToDB30)
	m.addMigrations("2.7", m.migrateDBVersionToDB31)
	m.addMigrations("2.9", m.migrateDBVersionToDB32)
	m.addMigrations("2.9.2", m.migrateDBVersionToDB33)
	m.addMigrations("2.10", m.migrateDBVersionToDB34)
	m.addMigrations("2.9.3", m.migrateDBVersionToDB35)
	m.addMigrations("2.11", m.migrateDBVersionToDB36)
	m.addMigrations("2.13", m.migrateDBVersionToDB40)
	m.addMigrations("2.14", m.migrateDBVersionToDB50)
	m.addMigrations("2.15", m.migrateDBVersionToDB60)
	m.addMigrations("2.16", m.migrateDBVersionToDB70)
	m.addMigrations("2.16.1", m.migrateDBVersionToDB71)

	// Add new migrations below...
	// One function per migration, each versions migration funcs in the same file.

	// !NOTE: refreshRBACRoles is now always called (see Always below)
	// !      there is no need to add it to the migrations list above
}

// Always is always run at the end of migrations
func (m *Migrator) Always() error {
	// Always calling CreateOrUpdatePredefinedRoles as sometimes, it was noticed that
	// this method was not run and some roles were missing from the database
	err := m.roleService.CreateOrUpdatePredefinedRoles()
	if err != nil {
		return errors.Wrap(err, "failed refreshing predefined roles")
	}

	err = m.refreshRBACRoles()
	if err != nil {
		return errors.Wrap(err, "failed refreshing RBAC roles")
	}

	return err
}

func dbTooOldError() error {
	return errors.New("migrating from less than Portainer 1.21.0 is not supported, please contact Portainer support.")
}

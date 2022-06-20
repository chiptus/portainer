package migrator

import (
	"github.com/portainer/portainer-ee/api/dataservices/cloudcredential"
	"github.com/portainer/portainer-ee/api/dataservices/cloudprovisioning"
	"github.com/portainer/portainer-ee/api/dataservices/fdoprofile"
	"github.com/portainer/portainer-ee/api/dataservices/podsecurity"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices/dockerhub"
	"github.com/portainer/portainer-ee/api/dataservices/endpoint"
	"github.com/portainer/portainer-ee/api/dataservices/endpointgroup"
	"github.com/portainer/portainer-ee/api/dataservices/endpointrelation"
	"github.com/portainer/portainer-ee/api/dataservices/extension"
	"github.com/portainer/portainer-ee/api/dataservices/registry"
	"github.com/portainer/portainer-ee/api/dataservices/resourcecontrol"
	"github.com/portainer/portainer-ee/api/dataservices/role"
	"github.com/portainer/portainer-ee/api/dataservices/schedule"
	"github.com/portainer/portainer-ee/api/dataservices/settings"
	"github.com/portainer/portainer-ee/api/dataservices/stack"
	"github.com/portainer/portainer-ee/api/dataservices/tag"
	"github.com/portainer/portainer-ee/api/dataservices/teammembership"
	"github.com/portainer/portainer-ee/api/dataservices/user"
	"github.com/portainer/portainer-ee/api/dataservices/version"
	plog "github.com/portainer/portainer-ee/api/datastore/log"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
)

var migrateLog = plog.NewScopedLog("database, migrate")

type (
	// Migrator defines a service to migrate data after a Portainer version update.
	Migrator struct {
		currentDBVersion int
		currentEdition   portaineree.SoftwareEdition

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
		DatabaseVersion int
		CurrentEdition  portaineree.SoftwareEdition

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
)

// NewMigrator creates a new Migrator.
func NewMigrator(parameters *MigratorParameters) *Migrator {
	return &Migrator{
		currentEdition:          parameters.CurrentEdition,
		currentDBVersion:        parameters.DatabaseVersion,
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
}

// Version exposes version of database
func (migrator *Migrator) Version() int {
	return migrator.currentDBVersion
}

// Edition exposes edition of portainer
func (migrator *Migrator) Edition() portaineree.SoftwareEdition {
	return migrator.currentEdition
}

// Migrate helper to upgrade DB
func (migrator *Migrator) Migrate(version int) error {
	migrateLog.Infof("Migrating %s database from version %d to %d.", migrator.Edition().GetEditionLabel(), migrator.currentDBVersion, version)
	err := migrator.MigrateCE() //CE
	if err != nil {
		migrateLog.Error("An error occurred during database migration", err)
		return err
	}

	migrator.versionService.StoreDBVersion(version)
	migrator.currentDBVersion = version
	return nil
}

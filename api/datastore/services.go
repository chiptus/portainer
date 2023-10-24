package datastore

import (
	"fmt"
	"os"

	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/dataservices/apikeyrepository"
	"github.com/portainer/portainer-ee/api/dataservices/cloudcredential"
	"github.com/portainer/portainer-ee/api/dataservices/cloudprovisioning"
	"github.com/portainer/portainer-ee/api/dataservices/dockerhub"
	"github.com/portainer/portainer-ee/api/dataservices/edgeasynccommand"
	"github.com/portainer/portainer-ee/api/dataservices/edgeconfig"
	"github.com/portainer/portainer-ee/api/dataservices/edgeconfigstate"
	"github.com/portainer/portainer-ee/api/dataservices/edgegroup"
	"github.com/portainer/portainer-ee/api/dataservices/edgestack"
	"github.com/portainer/portainer-ee/api/dataservices/edgestacklog"
	"github.com/portainer/portainer-ee/api/dataservices/edgeupdateschedule"
	"github.com/portainer/portainer-ee/api/dataservices/endpoint"
	"github.com/portainer/portainer-ee/api/dataservices/endpointrelation"
	"github.com/portainer/portainer-ee/api/dataservices/enforcement"
	"github.com/portainer/portainer-ee/api/dataservices/extension"
	"github.com/portainer/portainer-ee/api/dataservices/fdoprofile"
	"github.com/portainer/portainer-ee/api/dataservices/gitcredential"
	"github.com/portainer/portainer-ee/api/dataservices/helmuserrepository"
	"github.com/portainer/portainer-ee/api/dataservices/license"
	"github.com/portainer/portainer-ee/api/dataservices/podsecurity"
	"github.com/portainer/portainer-ee/api/dataservices/registry"
	"github.com/portainer/portainer-ee/api/dataservices/role"
	"github.com/portainer/portainer-ee/api/dataservices/s3backup"
	"github.com/portainer/portainer-ee/api/dataservices/schedule"
	"github.com/portainer/portainer-ee/api/dataservices/settings"
	"github.com/portainer/portainer-ee/api/dataservices/snapshot"
	"github.com/portainer/portainer-ee/api/dataservices/ssl"
	"github.com/portainer/portainer-ee/api/dataservices/stack"
	"github.com/portainer/portainer-ee/api/dataservices/teammembership"
	"github.com/portainer/portainer-ee/api/dataservices/tunnelserver"
	"github.com/portainer/portainer-ee/api/dataservices/user"
	portainer "github.com/portainer/portainer/api"
	cemodels "github.com/portainer/portainer/api/database/models"
	"github.com/portainer/portainer/api/dataservices/customtemplate"
	"github.com/portainer/portainer/api/dataservices/edgejob"
	"github.com/portainer/portainer/api/dataservices/endpointgroup"
	"github.com/portainer/portainer/api/dataservices/pendingactions"
	"github.com/portainer/portainer/api/dataservices/resourcecontrol"
	"github.com/portainer/portainer/api/dataservices/tag"
	"github.com/portainer/portainer/api/dataservices/team"
	"github.com/portainer/portainer/api/dataservices/version"
	"github.com/portainer/portainer/api/dataservices/webhook"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/encoding/json"
)

// Store defines the implementation of dataservices.DataStore using
// dataservicesDB as the storage system.
type Store struct {
	flags                     *portaineree.CLIFlags
	path                      string
	connection                portainer.Connection
	fileService               portainer.FileService
	CloudProvisioningService  *cloudprovisioning.Service
	CustomTemplateService     *customtemplate.Service
	DockerHubService          *dockerhub.Service
	EdgeAsyncCommandService   *edgeasynccommand.Service
	EdgeConfigService         *edgeconfig.Service
	EdgeConfigStateService    *edgeconfigstate.Service
	EdgeGroupService          *edgegroup.Service
	EdgeJobService            *edgejob.Service
	EdgeUpdateScheduleService *edgeupdateschedule.Service
	EdgeStackService          *edgestack.Service
	EdgeStackLogService       *edgestacklog.Service
	EndpointGroupService      *endpointgroup.Service
	EndpointService           *endpoint.Service
	EndpointRelationService   *endpointrelation.Service
	EnforcementService        *enforcement.Service
	ExtensionService          *extension.Service
	FDOProfilesService        *fdoprofile.Service
	HelmUserRepositoryService *helmuserrepository.Service
	LicenseService            *license.Service
	RegistryService           *registry.Service
	ResourceControlService    *resourcecontrol.Service
	RoleService               *role.Service
	APIKeyRepositoryService   *apikeyrepository.Service
	GitCredentialService      *gitcredential.Service
	S3BackupService           *s3backup.Service
	ScheduleService           *schedule.Service
	SettingsService           *settings.Service
	SnapshotService           *snapshot.Service
	SSLSettingsService        *ssl.Service
	StackService              *stack.Service
	PodSecurityService        *podsecurity.Service
	TagService                *tag.Service
	TeamMembershipService     *teammembership.Service
	TeamService               *team.Service
	TunnelServerService       *tunnelserver.Service
	UserService               *user.Service
	VersionService            *version.Service
	WebhookService            *webhook.Service
	CloudCredentialService    *cloudcredential.Service
	PendingActionsService     *pendingactions.Service
}

func (store *Store) initServices() error {
	authorizationsetService, err := role.NewService(store.connection)
	if err != nil {
		return err
	}
	store.RoleService = authorizationsetService

	cloudProvisioningService, err := cloudprovisioning.NewService(store.connection)
	if err != nil {
		return err
	}
	store.CloudProvisioningService = cloudProvisioningService

	customTemplateService, err := customtemplate.NewService(store.connection)
	if err != nil {
		return err
	}
	store.CustomTemplateService = customTemplateService

	dockerhubService, err := dockerhub.NewService(store.connection)
	if err != nil {
		return err
	}
	store.DockerHubService = dockerhubService

	edgeAsyncCommandService, err := edgeasynccommand.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeAsyncCommandService = edgeAsyncCommandService

	store.EdgeConfigService, err = edgeconfig.NewService(store.connection)
	if err != nil {
		return err
	}

	store.EdgeConfigStateService, err = edgeconfigstate.NewService(store.connection)
	if err != nil {
		return err
	}

	edgeUpdateScheduleService, err := edgeupdateschedule.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeUpdateScheduleService = edgeUpdateScheduleService

	endpointRelationService, err := endpointrelation.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EndpointRelationService = endpointRelationService

	edgeStackService, err := edgestack.NewService(store.connection, endpointRelationService.InvalidateEdgeCacheForEdgeStack)
	if err != nil {
		return err
	}
	store.EdgeStackService = edgeStackService
	endpointRelationService.RegisterUpdateStackFunction(edgeStackService.UpdateEdgeStackFunc, edgeStackService.UpdateEdgeStackFuncTx)

	edgeStacklogService, err := edgestacklog.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeStackLogService = edgeStacklogService

	edgeGroupService, err := edgegroup.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeGroupService = edgeGroupService

	edgeJobService, err := edgejob.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeJobService = edgeJobService

	endpointgroupService, err := endpointgroup.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EndpointGroupService = endpointgroupService

	endpointService, err := endpoint.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EndpointService = endpointService

	enforcementService, err := enforcement.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EnforcementService = enforcementService

	extensionService, err := extension.NewService(store.connection)
	if err != nil {
		return err
	}
	store.ExtensionService = extensionService

	fdoProfilesService, err := fdoprofile.NewService(store.connection)
	if err != nil {
		return err
	}
	store.FDOProfilesService = fdoProfilesService

	helmUserRepositoryService, err := helmuserrepository.NewService(store.connection)
	if err != nil {
		return err
	}
	store.HelmUserRepositoryService = helmUserRepositoryService

	licenseService, err := license.NewService(store.connection)
	if err != nil {
		return err
	}
	store.LicenseService = licenseService

	registryService, err := registry.NewService(store.connection)
	if err != nil {
		return err
	}
	store.RegistryService = registryService

	resourcecontrolService, err := resourcecontrol.NewService(store.connection)
	if err != nil {
		return err
	}
	store.ResourceControlService = resourcecontrolService

	s3backupService, err := s3backup.NewService(store.connection)
	if err != nil {
		return nil
	}
	store.S3BackupService = s3backupService

	settingsService, err := settings.NewService(store.connection)
	if err != nil {
		return err
	}
	store.SettingsService = settingsService

	snapshotService, err := snapshot.NewService(store.connection)
	if err != nil {
		return err
	}
	store.SnapshotService = snapshotService

	sslSettingsService, err := ssl.NewService(store.connection)
	if err != nil {
		return err
	}
	store.SSLSettingsService = sslSettingsService

	stackService, err := stack.NewService(store.connection)
	if err != nil {
		return err
	}
	store.StackService = stackService
	podsecurityService, err := podsecurity.NewService(store.connection)
	if err != nil {
		return err
	}
	store.PodSecurityService = podsecurityService

	tagService, err := tag.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TagService = tagService

	teammembershipService, err := teammembership.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TeamMembershipService = teammembershipService

	teamService, err := team.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TeamService = teamService

	tunnelServerService, err := tunnelserver.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TunnelServerService = tunnelServerService

	userService, err := user.NewService(store.connection)
	if err != nil {
		return err
	}
	store.UserService = userService

	apiKeyService, err := apikeyrepository.NewService(store.connection)
	if err != nil {
		return err
	}
	store.APIKeyRepositoryService = apiKeyService

	gitCredentialService, err := gitcredential.NewService(store.connection)
	if err != nil {
		return err
	}
	store.GitCredentialService = gitCredentialService

	versionService, err := version.NewService(store.connection)
	if err != nil {
		return err
	}
	store.VersionService = versionService

	webhookService, err := webhook.NewService(store.connection)
	if err != nil {
		return err
	}
	store.WebhookService = webhookService

	scheduleService, err := schedule.NewService(store.connection)
	if err != nil {
		return err
	}
	store.ScheduleService = scheduleService

	cloudCredentialService, err := cloudcredential.NewService(store.connection)
	if err != nil {
		return err
	}
	store.CloudCredentialService = cloudCredentialService

	pendingActionsService, err := pendingactions.NewService(store.connection)
	if err != nil {
		return err
	}
	store.PendingActionsService = pendingActionsService

	return nil
}

// PendingActions gives access to the PendingActions data management layer
func (store *Store) PendingActions() dataservices.PendingActionsService {
	return store.PendingActionsService
}

// CustomTemplate gives access to the CustomTemplate data management layer
func (store *Store) CustomTemplate() dataservices.CustomTemplateService {
	return store.CustomTemplateService
}

// EdgeAsyncCommand gives access to the EdgeAsyncCommand data management layer
func (store *Store) EdgeAsyncCommand() dataservices.EdgeAsyncCommandService {
	return store.EdgeAsyncCommandService
}

func (store *Store) EdgeConfig() dataservices.EdgeConfigService {
	return store.EdgeConfigService
}

func (store *Store) EdgeConfigState() dataservices.EdgeConfigStateService {
	return store.EdgeConfigStateService
}

func (store *Store) CloudProvisioning() dataservices.CloudProvisioningService {
	return store.CloudProvisioningService
}

// EdgeGroup gives access to the EdgeGroup data management layer
func (store *Store) EdgeGroup() dataservices.EdgeGroupService {
	return store.EdgeGroupService
}

// EdgeJob gives access to the EdgeJob data management layer
func (store *Store) EdgeJob() dataservices.EdgeJobService {
	return store.EdgeJobService
}

// EdgeUpdateSchedule gives access to the EdgeUpdateSchedule data management layer
func (store *Store) EdgeUpdateSchedule() dataservices.EdgeUpdateScheduleService {
	return store.EdgeUpdateScheduleService
}

// EdgeStack gives access to the EdgeStack data management layer
func (store *Store) EdgeStack() dataservices.EdgeStackService {
	return store.EdgeStackService
}

// EdgeStackLog gives access to the EdgeStackLog data management layer
func (store *Store) EdgeStackLog() dataservices.EdgeStackLogService {
	return store.EdgeStackLogService
}

// Environment(Endpoint) gives access to the Environment(Endpoint) data management layer
func (store *Store) Endpoint() dataservices.EndpointService {
	return store.EndpointService
}

// EndpointGroup gives access to the EndpointGroup data management layer
func (store *Store) EndpointGroup() dataservices.EndpointGroupService {
	return store.EndpointGroupService
}

// EndpointRelation gives access to the EndpointRelation data management layer
func (store *Store) EndpointRelation() dataservices.EndpointRelationService {
	return store.EndpointRelationService
}

func (store *Store) Enforcement() dataservices.EnforcementService {
	return store.EnforcementService
}

// FDOProfile gives access to the FDOProfile data management layer
func (store *Store) FDOProfile() dataservices.FDOProfileService {
	return store.FDOProfilesService
}

func (store *Store) HelmUserRepository() dataservices.HelmUserRepositoryService {
	return store.HelmUserRepositoryService
}

// License provides access to the License data management layer
func (store *Store) License() dataservices.LicenseRepository {
	return store.LicenseService
}

// Registry gives access to the Registry data management layer
func (store *Store) Registry() dataservices.RegistryService {
	return store.RegistryService
}

// ResourceControl gives access to the ResourceControl data management layer
func (store *Store) ResourceControl() dataservices.ResourceControlService {
	return store.ResourceControlService
}

// Role gives access to the Role data management layer
func (store *Store) Role() dataservices.RoleService {
	return store.RoleService
}

// APIKeyRepository gives access to the api-key data management layer
func (store *Store) APIKeyRepository() dataservices.APIKeyRepository {
	return store.APIKeyRepositoryService
}

// GitCredential gives access to the git-credential data management layer
func (store *Store) GitCredential() dataservices.GitCredential {
	return store.GitCredentialService
}

// S3Backup gives access to S3 backup settings and status
func (store *Store) S3Backup() dataservices.S3BackupService {
	return store.S3BackupService
}

// Settings gives access to the Settings data management layer
func (store *Store) Settings() dataservices.SettingsService {
	return store.SettingsService
}

func (store *Store) Snapshot() dataservices.SnapshotService {
	return store.SnapshotService
}

// SSLSettings gives access to the SSL Settings data management layer
func (store *Store) SSLSettings() dataservices.SSLSettingsService {
	return store.SSLSettingsService
}

// Stack gives access to the Stack data management layer
func (store *Store) Stack() dataservices.StackService {
	return store.StackService
}

// Tag gives access to the Tag data management layer
func (store *Store) Tag() dataservices.TagService {
	return store.TagService
}

// TeamMembership gives access to the TeamMembership data management layer
func (store *Store) TeamMembership() dataservices.TeamMembershipService {
	return store.TeamMembershipService
}

// Team gives access to the Team data management layer
func (store *Store) Team() dataservices.TeamService {
	return store.TeamService
}

func (store *Store) PodSecurity() dataservices.PodSecurityService {
	return store.PodSecurityService
}

// TunnelServer gives access to the TunnelServer data management layer
func (store *Store) TunnelServer() dataservices.TunnelServerService {
	return store.TunnelServerService
}

// User gives access to the User data management layer
func (store *Store) User() dataservices.UserService {
	return store.UserService
}

// Version gives access to the Version data management layer
func (store *Store) Version() dataservices.VersionService {
	return store.VersionService
}

// Webhook gives access to the Webhook data management layer
func (store *Store) Webhook() dataservices.WebhookService {
	return store.WebhookService
}

// CloudCredential gives access to the Webhook data management layer
func (store *Store) CloudCredential() dataservices.CloudCredentialService {
	return store.CloudCredentialService
}

type storeExport struct {
	ApiKey             []portaineree.APIKey             `json:"api_key,omitempty"`
	CustomTemplate     []portainer.CustomTemplate       `json:"customtemplates,omitempty"`
	EdgeStack          []portaineree.EdgeStack          `json:"edge_stack,omitempty"`
	EdgeGroup          []portaineree.EdgeGroup          `json:"edgegroups,omitempty"`
	EdgeJob            []portainer.EdgeJob              `json:"edgejobs,omitempty"`
	EndpointGroup      []portainer.EndpointGroup        `json:"endpoint_groups,omitempty"`
	EndpointRelation   []portainer.EndpointRelation     `json:"endpoint_relations,omitempty"`
	Endpoint           []portaineree.Endpoint           `json:"endpoints,omitempty"`
	Extensions         []portaineree.Extension          `json:"extension,omitempty"`
	FDOProfile         []portainer.FDOProfile           `json:"fdo_profiles,omitempty"`
	GitCredentials     []portaineree.GitCredential      `json:"git_credentials,omitempty"`
	HelmUserRepository []portaineree.HelmUserRepository `json:"helm_user_repository,omitempty"`
	License            []liblicense.PortainerLicense    `json:"license,omitempty"`
	Registry           []portaineree.Registry           `json:"registries,omitempty"`
	ResourceControl    []portainer.ResourceControl      `json:"resource_control,omitempty"`
	Role               []portaineree.Role               `json:"roles,omitempty"`
	S3BackupSettings   portaineree.S3BackupSettings     `json:"s3backup,omitempty"`
	Schedules          []portaineree.Schedule           `json:"schedules,omitempty"`
	Settings           portaineree.Settings             `json:"settings,omitempty"`
	Snapshot           []portaineree.Snapshot           `json:"snapshots,omitempty"`
	SSLSettings        portaineree.SSLSettings          `json:"ssl,omitempty"`
	Stack              []portaineree.Stack              `json:"stacks,omitempty"`
	Tag                []portainer.Tag                  `json:"tags,omitempty"`
	TeamMembership     []portainer.TeamMembership       `json:"team_membership,omitempty"`
	Team               []portainer.Team                 `json:"teams,omitempty"`
	TunnelServer       portainer.TunnelServerInfo       `json:"tunnel_server,omitempty"`
	User               []portaineree.User               `json:"users,omitempty"`
	Version            cemodels.Version                 `json:"version,omitempty"`
	Webhook            []portainer.Webhook              `json:"webhooks,omitempty"`
	Metadata           map[string]interface{}           `json:"metadata,omitempty"`
	CloudCredential    []models.CloudCredential         `json:"cloudcredential,omitempty"`
}

func (store *Store) Export(filename string) (err error) {
	backup := storeExport{}

	if c, err := store.CustomTemplate().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Custom Templates")
		}
	} else {
		backup.CustomTemplate = c
	}

	if e, err := store.EdgeGroup().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Edge Groups")
		}
	} else {
		backup.EdgeGroup = e
	}

	if e, err := store.EdgeJob().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Edge Jobs")
		}
	} else {
		backup.EdgeJob = e
	}

	if e, err := store.EdgeStack().EdgeStacks(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Edge Stacks")
		}
	} else {
		backup.EdgeStack = e
	}

	if e, err := store.Endpoint().Endpoints(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Endpoints")
		}
	} else {
		backup.Endpoint = e
	}

	if e, err := store.EndpointGroup().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Endpoint Groups")
		}
	} else {
		backup.EndpointGroup = e
	}

	if r, err := store.EndpointRelation().EndpointRelations(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Endpoint Relations")
		}
	} else {
		backup.EndpointRelation = r
	}

	if r, err := store.ExtensionService.Extensions(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Extensions")
		}
	} else {
		backup.Extensions = r
	}

	if r, err := store.FDOProfile().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("Exporting FDO Profiles")
		}
	} else {
		backup.FDOProfile = r
	}

	if r, err := store.GitCredential().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("Exporting Git Credentials")
		}
	} else {
		backup.GitCredentials = r
	}

	if r, err := store.HelmUserRepository().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Helm User Repositories")
		}
	} else {
		backup.HelmUserRepository = r
	}

	if r, err := store.LicenseService.Licenses(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Licenses")
		}
	} else {
		backup.License = r
	}

	if r, err := store.Registry().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Registries")
		}
	} else {
		backup.Registry = r
	}

	if c, err := store.ResourceControl().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Resource Controls")
		}
	} else {
		backup.ResourceControl = c
	}

	if role, err := store.Role().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Roles")
		}
	} else {
		backup.Role = role
	}

	if s3BackupSettings, err := store.S3Backup().GetSettings(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting S3 Backup Settings")
		}
	} else {
		backup.S3BackupSettings = s3BackupSettings
	}

	if r, err := store.ScheduleService.Schedules(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Schedules")
		}
	} else {
		backup.Schedules = r
	}

	if settings, err := store.Settings().Settings(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Settings")
		}
	} else {
		backup.Settings = *settings
	}

	if snapshot, err := store.Snapshot().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Snapshots")
		}
	} else {
		backup.Snapshot = snapshot
	}

	if settings, err := store.SSLSettings().Settings(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting SSL Settings")
		}
	} else {
		backup.SSLSettings = *settings
	}

	if t, err := store.Stack().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Stacks")
		}
	} else {
		backup.Stack = t
	}

	if t, err := store.Tag().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Tags")
		}
	} else {
		backup.Tag = t
	}

	if t, err := store.TeamMembership().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Team Memberships")
		}
	} else {
		backup.TeamMembership = t
	}

	if t, err := store.Team().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Teams")
		}
	} else {
		backup.Team = t
	}

	if info, err := store.TunnelServer().Info(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Tunnel Server")
		}
	} else {
		backup.TunnelServer = *info
	}

	if users, err := store.User().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Users")
		}
	} else {
		backup.User = users
	}

	if webhooks, err := store.Webhook().ReadAll(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Webhooks")
		}
	} else {
		backup.Webhook = webhooks
	}

	if version, err := store.Version().Version(); err != nil {
		if !store.IsErrObjectNotFound(err) {
			log.Error().Err(err).Msg("exporting Version")
		}
	} else {
		backup.Version = *version
	}

	backup.Metadata, err = store.connection.BackupMetadata()
	if err != nil {
		log.Error().Err(err).Msg("exporting Metadata")
	}

	b, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0o600)
}

func (store *Store) Import(filename string) (err error) {
	backup := storeExport{}

	s, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(s), &backup)
	if err != nil {
		return err
	}

	store.Version().UpdateVersion(&backup.Version)
	for _, v := range backup.CustomTemplate {
		store.CustomTemplate().Update(v.ID, &v)
	}

	for _, v := range backup.EdgeGroup {
		store.EdgeGroup().Update(v.ID, &v)
	}

	for _, v := range backup.EdgeJob {
		store.EdgeJob().Update(v.ID, &v)
	}

	for _, v := range backup.EdgeStack {
		store.EdgeStack().UpdateEdgeStack(v.ID, &v, true)
	}

	for _, v := range backup.Endpoint {
		store.Endpoint().UpdateEndpoint(v.ID, &v)
	}

	for _, v := range backup.EndpointGroup {
		store.EndpointGroup().Update(v.ID, &v)
	}

	for _, v := range backup.EndpointRelation {
		store.EndpointRelation().UpdateEndpointRelation(v.EndpointID, &v)
	}

	for _, v := range backup.FDOProfile {
		store.FDOProfile().Update(v.ID, &v)
	}

	for _, v := range backup.GitCredentials {
		store.GitCredential().Update(v.ID, &v)
	}

	for _, v := range backup.HelmUserRepository {
		store.HelmUserRepository().Update(v.ID, &v)
	}

	for _, v := range backup.License {
		store.LicenseService.AddLicense(v.ID, &v)
	}

	for _, v := range backup.Registry {
		store.Registry().Update(v.ID, &v)
	}

	for _, v := range backup.ResourceControl {
		store.ResourceControl().Update(v.ID, &v)
	}

	for _, v := range backup.Role {
		store.Role().Update(v.ID, &v)
	}

	store.S3Backup().UpdateSettings(backup.S3BackupSettings)
	store.Settings().UpdateSettings(&backup.Settings)
	store.SSLSettings().UpdateSettings(&backup.SSLSettings)

	for _, v := range backup.Snapshot {
		store.Snapshot().Update(v.EndpointID, &v)
	}

	for _, v := range backup.Stack {
		store.Stack().Update(v.ID, &v)
	}

	for _, v := range backup.Tag {
		store.Tag().Update(v.ID, &v)
	}

	for _, v := range backup.TeamMembership {
		store.TeamMembership().Update(v.ID, &v)
	}

	for _, v := range backup.Team {
		store.Team().Update(v.ID, &v)
	}

	store.TunnelServer().UpdateInfo(&backup.TunnelServer)

	for _, user := range backup.User {
		if err := store.User().Update(user.ID, &user); err != nil {
			log.Debug().Str("user", fmt.Sprintf("%+v", user)).Err(err).Msg("failed to update the user in the database")
		}
	}

	for _, v := range backup.Webhook {
		store.Webhook().Update(v.ID, &v)
	}

	return store.connection.RestoreMetadata(backup.Metadata)
}

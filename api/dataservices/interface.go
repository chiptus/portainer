package dataservices

import (
	"io"
	"time"

	"github.com/portainer/liblicense"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	portainer "github.com/portainer/portainer/api"
)

type (
	DataStoreTx interface {
		IsErrObjectNotFound(err error) bool
		CloudProvisioning() CloudProvisioningService
		CustomTemplate() CustomTemplateService
		EdgeAsyncCommand() EdgeAsyncCommandService
		EdgeGroup() EdgeGroupService
		EdgeJob() EdgeJobService
		EdgeStack() EdgeStackService
		EdgeStackLog() EdgeStackLogService
		EdgeUpdateSchedule() EdgeUpdateScheduleService
		Endpoint() EndpointService
		EndpointGroup() EndpointGroupService
		EndpointRelation() EndpointRelationService
		Enforcement() EnforcementService
		FDOProfile() FDOProfileService
		HelmUserRepository() HelmUserRepositoryService
		License() LicenseRepository
		Registry() RegistryService
		ResourceControl() ResourceControlService
		Role() RoleService
		APIKeyRepository() APIKeyRepository
		GitCredential() GitCredential
		S3Backup() S3BackupService
		Settings() SettingsService
		Snapshot() SnapshotService
		SSLSettings() SSLSettingsService
		Stack() StackService
		Tag() TagService
		TeamMembership() TeamMembershipService
		Team() TeamService
		TunnelServer() TunnelServerService
		User() UserService
		PodSecurity() PodSecurityService
		Version() VersionService
		Webhook() WebhookService
		CloudCredential() CloudCredentialService
	}

	DataStore interface {
		Connection() portainer.Connection
		Open() (newStore bool, err error)
		Init() error
		Close() error
		UpdateTx(func(DataStoreTx) error) error
		ViewTx(func(DataStoreTx) error) error
		MigrateData() error
		Rollback(force bool) error
		RollbackToCE() error
		CheckCurrentEdition() error
		BackupTo(w io.Writer) error
		Export(filename string) (err error)

		DataStoreTx
	}

	// CustomTemplateService represents a service to manage custom templates
	CustomTemplateService interface {
		BaseCRUD[portaineree.CustomTemplate, portaineree.CustomTemplateID]
		GetNextIdentifier() int
	}

	// EdgeAsyncCommandService represents a service to manage EdgeAsyncCommands
	EdgeAsyncCommandService interface {
		Create(command *portaineree.EdgeAsyncCommand) error
		Update(id int, command *portaineree.EdgeAsyncCommand) error
		EndpointCommands(endpointID portaineree.EndpointID) ([]portaineree.EdgeAsyncCommand, error)
	}

	// CloudProvisioningService
	CloudProvisioningService interface {
		BaseCRUD[portaineree.CloudProvisioningTask, portaineree.CloudProvisioningTaskID]
		GetNextIdentifier() int
	}

	CloudCredentialService interface {
		BaseCRUD[models.CloudCredential, models.CloudCredentialID]
	}

	// EdgeGroupService represents a service to manage Edge groups
	EdgeGroupService interface {
		BaseCRUD[portaineree.EdgeGroup, portaineree.EdgeGroupID]
		UpdateEdgeGroupFunc(ID portaineree.EdgeGroupID, updateFunc func(group *portaineree.EdgeGroup)) error
	}

	// EdgeJobService represents a service to manage Edge jobs
	EdgeJobService interface {
		BaseCRUD[portaineree.EdgeJob, portaineree.EdgeJobID]
		CreateWithID(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error
		UpdateEdgeJobFunc(ID portaineree.EdgeJobID, updateFunc func(edgeJob *portaineree.EdgeJob)) error
		GetNextIdentifier() int
	}

	EdgeUpdateScheduleService interface {
		List() ([]edgetypes.UpdateSchedule, error)
		Item(ID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error)
		Create(edgeUpdateSchedule *edgetypes.UpdateSchedule) error
		Update(ID edgetypes.UpdateScheduleID, edgeUpdateSchedule *edgetypes.UpdateSchedule) error
		Delete(ID edgetypes.UpdateScheduleID) error
		BucketName() string
	}

	// EdgeStackService represents a service to manage Edge stacks
	EdgeStackService interface {
		EdgeStacks() ([]portaineree.EdgeStack, error)
		EdgeStack(ID portaineree.EdgeStackID) (*portaineree.EdgeStack, error)
		EdgeStackVersion(ID portaineree.EdgeStackID) (int, bool)
		Create(id portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error
		UpdateEdgeStack(ID portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error
		UpdateEdgeStackFunc(ID portaineree.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error
		DeleteEdgeStack(ID portaineree.EdgeStackID) error
		GetNextIdentifier() int
		BucketName() string
	}

	EdgeStackLogService interface {
		Create(edgeStack *portaineree.EdgeStackLog) error
		Update(edgeStack *portaineree.EdgeStackLog) error
		Delete(edgeStackID portaineree.EdgeStackID, endpointID portaineree.EndpointID) error
		EdgeStackLog(edgeStackID portaineree.EdgeStackID, endpointID portaineree.EndpointID) (*portaineree.EdgeStackLog, error)
	}

	// EndpointService represents a service for managing environment(endpoint) data
	EndpointService interface {
		Endpoint(ID portaineree.EndpointID) (*portaineree.Endpoint, error)
		EndpointIDByEdgeID(edgeID string) (portaineree.EndpointID, bool)
		Heartbeat(endpointID portaineree.EndpointID) (int64, bool)
		UpdateHeartbeat(endpointID portaineree.EndpointID)
		Endpoints() ([]portaineree.Endpoint, error)
		Create(endpoint *portaineree.Endpoint) error
		UpdateEndpoint(ID portaineree.EndpointID, endpoint *portaineree.Endpoint) error
		DeleteEndpoint(ID portaineree.EndpointID) error
		GetNextIdentifier() int
		BucketName() string
		SetMessage(ID portaineree.EndpointID, statusMessage portaineree.EndpointStatusMessage) error
	}

	// EndpointGroupService represents a service for managing environment(endpoint) group data
	EndpointGroupService interface {
		BaseCRUD[portaineree.EndpointGroup, portaineree.EndpointGroupID]
	}

	// EndpointRelationService represents a service for managing environment(endpoint) relations data
	EndpointRelationService interface {
		EndpointRelations() ([]portaineree.EndpointRelation, error)
		EndpointRelation(EndpointID portaineree.EndpointID) (*portaineree.EndpointRelation, error)
		Create(endpointRelation *portaineree.EndpointRelation) error
		UpdateEndpointRelation(EndpointID portaineree.EndpointID, endpointRelation *portaineree.EndpointRelation) error
		DeleteEndpointRelation(EndpointID portaineree.EndpointID) error
		BucketName() string
	}

	// Service manages license enforcement record
	EnforcementService interface {
		LicenseEnforcement() (*models.LicenseEnforcement, error)
		UpdateOveruseStartedTimestamp(timestamp int64) error
	}

	// FDOProfileService represents a service to manage FDO Profiles
	FDOProfileService interface {
		BaseCRUD[portaineree.FDOProfile, portaineree.FDOProfileID]
		GetNextIdentifier() int
	}

	// HelmUserRepositoryService represents a service to manage HelmUserRepositories
	HelmUserRepositoryService interface {
		BaseCRUD[portaineree.HelmUserRepository, portaineree.HelmUserRepositoryID]
		HelmUserRepositoryByUserID(userID portaineree.UserID) ([]portaineree.HelmUserRepository, error)
	}

	// JWTService represents a service for managing JWT tokens
	JWTService interface {
		GenerateToken(data *portaineree.TokenData) (string, error)
		GenerateTokenForOAuth(data *portaineree.TokenData, expiryTime *time.Time) (string, error)
		GenerateTokenForKubeconfig(data *portaineree.TokenData) (string, error)
		ParseAndVerifyToken(token string) (*portaineree.TokenData, error)
		SetUserSessionDuration(userSessionDuration time.Duration)
	}

	// LicenseRepository represents a service used to manage licenses store
	LicenseRepository interface {
		Licenses() ([]liblicense.PortainerLicense, error)
		License(licenseKey string) (*liblicense.PortainerLicense, error)
		AddLicense(licenseKey string, license *liblicense.PortainerLicense) error
		UpdateLicense(licenseKey string, license *liblicense.PortainerLicense) error
		DeleteLicense(licenseKey string) error
	}

	// RegistryService represents a service for managing registry data
	RegistryService interface {
		BaseCRUD[portaineree.Registry, portaineree.RegistryID]
	}

	// ResourceControlService represents a service for managing resource control data
	ResourceControlService interface {
		BaseCRUD[portaineree.ResourceControl, portaineree.ResourceControlID]
		ResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType) (*portaineree.ResourceControl, error)
	}

	// RoleService represents a service for managing user roles
	RoleService interface {
		BaseCRUD[portaineree.Role, portaineree.RoleID]
	}

	// APIKeyRepository
	APIKeyRepository interface {
		BaseCRUD[portaineree.APIKey, portaineree.APIKeyID]
		GetAPIKeysByUserID(userID portaineree.UserID) ([]portaineree.APIKey, error)
		GetAPIKeyByDigest(digest []byte) (*portaineree.APIKey, error)
	}

	// GitCredential
	GitCredential interface {
		BaseCRUD[portaineree.GitCredential, portaineree.GitCredentialID]
		GetGitCredentialsByUserID(userID portaineree.UserID) ([]portaineree.GitCredential, error)
		GetGitCredentialByName(userID portaineree.UserID, name string) (*portaineree.GitCredential, error)
	}

	// S3BackupService represents a storage service for managing S3 backup settings and status
	S3BackupService interface {
		GetStatus() (portaineree.S3BackupStatus, error)
		DropStatus() error
		UpdateStatus(status portaineree.S3BackupStatus) error
		UpdateSettings(settings portaineree.S3BackupSettings) error
		GetSettings() (portaineree.S3BackupSettings, error)
	}

	// SettingsService represents a service for managing application settings
	SettingsService interface {
		Settings() (*portaineree.Settings, error)
		UpdateSettings(settings *portaineree.Settings) error
		BucketName() string
	}

	SnapshotService interface {
		BaseCRUD[portaineree.Snapshot, portaineree.EndpointID]
	}

	// SSLSettingsService represents a service for managing application settings
	SSLSettingsService interface {
		Settings() (*portaineree.SSLSettings, error)
		UpdateSettings(settings *portaineree.SSLSettings) error
		BucketName() string
	}

	// StackService represents a service for managing stack data
	StackService interface {
		BaseCRUD[portaineree.Stack, portaineree.StackID]
		StackByName(name string) (*portaineree.Stack, error)
		StacksByName(name string) ([]portaineree.Stack, error)
		GetNextIdentifier() int
		StackByWebhookID(ID string) (*portaineree.Stack, error)
		RefreshableStacks() ([]portaineree.Stack, error)
	}

	PodSecurityService interface {
		PodSecurityRule(ID podsecurity.PodSecurityRuleID) (*podsecurity.PodSecurityRule, error)
		PodSecurityByEndpointID(endpointID int) (*podsecurity.PodSecurityRule, error)
		Create(podsecurity *podsecurity.PodSecurityRule) error
		UpdatePodSecurityRule(endpointID int, podsecurity *podsecurity.PodSecurityRule) error
		DeletePodSecurityRule(endpointID int) error
		GetNextIdentifier() int
	}

	// TagService represents a service for managing tag data
	TagService interface {
		BaseCRUD[portaineree.Tag, portaineree.TagID]
		UpdateTagFunc(ID portaineree.TagID, updateFunc func(tag *portaineree.Tag)) error
	}

	// TeamService represents a service for managing user data
	TeamService interface {
		BaseCRUD[portaineree.Team, portaineree.TeamID]
		TeamByName(name string) (*portaineree.Team, error)
	}

	// TeamMembershipService represents a service for managing team membership data
	TeamMembershipService interface {
		BaseCRUD[portaineree.TeamMembership, portaineree.TeamMembershipID]
		TeamMembershipsByUserID(userID portaineree.UserID) ([]portaineree.TeamMembership, error)
		TeamMembershipsByTeamID(teamID portaineree.TeamID) ([]portaineree.TeamMembership, error)
		DeleteTeamMembershipByUserID(userID portaineree.UserID) error
		DeleteTeamMembershipByTeamID(teamID portaineree.TeamID) error
		DeleteTeamMembershipByTeamIDAndUserID(teamID portaineree.TeamID, userID portaineree.UserID) error
	}

	// TunnelServerService represents a service for managing data associated to the tunnel server
	TunnelServerService interface {
		Info() (*portaineree.TunnelServerInfo, error)
		UpdateInfo(info *portaineree.TunnelServerInfo) error
		BucketName() string
	}

	// UserService represents a service for managing user data
	UserService interface {
		BaseCRUD[portaineree.User, portaineree.UserID]
		UserByUsername(username string) (*portaineree.User, error)
		UsersByRole(role portaineree.UserRole) ([]portaineree.User, error)
	}

	// VersionService represents a service for managing version data
	VersionService interface {
		InstanceID() (string, error)
		UpdateInstanceID(ID string) error
		Edition() (portaineree.SoftwareEdition, error)
		Version() (*models.Version, error)
		UpdateVersion(*models.Version) error
	}

	// WebhookService represents a service for managing webhook data.
	WebhookService interface {
		BaseCRUD[portaineree.Webhook, portaineree.WebhookID]
		WebhookByResourceID(resourceID string) (*portaineree.Webhook, error)
		WebhookByToken(token string) (*portaineree.Webhook, error)
	}
)

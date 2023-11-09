package dataservices

import (
	"io"
	"time"

	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	portainer "github.com/portainer/portainer/api"
	cemodels "github.com/portainer/portainer/api/database/models"
)

type (
	DataStoreTx interface {
		IsErrObjectNotFound(err error) bool
		CloudProvisioning() CloudProvisioningService
		CustomTemplate() CustomTemplateService
		EdgeAsyncCommand() EdgeAsyncCommandService
		EdgeConfig() EdgeConfigService
		EdgeConfigState() EdgeConfigStateService
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
		PendingActions() PendingActionsService
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
		BaseCRUD[portainer.CustomTemplate, portainer.CustomTemplateID]
		GetNextIdentifier() int
	}

	PendingActionsService interface {
		BaseCRUD[portainer.PendingActions, portainer.PendingActionsID]
		GetNextIdentifier() int
	}

	// EdgeAsyncCommandService represents a service to manage EdgeAsyncCommands
	EdgeAsyncCommandService interface {
		Create(command *portaineree.EdgeAsyncCommand) error
		Update(id int, command *portaineree.EdgeAsyncCommand) error
		EndpointCommands(endpointID portainer.EndpointID) ([]portaineree.EdgeAsyncCommand, error)
	}

	// CloudProvisioningService
	CloudProvisioningService interface {
		BaseCRUD[portaineree.CloudProvisioningTask, portaineree.CloudProvisioningTaskID]
		GetNextIdentifier() int
	}

	CloudCredentialService interface {
		BaseCRUD[models.CloudCredential, models.CloudCredentialID]
	}

	EdgeConfigService interface {
		BaseCRUD[portaineree.EdgeConfig, portaineree.EdgeConfigID]
	}

	EdgeConfigStateService interface {
		BaseCRUD[portaineree.EdgeConfigState, portainer.EndpointID]
	}

	// EdgeGroupService represents a service to manage Edge groups
	EdgeGroupService interface {
		BaseCRUD[portaineree.EdgeGroup, portainer.EdgeGroupID]
		UpdateEdgeGroupFunc(ID portainer.EdgeGroupID, updateFunc func(group *portaineree.EdgeGroup)) error
	}

	// EdgeJobService represents a service to manage Edge jobs
	EdgeJobService interface {
		BaseCRUD[portainer.EdgeJob, portainer.EdgeJobID]
		CreateWithID(ID portainer.EdgeJobID, edgeJob *portainer.EdgeJob) error
		UpdateEdgeJobFunc(ID portainer.EdgeJobID, updateFunc func(edgeJob *portainer.EdgeJob)) error
		GetNextIdentifier() int
	}

	EdgeUpdateScheduleService interface {
		BaseCRUD[edgetypes.UpdateSchedule, edgetypes.UpdateScheduleID]
	}

	// EdgeStackService represents a service to manage Edge stacks
	EdgeStackService interface {
		EdgeStacks() ([]portaineree.EdgeStack, error)
		EdgeStack(ID portainer.EdgeStackID) (*portaineree.EdgeStack, error)
		EdgeStackVersion(ID portainer.EdgeStackID) (int, bool)
		Create(id portainer.EdgeStackID, edgeStack *portaineree.EdgeStack) error
		UpdateEdgeStack(ID portainer.EdgeStackID, edgeStack *portaineree.EdgeStack, cleanupCache bool) error
		UpdateEdgeStackFunc(ID portainer.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error
		DeleteEdgeStack(ID portainer.EdgeStackID) error
		GetNextIdentifier() int
		BucketName() string
	}

	EdgeStackLogService interface {
		Create(edgeStack *portaineree.EdgeStackLog) error
		Update(edgeStack *portaineree.EdgeStackLog) error
		Delete(edgeStackID portainer.EdgeStackID, endpointID portainer.EndpointID) error
		EdgeStackLog(edgeStackID portainer.EdgeStackID, endpointID portainer.EndpointID) (*portaineree.EdgeStackLog, error)
	}

	// EndpointService represents a service for managing environment(endpoint) data
	EndpointService interface {
		Endpoint(ID portainer.EndpointID) (*portaineree.Endpoint, error)
		EndpointIDByEdgeID(edgeID string) (portainer.EndpointID, bool)
		EndpointsByTeamID(teamID portainer.TeamID) ([]portaineree.Endpoint, error)
		Heartbeat(endpointID portainer.EndpointID) (int64, bool)
		UpdateHeartbeat(endpointID portainer.EndpointID)
		Endpoints() ([]portaineree.Endpoint, error)
		Create(endpoint *portaineree.Endpoint) error
		UpdateEndpoint(ID portainer.EndpointID, endpoint *portaineree.Endpoint) error
		DeleteEndpoint(ID portainer.EndpointID) error
		GetNextIdentifier() int
		BucketName() string
		SetMessage(ID portainer.EndpointID, statusMessage portaineree.EndpointStatusMessage) error
	}

	// EndpointGroupService represents a service for managing environment(endpoint) group data
	EndpointGroupService interface {
		BaseCRUD[portainer.EndpointGroup, portainer.EndpointGroupID]
	}

	// EndpointRelationService represents a service for managing environment(endpoint) relations data
	EndpointRelationService interface {
		EndpointRelations() ([]portainer.EndpointRelation, error)
		EndpointRelation(EndpointID portainer.EndpointID) (*portainer.EndpointRelation, error)
		Create(endpointRelation *portainer.EndpointRelation) error
		UpdateEndpointRelation(EndpointID portainer.EndpointID, endpointRelation *portainer.EndpointRelation) error
		DeleteEndpointRelation(EndpointID portainer.EndpointID) error
		BucketName() string
	}

	// Service manages license enforcement record
	EnforcementService interface {
		LicenseEnforcement() (*models.LicenseEnforcement, error)
		UpdateOveruseStartedTimestamp(timestamp int64) error
	}

	// FDOProfileService represents a service to manage FDO Profiles
	FDOProfileService interface {
		BaseCRUD[portainer.FDOProfile, portainer.FDOProfileID]
		GetNextIdentifier() int
	}

	// HelmUserRepositoryService represents a service to manage HelmUserRepositories
	HelmUserRepositoryService interface {
		BaseCRUD[portaineree.HelmUserRepository, portainer.HelmUserRepositoryID]
		HelmUserRepositoryByUserID(userID portainer.UserID) ([]portaineree.HelmUserRepository, error)
	}

	// JWTService represents a service for managing JWT tokens
	JWTService interface {
		GenerateToken(data *portainer.TokenData) (string, error)
		GenerateTokenForOAuth(data *portainer.TokenData, expiryTime *time.Time) (string, error)
		GenerateTokenForKubeconfig(data *portainer.TokenData) (string, error)
		ParseAndVerifyToken(token string) (*portainer.TokenData, error)
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
		BaseCRUD[portaineree.Registry, portainer.RegistryID]
	}

	// ResourceControlService represents a service for managing resource control data
	ResourceControlService interface {
		BaseCRUD[portainer.ResourceControl, portainer.ResourceControlID]
		ResourceControlByResourceIDAndType(resourceID string, resourceType portainer.ResourceControlType) (*portainer.ResourceControl, error)
	}

	// RoleService represents a service for managing user roles
	RoleService interface {
		BaseCRUD[portaineree.Role, portainer.RoleID]
	}

	// APIKeyRepository
	APIKeyRepository interface {
		BaseCRUD[portaineree.APIKey, portainer.APIKeyID]
		GetAPIKeysByUserID(userID portainer.UserID) ([]portaineree.APIKey, error)
		GetAPIKeyByDigest(digest []byte) (*portaineree.APIKey, error)
	}

	// GitCredential
	GitCredential interface {
		BaseCRUD[portaineree.GitCredential, portaineree.GitCredentialID]
		GetGitCredentialsByUserID(userID portainer.UserID) ([]portaineree.GitCredential, error)
		GetGitCredentialByName(userID portainer.UserID, name string) (*portaineree.GitCredential, error)
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
		BaseCRUD[portaineree.Snapshot, portainer.EndpointID]
	}

	// SSLSettingsService represents a service for managing application settings
	SSLSettingsService interface {
		Settings() (*portaineree.SSLSettings, error)
		UpdateSettings(settings *portaineree.SSLSettings) error
		BucketName() string
	}

	// StackService represents a service for managing stack data
	StackService interface {
		BaseCRUD[portaineree.Stack, portainer.StackID]
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
		BaseCRUD[portainer.Tag, portainer.TagID]
		UpdateTagFunc(ID portainer.TagID, updateFunc func(tag *portainer.Tag)) error
	}

	// TeamService represents a service for managing user data
	TeamService interface {
		BaseCRUD[portainer.Team, portainer.TeamID]
		TeamByName(name string) (*portainer.Team, error)
	}

	// TeamMembershipService represents a service for managing team membership data
	TeamMembershipService interface {
		BaseCRUD[portainer.TeamMembership, portainer.TeamMembershipID]
		TeamMembershipsByUserID(userID portainer.UserID) ([]portainer.TeamMembership, error)
		TeamMembershipsByTeamID(teamID portainer.TeamID) ([]portainer.TeamMembership, error)
		DeleteTeamMembershipByUserID(userID portainer.UserID) error
		DeleteTeamMembershipByTeamID(teamID portainer.TeamID) error
		DeleteTeamMembershipByTeamIDAndUserID(teamID portainer.TeamID, userID portainer.UserID) error
	}

	// TunnelServerService represents a service for managing data associated to the tunnel server
	TunnelServerService interface {
		Info() (*portainer.TunnelServerInfo, error)
		UpdateInfo(info *portainer.TunnelServerInfo) error
		BucketName() string
	}

	// UserService represents a service for managing user data
	UserService interface {
		BaseCRUD[portaineree.User, portainer.UserID]
		UserByUsername(username string) (*portaineree.User, error)
		UsersByRole(role portainer.UserRole) ([]portaineree.User, error)
	}

	// VersionService represents a service for managing version data
	VersionService interface {
		InstanceID() (string, error)
		UpdateInstanceID(ID string) error
		Edition() (portainer.SoftwareEdition, error)
		Version() (*cemodels.Version, error)
		UpdateVersion(*cemodels.Version) error
	}

	// WebhookService represents a service for managing webhook data.
	WebhookService interface {
		BaseCRUD[portainer.Webhook, portainer.WebhookID]
		WebhookByResourceID(resourceID string) (*portainer.Webhook, error)
		WebhookByToken(token string) (*portainer.Webhook, error)
	}
)

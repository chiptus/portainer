package dataservices

import (
	"io"
	"time"

	"github.com/portainer/liblicense"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices/errors"
)

type (
	// DataStore defines the interface to manage the data
	DataStore interface {
		Open() (newStore bool, err error)
		Init() error
		Close() error
		MigrateData() error
		CheckCurrentEdition() error
		Rollback(force bool) error
		RollbackToCE() error

		BackupTo(w io.Writer) error
		Export(filename string) (err error)
		IsErrObjectNotFound(err error) bool
		Connection() portainer.Connection

		CustomTemplate() CustomTemplateService
		EdgeGroup() EdgeGroupService
		EdgeJob() EdgeJobService
		EdgeStack() EdgeStackService
		Endpoint() EndpointService
		EndpointGroup() EndpointGroupService
		EndpointRelation() EndpointRelationService
		FDOProfile() FDOProfileService
		HelmUserRepository() HelmUserRepositoryService
		License() LicenseRepository
		Registry() RegistryService
		ResourceControl() ResourceControlService
		Role() RoleService
		APIKeyRepository() APIKeyRepository
		S3Backup() S3BackupService
		SSLSettings() SSLSettingsService
		Settings() SettingsService
		Stack() StackService
		Tag() TagService
		TeamMembership() TeamMembershipService
		Team() TeamService
		TunnelServer() TunnelServerService
		User() UserService
		Version() VersionService
		Webhook() WebhookService
	}

	// CustomTemplateService represents a service to manage custom templates
	CustomTemplateService interface {
		GetNextIdentifier() int
		CustomTemplates() ([]portaineree.CustomTemplate, error)
		CustomTemplate(ID portaineree.CustomTemplateID) (*portaineree.CustomTemplate, error)
		Create(customTemplate *portaineree.CustomTemplate) error
		UpdateCustomTemplate(ID portaineree.CustomTemplateID, customTemplate *portaineree.CustomTemplate) error
		DeleteCustomTemplate(ID portaineree.CustomTemplateID) error
		BucketName() string
	}

	// EdgeGroupService represents a service to manage Edge groups
	EdgeGroupService interface {
		EdgeGroups() ([]portaineree.EdgeGroup, error)
		EdgeGroup(ID portaineree.EdgeGroupID) (*portaineree.EdgeGroup, error)
		Create(group *portaineree.EdgeGroup) error
		UpdateEdgeGroup(ID portaineree.EdgeGroupID, group *portaineree.EdgeGroup) error
		DeleteEdgeGroup(ID portaineree.EdgeGroupID) error
		BucketName() string
	}

	// EdgeJobService represents a service to manage Edge jobs
	EdgeJobService interface {
		EdgeJobs() ([]portaineree.EdgeJob, error)
		EdgeJob(ID portaineree.EdgeJobID) (*portaineree.EdgeJob, error)
		Create(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error
		UpdateEdgeJob(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error
		DeleteEdgeJob(ID portaineree.EdgeJobID) error
		GetNextIdentifier() int
		BucketName() string
	}

	// EdgeStackService represents a service to manage Edge stacks
	EdgeStackService interface {
		EdgeStacks() ([]portaineree.EdgeStack, error)
		EdgeStack(ID portaineree.EdgeStackID) (*portaineree.EdgeStack, error)
		Create(id portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error
		UpdateEdgeStack(ID portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error
		DeleteEdgeStack(ID portaineree.EdgeStackID) error
		GetNextIdentifier() int
		BucketName() string
	}

	// EndpointService represents a service for managing environment(endpoint) data
	EndpointService interface {
		Endpoint(ID portaineree.EndpointID) (*portaineree.Endpoint, error)
		Endpoints() ([]portaineree.Endpoint, error)
		Create(endpoint *portaineree.Endpoint) error
		UpdateEndpoint(ID portaineree.EndpointID, endpoint *portaineree.Endpoint) error
		DeleteEndpoint(ID portaineree.EndpointID) error
		GetNextIdentifier() int
		BucketName() string
	}

	// EndpointGroupService represents a service for managing environment(endpoint) group data
	EndpointGroupService interface {
		EndpointGroup(ID portaineree.EndpointGroupID) (*portaineree.EndpointGroup, error)
		EndpointGroups() ([]portaineree.EndpointGroup, error)
		Create(group *portaineree.EndpointGroup) error
		UpdateEndpointGroup(ID portaineree.EndpointGroupID, group *portaineree.EndpointGroup) error
		DeleteEndpointGroup(ID portaineree.EndpointGroupID) error
		BucketName() string
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

	// FDOProfileService represents a service to manage FDO Profiles
	FDOProfileService interface {
		FDOProfiles() ([]portaineree.FDOProfile, error)
		FDOProfile(ID portaineree.FDOProfileID) (*portaineree.FDOProfile, error)
		Create(FDOProfile *portaineree.FDOProfile) error
		Update(ID portaineree.FDOProfileID, FDOProfile *portaineree.FDOProfile) error
		Delete(ID portaineree.FDOProfileID) error
		GetNextIdentifier() int
		BucketName() string
	}

	// HelmUserRepositoryService represents a service to manage HelmUserRepositories
	HelmUserRepositoryService interface {
		HelmUserRepositories() ([]portaineree.HelmUserRepository, error)
		HelmUserRepositoryByUserID(userID portaineree.UserID) ([]portaineree.HelmUserRepository, error)
		Create(record *portaineree.HelmUserRepository) error
		UpdateHelmUserRepository(ID portaineree.HelmUserRepositoryID, repository *portaineree.HelmUserRepository) error
		DeleteHelmUserRepository(ID portaineree.HelmUserRepositoryID) error
		BucketName() string
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
		Registry(ID portaineree.RegistryID) (*portaineree.Registry, error)
		Registries() ([]portaineree.Registry, error)
		Create(registry *portaineree.Registry) error
		UpdateRegistry(ID portaineree.RegistryID, registry *portaineree.Registry) error
		DeleteRegistry(ID portaineree.RegistryID) error
		BucketName() string
	}

	// ResourceControlService represents a service for managing resource control data
	ResourceControlService interface {
		ResourceControl(ID portaineree.ResourceControlID) (*portaineree.ResourceControl, error)
		ResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType) (*portaineree.ResourceControl, error)
		ResourceControls() ([]portaineree.ResourceControl, error)
		Create(rc *portaineree.ResourceControl) error
		UpdateResourceControl(ID portaineree.ResourceControlID, resourceControl *portaineree.ResourceControl) error
		DeleteResourceControl(ID portaineree.ResourceControlID) error
		BucketName() string
	}

	// RoleService represents a service for managing user roles
	RoleService interface {
		Role(ID portaineree.RoleID) (*portaineree.Role, error)
		Roles() ([]portaineree.Role, error)
		Create(role *portaineree.Role) error
		UpdateRole(ID portaineree.RoleID, role *portaineree.Role) error
		BucketName() string
	}

	// APIKeyRepository
	APIKeyRepository interface {
		Create(key *portaineree.APIKey) error
		GetAPIKey(keyID portaineree.APIKeyID) (*portaineree.APIKey, error)
		UpdateAPIKey(key *portaineree.APIKey) error
		DeleteAPIKey(ID portaineree.APIKeyID) error
		GetAPIKeysByUserID(userID portaineree.UserID) ([]portaineree.APIKey, error)
		GetAPIKeyByDigest(digest []byte) (*portaineree.APIKey, error)
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
		IsFeatureFlagEnabled(feature portaineree.Feature) bool
		BucketName() string
	}

	// SSLSettingsService represents a service for managing application settings
	SSLSettingsService interface {
		Settings() (*portainer.SSLSettings, error)
		UpdateSettings(settings *portainer.SSLSettings) error
		BucketName() string
	}

	// StackService represents a service for managing stack data
	StackService interface {
		Stack(ID portaineree.StackID) (*portaineree.Stack, error)
		StackByName(name string) (*portaineree.Stack, error)
		StacksByName(name string) ([]portaineree.Stack, error)
		Stacks() ([]portaineree.Stack, error)
		Create(stack *portaineree.Stack) error
		UpdateStack(ID portaineree.StackID, stack *portaineree.Stack) error
		DeleteStack(ID portaineree.StackID) error
		GetNextIdentifier() int
		StackByWebhookID(ID string) (*portaineree.Stack, error)
		RefreshableStacks() ([]portaineree.Stack, error)
		BucketName() string
	}

	// TagService represents a service for managing tag data
	TagService interface {
		Tags() ([]portaineree.Tag, error)
		Tag(ID portaineree.TagID) (*portaineree.Tag, error)
		Create(tag *portaineree.Tag) error
		UpdateTag(ID portaineree.TagID, tag *portaineree.Tag) error
		DeleteTag(ID portaineree.TagID) error
	}

	// TeamService represents a service for managing user data
	TeamService interface {
		Team(ID portaineree.TeamID) (*portaineree.Team, error)
		TeamByName(name string) (*portaineree.Team, error)
		Teams() ([]portaineree.Team, error)
		Create(team *portaineree.Team) error
		UpdateTeam(ID portaineree.TeamID, team *portaineree.Team) error
		DeleteTeam(ID portaineree.TeamID) error
		BucketName() string
	}

	// TeamMembershipService represents a service for managing team membership data
	TeamMembershipService interface {
		TeamMembership(ID portaineree.TeamMembershipID) (*portaineree.TeamMembership, error)
		TeamMemberships() ([]portaineree.TeamMembership, error)
		TeamMembershipsByUserID(userID portaineree.UserID) ([]portaineree.TeamMembership, error)
		TeamMembershipsByTeamID(teamID portaineree.TeamID) ([]portaineree.TeamMembership, error)
		Create(membership *portaineree.TeamMembership) error
		UpdateTeamMembership(ID portaineree.TeamMembershipID, membership *portaineree.TeamMembership) error
		DeleteTeamMembership(ID portaineree.TeamMembershipID) error
		DeleteTeamMembershipByUserID(userID portaineree.UserID) error
		DeleteTeamMembershipByTeamID(teamID portaineree.TeamID) error
		BucketName() string
	}

	// TunnelServerService represents a service for managing data associated to the tunnel server
	TunnelServerService interface {
		Info() (*portaineree.TunnelServerInfo, error)
		UpdateInfo(info *portaineree.TunnelServerInfo) error
		BucketName() string
	}

	// UserService represents a service for managing user data
	UserService interface {
		User(ID portaineree.UserID) (*portaineree.User, error)
		UserByUsername(username string) (*portaineree.User, error)
		Users() ([]portaineree.User, error)
		UsersByRole(role portaineree.UserRole) ([]portaineree.User, error)
		Create(user *portaineree.User) error
		UpdateUser(ID portaineree.UserID, user *portaineree.User) error
		DeleteUser(ID portaineree.UserID) error
		BucketName() string
	}

	// VersionService represents a service for managing version data
	VersionService interface {
		DBVersion() (int, error)
		StoreDBVersion(version int) error
		InstanceID() (string, error)
		StoreInstanceID(ID string) error
		Edition() (portaineree.SoftwareEdition, error)
		StoreEdition(portaineree.SoftwareEdition) error
		PreviousDBVersion() (int, error)
		BucketName() string
	}

	// WebhookService represents a service for managing webhook data.
	WebhookService interface {
		Webhooks() ([]portaineree.Webhook, error)
		Webhook(ID portaineree.WebhookID) (*portaineree.Webhook, error)
		Create(portaineree *portaineree.Webhook) error
		UpdateWebhook(ID portaineree.WebhookID, webhook *portaineree.Webhook) error
		WebhookByResourceID(resourceID string) (*portaineree.Webhook, error)
		WebhookByToken(token string) (*portaineree.Webhook, error)
		DeleteWebhook(serviceID portaineree.WebhookID) error
		BucketName() string
	}
)

func IsErrObjectNotFound(e error) bool {
	return e == errors.ErrObjectNotFound
}

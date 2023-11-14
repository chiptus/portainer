package portaineree

import (
	"context"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	liblicense "github.com/portainer/liblicense/v3"
	"github.com/portainer/portainer-ee/api/database/models"
	kubeModels "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/pkg/featureflags"

	nomad "github.com/hashicorp/nomad/api"
	v1 "k8s.io/api/core/v1"
)

type (
	// APIOperationAuthorizationRequest represent an request for the portainer.Authorization to execute an API operation
	APIOperationAuthorizationRequest struct {
		Path           string
		Method         string
		Authorizations portainer.Authorizations
	}

	EdgeAsyncCommandType          string
	EdgeAsyncCommandOperation     string
	EdgeAsyncContainerOperation   string
	EdgeAsyncImageOperation       string
	EdgeAsyncVolumeOperation      string
	EdgeAsyncStackOperation       string
	EdgeAsyncNormalStackOperation string

	// EdgeAsyncCommand represents a command that is executed by an Edge Agent. Follows JSONPatch RFC https://datatracker.ietf.org/doc/html/rfc6902
	EdgeAsyncCommand struct {
		ID            int                       `json:"id"`
		Type          EdgeAsyncCommandType      `json:"type"`
		EndpointID    portainer.EndpointID      `json:"endpointID"`
		Timestamp     time.Time                 `json:"timestamp"`
		Executed      bool                      `json:"executed"`
		Operation     EdgeAsyncCommandOperation `json:"op"`
		Path          string                    `json:"path"`
		Value         interface{}               `json:"value"`
		ScheduledTime string                    `json:"scheduledTime"`
	}

	AuthActivityLog struct {
		UserActivityLogBase `storm:"inline"`
		Type                AuthenticationActivityType     `json:"type" storm:"index"`
		Origin              string                         `json:"origin" storm:"index"`
		Context             portainer.AuthenticationMethod `json:"context" storm:"index"`
	}

	// AuthLogsQuery represent the options used to get AuthActivity logs
	AuthLogsQuery struct {
		UserActivityLogBaseQuery
		ContextTypes  []portainer.AuthenticationMethod
		ActivityTypes []AuthenticationActivityType
	}

	// AuthenticationActivityType represents the type of an authentication action
	AuthenticationActivityType int

	// OpenAMTDeviceInformation represents an AMT managed device information
	OpenAMTDeviceInformation struct {
		GUID             string                                  `json:"guid"`
		HostName         string                                  `json:"hostname"`
		ConnectionStatus bool                                    `json:"connectionStatus"`
		PowerState       portainer.PowerState                    `json:"powerState"`
		EnabledFeatures  *portainer.OpenAMTDeviceEnabledFeatures `json:"features"`
	}

	// CloudProvider represents a Kubernetes as a service cloud provider.
	CloudProvider struct {
		Provider  string   `json:"Provider"`
		Name      string   `json:"Name"`
		URL       string   `json:"URL"`
		Region    string   `json:"Region"`
		Size      *string  `json:"Size"`
		NodeCount int      `json:"NodeCount"`
		CPU       *int     `json:"CPU"`
		RAM       *float64 `json:"RAM"`
		HDD       *int     `json:"HDD"`
		// Pointer will hide this field for providers other than civo which do
		// not use this field.
		NetworkID *string `json:"NetworkID"`
		// CredentialID holds an ID of the credential used to create the cluster
		CredentialID models.CloudCredentialID `json:"CredentialID"`

		// Azure specific fields
		ResourceGroup     string
		Tier              string
		PoolName          string
		DNSPrefix         string
		KubernetesVersion string

		// Amazon specific fields
		AmiType        *string `json:"AmiType"`
		InstanceType   *string `json:"InstanceType"`
		NodeVolumeSize *int    `json:"NodeVolumeSize"`

		// MicroK8S specific fields
		AddonsWithArgs []MicroK8sAddon `json:"AddonWithArgs"`
		NodeIPs        *string         `json:"NodeIPs"`

		// @deprecated
		Addons *string `json:"-"` // Dont't send it back to the client
	}

	MicroK8sAddon struct {
		Name       string `json:"name"`
		Args       string `json:"arguments"`
		Repository string `json:"repository"`
	}

	// CLIFlags represents the available flags on the CLI
	CLIFlags struct {
		Addr                      *string
		AddrHTTPS                 *string
		TunnelAddr                *string
		TunnelPort                *string
		AdminPassword             *string
		AdminPasswordFile         *string
		Assets                    *string
		Data                      *string
		FeatureFlags              *[]string
		DemoEnvironment           *bool
		EnableEdgeComputeFeatures *bool
		EndpointURL               *string
		Labels                    *[]portainer.Pair
		Logo                      *string
		NoAnalytics               *bool
		Templates                 *string
		TLS                       *bool
		TLSSkipVerify             *bool
		TLSCacert                 *string
		TLSCert                   *string
		TLSKey                    *string
		HTTPDisabled              *bool
		HTTPEnabled               *bool
		SSL                       *bool
		SSLCert                   *string
		SSLKey                    *string
		SSLCACert                 *string
		MTLSCert                  *string
		MTLSKey                   *string
		MTLSCACert                *string
		Rollback                  *bool
		RollbackToCE              *bool
		SnapshotInterval          *string
		BaseURL                   *string
		InitialMmapSize           *int
		MaxBatchSize              *int
		MaxBatchDelay             *time.Duration
		SecretKeyName             *string
		LicenseKey                *string
		LicenseExpireAbsolute     *bool
		LogLevel                  *string
		LogMode                   *string
	}

	// EdgeGroup represents an Edge group
	EdgeGroup struct {
		// EdgeGroup Identifier
		ID           portainer.EdgeGroupID  `json:"Id" example:"1"`
		Name         string                 `json:"Name"`
		Dynamic      bool                   `json:"Dynamic"`
		TagIDs       []portainer.TagID      `json:"TagIds"`
		Endpoints    []portainer.EndpointID `json:"Endpoints"`
		PartialMatch bool                   `json:"PartialMatch"`
		EdgeUpdateID int
	}

	// EdgeJobStatus represents an Edge job status
	EdgeJobStatus struct {
		JobID          int    `json:"JobID"`
		LogFileContent string `json:"LogFileContent"`
	}

	// EdgeStaggerOption represents an Edge stack stagger option
	EdgeStaggerOption int

	// EdgeStaggerParallelOption represents an Edge stack stagger parallel option
	EdgeStaggerParallelOption int

	// EdgeUpdateFailureAction represents an Edge stack update failure action
	EdgeUpdateFailureAction int

	EdgeStaggerConfig struct {
		StaggerOption           EdgeStaggerOption
		StaggerParallelOption   EdgeStaggerParallelOption
		DeviceNumber            int
		DeviceNumberStartFrom   int
		DeviceNumberIncrementBy int
		// Timeout unit is minute
		Timeout string `example:"5"`
		// UpdateDelay unit is minute
		UpdateDelay         string `example:"5"`
		UpdateFailureAction EdgeUpdateFailureAction
	}

	//EdgeStack represents an edge stack
	EdgeStack struct {
		// EdgeStack Identifier
		ID             portainer.EdgeStackID                              `json:"Id" example:"1"`
		Name           string                                             `json:"Name"`
		CreationDate   int64                                              `json:"CreationDate"`
		EdgeGroups     []portainer.EdgeGroupID                            `json:"EdgeGroups"`
		Registries     []portainer.RegistryID                             `json:"Registries"`
		Status         map[portainer.EndpointID]portainer.EdgeStackStatus `json:"Status"`
		ProjectPath    string                                             `json:"ProjectPath"`
		EntryPoint     string                                             `json:"EntryPoint"`
		Version        int                                                `json:"Version"`
		NumDeployments int                                                `json:"NumDeployments"`
		ManifestPath   string                                             `json:"ManifestPath"`
		DeploymentType portainer.EdgeStackDeploymentType                  `json:"DeploymentType"`
		// EdgeUpdateID represents the parent update ID, will be zero if this stack is not part of an update
		EdgeUpdateID int
		// Schedule represents the schedule of the Edge stack (optional, format - 'YYYY-MM-DD HH:mm:ss')
		ScheduledTime string `example:"2020-11-13 14:53:00"`
		// Uses the manifest's namespaces instead of the default one
		UseManifestNamespaces bool
		// Pre Pull Image
		PrePullImage bool `json:"PrePullImage"`
		// Re-Pull Image
		RePullImage bool `json:"RePullImage"`
		// Retry deploy
		RetryDeploy bool `example:"false"`

		// The GitOps update settings of a git stack
		AutoUpdate *portainer.AutoUpdateSettings `json:"AutoUpdate"`
		// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
		Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
		// The git configuration of a git stack
		GitConfig *gittypes.RepoConfig

		// Whether the stack supports relative path volume
		SupportRelativePath bool `example:"false"`
		// Local filesystem path
		FilesystemPath string `example:"/tmp"`

		// Whether the edge stack supports per device configs
		SupportPerDeviceConfigs bool `example:"false"`
		// Per device configs match type
		PerDeviceConfigsMatchType portainer.PerDevConfigsFilterType `example:"file" enums:"file, dir"`
		// Per device configs group match type
		PerDeviceConfigsGroupMatchType portainer.PerDevConfigsFilterType `example:"file" enums:"file, dir"`
		// Per device configs path
		PerDeviceConfigsPath string `example:"configs"`

		// StackFileVersion represents the version of the stack file, such yaml, hcl, manifest file
		StackFileVersion int `json:"StackFileVersion" example:"1"`
		// PreviousDeploymentInfo represents the previous deployment info of the stack
		PreviousDeploymentInfo *portainer.StackDeploymentInfo `json:"PreviousDeploymentInfo"`
		// EnvVars is a list of environment variables to inject into the stack
		EnvVars []portainer.Pair
		// StaggerConfig is the configuration for staggered update
		StaggerConfig *EdgeStaggerConfig

		// Deprecated
		Prune bool `json:"Prune,omitempty"`
	}

	EndpointLog struct {
		DockerContainerID string `json:"dockerContainerID,omitempty"`
		StdOut            string `json:"stdOut,omitempty"`
		StdErr            string `json:"stdErr,omitempty"`
	}

	EdgeStackLog struct {
		EdgeStackID portainer.EdgeStackID `json:"edgeStackID,omitempty"`
		EndpointID  portainer.EndpointID  `json:"endpointID,omitempty"`
		Logs        []EndpointLog         `json:"logs,omitempty"`
	}

	// EndpointChangeWindow determine when GitOps stack/app updates may occur
	EndpointChangeWindow struct {
		Enabled   bool   `json:"Enabled" example:"true"`
		StartTime string `json:"StartTime" example:"22:00"`
		EndTime   string `json:"EndTime" example:"02:00"`
	}

	// DeploymentOptions hides manual deployment forms for an environment
	DeploymentOptions struct {
		OverrideGlobalOptions bool `json:"overrideGlobalOptions"`
		// Hide manual deploy forms in portainer
		HideAddWithForm bool `json:"hideAddWithForm" example:"true"`
		// Hide the webeditor in the remaining visible forms
		HideWebEditor bool `json:"hideWebEditor" example:"false"`
		// Hide the file upload option in the remaining visible forms
		HideFileUpload bool `json:"hideFileUpload" example:"false"`
	}

	// Environment(Endpoint) represents a Docker environment(endpoint) with all the info required
	// to connect to it
	Endpoint struct {
		// Environment(Endpoint) Identifier
		ID portainer.EndpointID `json:"Id" example:"1"`
		// Environment(Endpoint) name
		Name string `json:"Name" example:"my-environment"`
		// Environment(Endpoint) environment(endpoint) type. 1 for a Docker environment(endpoint), 2 for an agent on Docker environment(endpoint) or 3 for an Azure environment(endpoint).
		Type portainer.EndpointType `json:"Type" example:"1"`
		// URL or IP address of the Docker host associated to this environment(endpoint)
		URL string `json:"URL" example:"docker.mydomain.tld:2375"`
		// Environment(Endpoint) group identifier
		GroupID portainer.EndpointGroupID `json:"GroupId" example:"1"`
		// URL or IP address where exposed containers will be reachable
		PublicURL        string                     `json:"PublicURL" example:"docker.mydomain.tld:2375"`
		Gpus             []portainer.Pair           `json:"Gpus"`
		TLSConfig        portainer.TLSConfiguration `json:"TLSConfig"`
		AzureCredentials portainer.AzureCredentials `json:"AzureCredentials,omitempty"`
		// List of tag identifiers to which this environment(endpoint) is associated
		TagIDs []portainer.TagID `json:"TagIds"`
		// The status of the environment(endpoint) (1 - up, 2 - down, 3 -
		// provisioning, 4 - error)
		Status portainer.EndpointStatus `json:"Status" example:"1"`
		// A message that describes the status. Should be included for Status 3
		// or 4.
		StatusMessage EndpointStatusMessage `json:"StatusMessage"`
		// A Kubernetes as a service cloud provider. Only included if this
		// endpoint was created using KaaS provisioning.
		CloudProvider *CloudProvider `json:"CloudProvider"`
		// List of snapshots
		Snapshots []portainer.DockerSnapshot `json:"Snapshots"`
		// List of user identifiers authorized to connect to this environment(endpoint)
		UserAccessPolicies portainer.UserAccessPolicies `json:"UserAccessPolicies"`
		// List of team identifiers authorized to connect to this environment(endpoint)
		TeamAccessPolicies portainer.TeamAccessPolicies `json:"TeamAccessPolicies"`
		// The identifier of the edge agent associated with this environment(endpoint)
		EdgeID string `json:"EdgeID,omitempty"`
		// The key which is used to map the agent to Portainer
		EdgeKey string `json:"EdgeKey"`

		// Associated Kubernetes data
		Kubernetes KubernetesData `json:"Kubernetes"`
		// Associated Nomad data
		Nomad NomadData `json:"Nomad"`
		// Maximum version of docker-compose
		ComposeSyntaxMaxVersion string `json:"ComposeSyntaxMaxVersion" example:"3.8"`
		// Environment(Endpoint) specific security settings
		SecuritySettings portainer.EndpointSecuritySettings
		// The identifier of the AMT Device associated with this environment(endpoint)
		AMTDeviceGUID string `json:"AMTDeviceGUID,omitempty" example:"4c4c4544-004b-3910-8037-b6c04f504633"`
		// LastCheckInDate mark last check-in date on checkin
		LastCheckInDate int64
		// QueryDate of each query with the endpoints list
		QueryDate int64
		// Heartbeat indicates the heartbeat status of an edge environment
		Heartbeat bool `json:"Heartbeat" example:"true"`

		// Whether the device has been trusted or not by the user
		UserTrusted bool

		// Whether we need to run any "post init migrations".
		PostInitMigrations EndpointPostInitMigrations `json:"PostInitMigrations"`

		// The check in interval for edge agent (in seconds)
		EdgeCheckinInterval int `json:"EdgeCheckinInterval" example:"5"`

		Edge portainer.EnvironmentEdgeSettings

		Agent EnvironmentAgentData

		// LocalTimeZone is the local time zone of the endpoint
		LocalTimeZone string

		// GitOps update change window restriction for stacks and apps
		ChangeWindow EndpointChangeWindow `json:"ChangeWindow"`
		// Hide manual deployment forms for an environment
		DeploymentOptions *DeploymentOptions `json:"DeploymentOptions"`

		EnableImageNotification bool `json:"EnableImageNotification"`

		EnableGPUManagement bool `json:"EnableGPUManagement"`

		// Deprecated fields
		// Deprecated in DBVersion == 4
		TLS           bool   `json:"TLS,omitempty"`
		TLSCACertPath string `json:"TLSCACert,omitempty"`
		TLSCertPath   string `json:"TLSCert,omitempty"`
		TLSKeyPath    string `json:"TLSKey,omitempty"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []portainer.UserID `json:"AuthorizedUsers"`
		AuthorizedTeams []portainer.TeamID `json:"AuthorizedTeams"`

		// Deprecated in DBVersion == 22
		Tags []string `json:"Tags"`

		// Deprecated v2.18
		IsEdgeDevice bool `json:"IsEdgeDevice,omitempty"`
	}

	// EnvironmentAgentData represents the data associated to an agent deployed
	EnvironmentAgentData struct {
		Version         string `json:"Version,omitempty" example:"1.0.0"`
		PreviousVersion string `json:"PreviousVersion,omitempty" example:"1.0.0"`
	}

	// EndpointStatusMessage represents the current status of a provisioning or
	// failed endpoint.
	EndpointStatusMessage struct {
		Summary string `json:"summary"`
		Detail  string `json:"detail"`

		// TODO: in future versions, we should think about removing these fields and
		// create a separate bucket to store cluster operation messages instead or try to find a better way.
		// Operation/OperationStatus blank means, nothing is happening
		Operation       string                  `json:"operation"`       // ,scale,upgrade,addons
		OperationStatus EndpointOperationStatus `json:"operationStatus"` // ,processing,error
	}

	// EndpointPostInitMigrations
	EndpointPostInitMigrations struct {
		MigrateIngresses  bool `json:"MigrateIngresses"`
		MigrateGPUs       bool `json:"MigrateGPUs"`
		MigrateGateKeeper bool `json:"MigrateGateKeeper"`
	}

	// Extension represents a deprecated Portainer extension
	Extension struct {
		// Extension Identifier
		ID               portainer.ExtensionID       `json:"Id" example:"1"`
		Enabled          bool                        `json:"Enabled"`
		Name             string                      `json:"Name,omitempty"`
		ShortDescription string                      `json:"ShortDescription,omitempty"`
		Description      string                      `json:"Description,omitempty"`
		DescriptionURL   string                      `json:"DescriptionURL,omitempty"`
		Price            string                      `json:"Price,omitempty"`
		PriceDescription string                      `json:"PriceDescription,omitempty"`
		Deal             bool                        `json:"Deal,omitempty"`
		Available        bool                        `json:"Available,omitempty"`
		License          ExtensionLicenseInformation `json:"License,omitempty"`
		Version          string                      `json:"Version"`
		UpdateAvailable  bool                        `json:"UpdateAvailable"`
		ShopURL          string                      `json:"ShopURL,omitempty"`
		Images           []string                    `json:"Images,omitempty"`
		Logo             string                      `json:"Logo,omitempty"`
	}

	// ExtensionLicenseInformation represents information about an extension license
	ExtensionLicenseInformation struct {
		LicenseKey string `json:"LicenseKey,omitempty"`
		Company    string `json:"Company,omitempty"`
		Expiration string `json:"Expiration,omitempty"`
		Valid      bool   `json:"Valid,omitempty"`
	}

	// HelmUserRepositories stores a Helm repository URL for the given user
	HelmUserRepository struct {
		// Membership Identifier
		ID portainer.HelmUserRepositoryID `json:"Id" example:"1"`
		// User identifier
		UserID portainer.UserID `json:"UserId" example:"1"`
		// Helm repository URL
		URL string `json:"URL" example:"https://charts.bitnami.com/bitnami"`
	}

	// GithubRegistryData represents data required for Github registry to work
	GithubRegistryData struct {
		UseOrganisation  bool   `json:"UseOrganisation"`
		OrganisationName string `json:"OrganisationName"`
	}

	K8sNamespaceInfo struct {
		IsSystem  bool        `json:"IsSystem"`
		IsDefault bool        `json:"IsDefault"`
		Status    interface{} `json:"Status"`
	}

	// K8sRole represents a K8s role name
	K8sRole string

	// KubernetesData contains all the Kubernetes related environment(endpoint) information
	KubernetesData struct {
		Snapshots     []portainer.KubernetesSnapshot `json:"Snapshots"`
		Configuration KubernetesConfiguration        `json:"Configuration"`
		Flags         portainer.KubernetesFlags      `json:"Flags"`
	}

	MTLSSettings struct {
		UseSeparateCert bool `json:"UseSeparateCert"`
		// CaCertFile is the path to the mTLS CA certificate file
		CaCertFile string `json:"CaCertFile"`
		// CertFile is the path to the mTLS certificate file
		CertFile string `json:"CertFile"`
		// KeyFile is the path to the mTLS key file
		KeyFile string `json:"KeyFile"`
	}

	// NomadData contains all the Nomad related environment(endpoint) information
	NomadData struct {
		Snapshots []NomadSnapshot `json:"Snapshots"`
	}

	NomadSnapshotTask struct {
		JobID        string    `json:"JobID"`
		Namespace    string    `json:"Namespace"`
		TaskName     string    `json:"TaskName"`
		State        string    `json:"State"`
		TaskGroup    string    `json:"TaskGroup"`
		AllocationID string    `json:"AllocationID"`
		StartedAt    time.Time `json:"StartedAt"`
	}

	NomadSnapshotJob struct {
		ID         string              `json:"ID"`
		Status     string              `json:"Status"`
		Namespace  string              `json:"Namespace"`
		SubmitTime int64               `json:"SubmitTime"`
		Tasks      []NomadSnapshotTask `json:"Tasks"`
	}

	// NomadSnapshot represents a snapshot of a specific Nomad environment(endpoint) at a specific time
	NomadSnapshot struct {
		Time             int64              `json:"Time"`
		NomadVersion     string             `json:"NomadVersion"`
		NodeCount        int                `json:"NodeCount"`
		JobCount         int                `json:"JobCount"`
		GroupCount       int                `json:"GroupCount"`
		TaskCount        int                `json:"TaskCount"`
		RunningTaskCount int                `json:"RunningTaskCount"`
		TotalCPU         int64              `json:"TotalCPU"`
		TotalMemory      int64              `json:"TotalMemory"`
		Jobs             []NomadSnapshotJob `json:"Jobs"`
	}

	// KubernetesConfiguration represents the configuration of a Kubernetes environment(endpoint)
	KubernetesConfiguration struct {
		UseLoadBalancer                 bool                                     `json:"UseLoadBalancer"`
		UseServerMetrics                bool                                     `json:"UseServerMetrics"`
		EnableResourceOverCommit        bool                                     `json:"EnableResourceOverCommit"`
		ResourceOverCommitPercentage    int                                      `json:"ResourceOverCommitPercentage"`
		StorageClasses                  []portainer.KubernetesStorageClassConfig `json:"StorageClasses"`
		IngressClasses                  []portainer.KubernetesIngressClassConfig `json:"IngressClasses"`
		RestrictDefaultNamespace        bool                                     `json:"RestrictDefaultNamespace"`
		IngressAvailabilityPerNamespace bool                                     `json:"IngressAvailabilityPerNamespace"`
		RestrictStandardUserIngressW    bool                                     `json:"RestrictStandardUserIngressW"`
		AllowNoneIngressClass           bool                                     `json:"AllowNoneIngressClass"`
	}

	// LDAPServerType represents the type of the LDAP server
	LDAPServerType int

	// LDAPSettings represents the settings used to connect to a LDAP server
	LDAPSettings struct {
		// Enable this option if the server is configured for Anonymous access. When enabled, ReaderDN and Password will not be used
		AnonymousMode bool `json:"AnonymousMode" example:"true" validate:"validate_bool"`
		// Account that will be used to search for users
		ReaderDN string `json:"ReaderDN" example:"cn=readonly-account,dc=ldap,dc=domain,dc=tld" validate:"required_if=AnonymousMode false"`
		// Password of the account that will be used to search users
		Password string `json:"Password,omitempty" example:"readonly-password" validate:"required_if=AnonymousMode false"`
		// URLs or IP addresses of the LDAP server
		URLs      []string                   `json:"URLs" validate:"validate_urls"`
		TLSConfig portainer.TLSConfiguration `json:"TLSConfig"`
		// Whether LDAP connection should use StartTLS
		StartTLS            bool                                `json:"StartTLS" example:"true"`
		SearchSettings      []portainer.LDAPSearchSettings      `json:"SearchSettings"`
		GroupSearchSettings []portainer.LDAPGroupSearchSettings `json:"GroupSearchSettings"`
		// Automatically provision users and assign them to matching LDAP group names
		AutoCreateUsers bool           `json:"AutoCreateUsers" example:"true"`
		ServerType      LDAPServerType `json:"ServerType" example:"1"`
		// Whether auto admin population is switched on or not
		AdminAutoPopulate        bool                                `json:"AdminAutoPopulate" example:"true"`
		AdminGroupSearchSettings []portainer.LDAPGroupSearchSettings `json:"AdminGroupSearchSettings"`
		// Saved admin group list, the user will be populated as an admin role if any user group matches the record in the list
		AdminGroups []string `json:"AdminGroups" example:"['manager','operator']"`
		// Deprecated
		URL string `json:"URL,omitempty" validate:"hostname_port"`
	}

	// LicenseInfo represents aggregated information about an instance license
	LicenseInfo struct {
		Company   string                          `json:"company"`
		ExpiresAt int64                           `json:"expiresAt"`
		Nodes     int                             `json:"nodes"`
		Type      liblicense.PortainerLicenseType `json:"type"`
		Valid     bool                            `json:"valid"`
		// unix timestamp when node usage exceeded avaiable license limit
		OveruseStartedTimestamp int64 `json:"overuseStartedTimestamp"`
	}

	// OAuthClaimMappings represents oAuth group claim value to portainer team name mapping
	OAuthClaimMappings struct {
		ClaimValRegex string `json:"ClaimValRegex"`
		Team          int    `json:"Team"`
	}

	// TeamMemberships represents oAuth group claim to portainer team membership mappings
	TeamMemberships struct {
		OAuthClaimName            string               `json:"OAuthClaimName"`
		OAuthClaimMappings        []OAuthClaimMappings `json:"OAuthClaimMappings"`
		AdminAutoPopulate         bool                 `json:"AdminAutoPopulate"`
		AdminGroupClaimsRegexList []string             `json:"AdminGroupClaimsRegexList"`
	}

	// OAuthSettings represents the settings used to authorize with an portainer.Authorization server
	OAuthSettings struct {
		MicrosoftTenantID           string           `json:"MicrosoftTenantID"`
		ClientID                    string           `json:"ClientID"`
		ClientSecret                string           `json:"ClientSecret,omitempty"`
		AccessTokenURI              string           `json:"AccessTokenURI"`
		AuthorizationURI            string           `json:"AuthorizationURI"`
		ResourceURI                 string           `json:"ResourceURI"`
		RedirectURI                 string           `json:"RedirectURI"`
		UserIdentifier              string           `json:"UserIdentifier"`
		Scopes                      string           `json:"Scopes"`
		OAuthAutoCreateUsers        bool             `json:"OAuthAutoCreateUsers"`
		OAuthAutoMapTeamMemberships bool             `json:"OAuthAutoMapTeamMemberships"`
		TeamMemberships             TeamMemberships  `json:"TeamMemberships"`
		DefaultTeamID               portainer.TeamID `json:"DefaultTeamID"`
		SSO                         bool             `json:"SSO"`
		HideInternalAuth            bool             `json:"HideInternalAuth"`
		LogoutURI                   string           `json:"LogoutURI"`
		KubeSecretKey               []byte           `json:"KubeSecretKey"`
	}

	// OAuthInfo represents extracted data from the resource object obtained from an OAuth providers resource URL
	OAuthInfo struct {
		Username string
		Teams    []string
	}

	// Registry represents a Docker registry with all the info required
	// to connect to it
	Registry struct {
		// Registry Identifier
		ID portainer.RegistryID `json:"Id" example:"1"`
		// Registry Type (1 - Quay, 2 - Azure, 3 - Custom, 4 - Gitlab, 5 - ProGet, 6 - DockerHub, 7 - ECR, 8 - Github)
		Type portainer.RegistryType `json:"Type" enums:"1,2,3,4,5,6,7,8"`
		// Registry Name
		Name string `json:"Name" example:"my-registry"`
		// URL or IP address of the Docker registry
		URL string `json:"URL" example:"registry.mydomain.tld:2375/feed-name"`
		// Base URL, introduced for ProGet registry
		BaseURL string `json:"BaseURL" example:"registry.mydomain.tld:2375"`
		// Is authentication against this registry enabled
		Authentication bool `json:"Authentication" example:"true"`
		// Username or AccessKeyID used to authenticate against this registry
		Username string `json:"Username" example:"registry user"`
		// Password or SecretAccessKey used to authenticate against this registry
		Password                string                                     `json:"Password,omitempty" example:"registry_password"`
		ManagementConfiguration *portainer.RegistryManagementConfiguration `json:"ManagementConfiguration"`
		Gitlab                  portainer.GitlabRegistryData               `json:"Gitlab"`
		Quay                    portainer.QuayRegistryData                 `json:"Quay"`
		Github                  GithubRegistryData                         `json:"Github"`
		Ecr                     portainer.EcrData                          `json:"Ecr"`
		RegistryAccesses        portainer.RegistryAccesses                 `json:"RegistryAccesses"`

		// Deprecated fields
		// Deprecated in DBVersion == 31
		UserAccessPolicies portainer.UserAccessPolicies `json:"UserAccessPolicies"`
		// Deprecated in DBVersion == 31
		TeamAccessPolicies portainer.TeamAccessPolicies `json:"TeamAccessPolicies"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []portainer.UserID `json:"AuthorizedUsers"`
		// Deprecated in DBVersion == 18
		AuthorizedTeams []portainer.TeamID `json:"AuthorizedTeams"`

		// Stores temporary access token
		AccessToken       string `json:"AccessToken,omitempty"`
		AccessTokenExpiry int64  `json:"AccessTokenExpiry,omitempty"`
	}

	// Role represents a set of authorizations that can be associated to a user or
	// to a team.
	Role struct {
		// Role Identifier
		ID portainer.RoleID `json:"Id" example:"1" validate:"required"`
		// Role name
		Name string `json:"Name" example:"HelpDesk" validate:"required"`
		// Role description
		Description string `json:"Description" example:"Read-only access of all resources in an environment(endpoint)" validate:"required"`
		// Authorizations associated to a role
		Authorizations portainer.Authorizations `json:"Authorizations" validate:"required"`
		Priority       int                      `json:"Priority" validate:"required"`
	}

	// APIKey represents an API key
	APIKey struct {
		ID          portainer.APIKeyID `json:"id" example:"1"`
		UserID      portainer.UserID   `json:"userId" example:"1"`
		Description string             `json:"description" example:"portainer-api-key"`
		Prefix      string             `json:"prefix"`           // API key identifier (7 char prefix)
		DateCreated int64              `json:"dateCreated"`      // Unix timestamp (UTC) when the API key was created
		LastUsed    int64              `json:"lastUsed"`         // Unix timestamp (UTC) when the API key was last used
		Digest      []byte             `json:"digest,omitempty"` // Digest represents SHA256 hash of the raw API key
	}

	// GitCredentialID represents a git credential identifier
	GitCredentialID int

	// GitCredential represents a git credential
	GitCredential struct {
		ID           GitCredentialID  `json:"id" example:"1"`
		UserID       portainer.UserID `json:"userId" example:"1"`
		Name         string           `json:"name"`
		Username     string           `json:"username"`
		Password     string           `json:"password,omitempty"`
		CreationDate int64            `json:"creationDate" example:"1587399600"`
	}

	// S3BackupSettings represents when and where to backup
	S3BackupSettings struct {
		// Crontab rule to make periodical backups
		CronRule string `json:"cronRule"`
		// AWS access key id
		AccessKeyID string `json:"accessKeyID"`
		// AWS secret access key
		SecretAccessKey string `json:"secretAccessKey"`
		// AWS S3 region. Default to "us-east-1"
		Region string `json:"region" example:"us-east-1"`
		// AWS S3 bucket name
		BucketName string `json:"bucketName"`
		// Password to encrypt the backup with
		Password string `json:"password"`
		// S3 compatible host
		S3CompatibleHost string `json:"s3CompatibleHost"`
	}

	// S3BackupStatus represents result of the scheduled s3 backup
	S3BackupStatus struct {
		Failed    bool
		Timestamp time.Time
	}

	// S3Location represents s3 file localtion
	S3Location struct {
		// AWS access key id
		AccessKeyID string
		// AWS secret access key
		SecretAccessKey string
		// AWS S3 region. Default to "us-east-1"
		Region string `example:"us-east-1"`
		// AWS S3 bucket name
		BucketName string
		// AWS S3 filename in the bucket
		Filename string
		// S3 compatible host
		S3CompatibleHost string
	}

	// Schedule represents a scheduled job.
	// It only contains a pointer to one of the JobRunner implementations
	// based on the JobType.
	// NOTE: The Recurring option is only used by ScriptExecutionJob at the moment
	// Deprecated in favor of EdgeJob
	Schedule struct {
		// Schedule Identifier
		ID             portainer.ScheduleID `json:"Id" example:"1"`
		Name           string
		CronExpression string
		Recurring      bool
		Created        int64
		JobType        portainer.JobType
		EdgeSchedule   *portainer.EdgeSchedule
	}

	CloudApiKeys struct {
		CivoApiKey        string `json:"CivoApiKey" example:"DgJ33kwIhnHumQFyc8ihGwWJql9cC8UJDiBhN8YImKqiX"`
		DigitalOceanToken string `json:"DigitalOceanToken" example:"dop_v1_n9rq7dkcbg3zb3bewtk9nnvmfkyfnr94d42n28lym22vhqu23rtkllsldygqm22v"`
		LinodeToken       string `json:"LinodeToken" example:"92gsh9r9u5helgs4eibcuvlo403vm45hrmc6mzbslotnrqmkwc1ovqgmolcyq0wc"`
		GKEApiKey         string `json:"GKEApiKey" example:"an entire base64ed key file"`
	}

	CloudManagementRequest interface{}

	// CloudProvisioningRequest represents a requested Cloud Kubernetes Cluster
	// which should be executed to create a CloudProvisioningTask.
	CloudProvisioningRequest struct {
		EndpointID        portainer.EndpointID
		Provider          string
		Region            string
		Name              string
		NodeSize          string
		NetworkID         string
		NodeCount         int
		CPU               int
		RAM               float64
		HDD               int
		KubernetesVersion string
		CredentialID      models.CloudCredentialID
		StartingState     int

		// Azure specific fields
		ResourceGroup     string
		ResourceGroupName string
		ResourceName      string
		Tier              string
		PoolName          string
		PoolType          containerservice.AgentPoolType
		DNSPrefix         string
		// Azure AKS
		// --------------------------------------------------
		// AvailabilityZones - The list of Availability zones to use for nodes.
		// This can only be specified if the AgentPoolType property is 'VirtualMachineScaleSets'.
		AvailabilityZones []string

		// Amazon specific fields
		AmiType        string
		InstanceType   string
		NodeVolumeSize int

		// Microk8S specific fields
		MasterNodes   []string
		WorkerNodes   []string
		AddonWithArgs []MicroK8sAddon

		CustomTemplateID      portainer.CustomTemplateID
		CustomTemplateContent string

		// --- Common portainer internal fields ---
		// the userid of the user who created this request.
		CreatedByUserID portainer.UserID

		// @deprecated
		Addons []string
	}

	CloudScalingRequest interface {
		Provider() string
		String() string
	}

	CloudUpgradeRequest interface {
		Provider() string
	}

	// CloudProvisioningID represents a cloud provisioning identifier
	CloudProvisioningTaskID int64

	// CloudProvisioningTask represents an active job queue for KaaS provisioning tasks
	//   used by portainer when stopping and restarting portainer
	CloudProvisioningTask struct {
		ID                    CloudProvisioningTaskID
		Provider              string
		ClusterID             string
		Region                string
		EndpointID            portainer.EndpointID
		CreatedAt             time.Time
		CreatedByUserID       portainer.UserID
		CustomTemplateID      portainer.CustomTemplateID
		CustomTemplateContent string

		State   int   `json:"-"`
		Retries int   `json:"-"`
		Err     error `json:"-"`

		// AZURE specific fields
		ResourceGroup string

		// Microk8s specific fields
		MasterNodes []string
		WorkerNodes []string
	}

	// GlobalDeploymentOptions hides manual deployment forms globally, to enforce infrastructure as code practices
	GlobalDeploymentOptions struct {
		// Hide manual deploy forms in portainer
		HideAddWithForm bool `json:"hideAddWithForm" example:"false"`
		// Configure this per environment or globally
		PerEnvOverride bool `json:"perEnvOverride" example:"false"`
		// Hide the webeditor in the remaining visible forms
		HideWebEditor bool `json:"hideWebEditor" example:"false"`
		// Hide the file upload option in the remaining visible forms
		HideFileUpload bool `json:"hideFileUpload" example:"false"`
		// Make note field mandatory if enabled
		RequireNoteOnApplications bool `json:"requireNoteOnApplications" example:"false"`
		MinApplicationNoteLength  int  `json:"minApplicationNoteLength" example:"10"`

		HideStacksFunctionality bool `json:"hideStacksFunctionality,omitempty" example:"false"`
	}

	Edge struct {
		// The command list interval for edge agent - used in edge async mode (in seconds)
		CommandInterval int `json:"CommandInterval" example:"5"`
		// The ping interval for edge agent - used in edge async mode (in seconds)
		PingInterval int `json:"PingInterval" example:"5"`
		// The snapshot interval for edge agent - used in edge async mode (in seconds)
		SnapshotInterval int `json:"SnapshotInterval" example:"5"`

		MTLS MTLSSettings
		// The address where the tunneling server can be reached by Edge agents
		TunnelServerAddress string `json:"TunnelServerAddress" example:"portainer.domain.tld"`

		// Deprecated 2.18
		AsyncMode bool `json:"AsyncMode,omitempty" example:"false"`
	}

	// Settings represents the application settings
	Settings struct {
		// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
		LogoURL string `json:"LogoURL" example:"https://mycompany.mydomain.tld/logo.png"`
		// The content in plaintext used to display in the login page. Will hide when value is empty string
		CustomLoginBanner string `json:"CustomLoginBanner"`
		// A list of label name & value that will be used to hide containers when querying containers
		BlackListedLabels []portainer.Pair `json:"BlackListedLabels"`
		// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
		AuthenticationMethod portainer.AuthenticationMethod `json:"AuthenticationMethod" example:"1"`
		InternalAuthSettings portainer.InternalAuthSettings `json:"InternalAuthSettings"`
		LDAPSettings         LDAPSettings                   `json:"LDAPSettings"`
		OAuthSettings        OAuthSettings                  `json:"OAuthSettings"`
		OpenAMTConfiguration portainer.OpenAMTConfiguration `json:"openAMTConfiguration"`
		FDOConfiguration     portainer.FDOConfiguration     `json:"fdoConfiguration"`
		// The interval in which environment(endpoint) snapshots are created
		SnapshotInterval string `json:"SnapshotInterval" example:"5m"`
		// URL to the templates that will be displayed in the UI when navigating to App Templates
		TemplatesURL string `json:"TemplatesURL" example:"https://raw.githubusercontent.com/portainer/templates/v3/templates.json"`
		// Deployment options for encouraging git ops workflows
		GlobalDeploymentOptions GlobalDeploymentOptions `json:"GlobalDeploymentOptions"`
		// Show the Kompose build option (discontinued in 2.18)
		ShowKomposeBuildOption bool `json:"ShowKomposeBuildOption" example:"false"`
		// Whether edge compute features are enabled
		EnableEdgeComputeFeatures bool `json:"EnableEdgeComputeFeatures"`
		// The duration of a user session
		UserSessionTimeout string `json:"UserSessionTimeout" example:"5m"`
		// The expiry of a Kubeconfig
		KubeconfigExpiry string `json:"KubeconfigExpiry" example:"24h"`
		// Whether telemetry is enabled
		EnableTelemetry bool `json:"EnableTelemetry" example:"false"`
		// Helm repository URL, defaults to "https://charts.bitnami.com/bitnami"
		HelmRepositoryURL string `json:"HelmRepositoryURL" example:"https://charts.bitnami.com/bitnami"`
		// KubectlImage, defaults to portainer/kubectl-shell
		KubectlShellImage string `json:"KubectlShellImage" example:"portainer/kubectl-shell"`
		// TrustOnFirstConnect makes Portainer accepting edge agent connection by default
		TrustOnFirstConnect bool `json:"TrustOnFirstConnect" example:"false"`
		// EnforceEdgeID makes Portainer store the Edge ID instead of accepting anyone
		EnforceEdgeID bool `json:"EnforceEdgeID" example:"false"`
		// Container environment parameter AGENT_SECRET
		AgentSecret string `json:"AgentSecret"`
		// EdgePortainerURL is the URL that is exposed to edge agents
		EdgePortainerURL string `json:"EdgePortainerUrl"`
		// CloudAPIKeys
		CloudApiKeys CloudApiKeys `json:"CloudApiKeys"`
		// The default check in interval for edge agent (in seconds)
		EdgeAgentCheckinInterval int `json:"EdgeAgentCheckinInterval" example:"5"`

		// the default builtin registry now is anonymous docker hub registry
		DefaultRegistry struct {
			Hide bool `json:"Hide" example:"false"`
		}

		Edge Edge `json:"Edge"`

		// Experimental features
		ExperimentalFeatures ExperimentalFeatures `json:"ExperimentalFeatures"`

		// Deprecated fields
		DisplayDonationHeader       bool `json:"DisplayDonationHeader,omitempty"`
		DisplayExternalContributors bool `json:"DisplayExternalContributors,omitempty"`

		// Deprecated fields v26
		EnableHostManagementFeatures              bool `json:"EnableHostManagementFeatures,omitempty"`
		AllowVolumeBrowserForRegularUsers         bool `json:"AllowVolumeBrowserForRegularUsers,omitempty"`
		AllowBindMountsForRegularUsers            bool `json:"AllowBindMountsForRegularUsers,omitempty"`
		AllowPrivilegedModeForRegularUsers        bool `json:"AllowPrivilegedModeForRegularUsers,omitempty"`
		AllowHostNamespaceForRegularUsers         bool `json:"AllowHostNamespaceForRegularUsers,omitempty"`
		AllowStackManagementForRegularUsers       bool `json:"AllowStackManagementForRegularUsers,omitempty"`
		AllowDeviceMappingForRegularUsers         bool `json:"AllowDeviceMappingForRegularUsers,omitempty"`
		AllowContainerCapabilitiesForRegularUsers bool `json:"AllowContainerCapabilitiesForRegularUsers,omitempty"`

		IsDockerDesktopExtension bool `json:"IsDockerDesktopExtension"`
	}

	// ExperimentalFeatures represents experimental features that can be enabled
	ExperimentalFeatures struct {
		OpenAIIntegration bool `json:"OpenAIIntegration"`
	}

	// SSLSettings represents a pair of SSL certificate and key
	SSLSettings struct {
		CertPath    string `json:"certPath"`
		KeyPath     string `json:"keyPath"`
		CACertPath  string `json:"caCertPath"`
		SelfSigned  bool   `json:"selfSigned"`
		HTTPEnabled bool   `json:"httpEnabled"`
	}

	// Stack represents a Docker stack created via docker stack deploy
	Stack struct {
		// Stack Identifier
		ID portainer.StackID `json:"Id" example:"1"`
		// Stack name
		Name string `json:"Name" example:"myStack"`
		// Stack type. 1 for a Swarm stack, 2 for a Compose stack, 3 for a Kubernetes stack
		Type portainer.StackType `json:"Type" example:"2"`
		// Environment(Endpoint) identifier. Reference the environment(endpoint) that will be used for deployment
		EndpointID portainer.EndpointID `json:"EndpointId" example:"1"`
		// Cluster identifier of the Swarm cluster where the stack is deployed
		SwarmID string `json:"SwarmId" example:"jpofkc0i9uo9wtx1zesuk649w"`
		// Path to the Stack file
		EntryPoint string `json:"EntryPoint" example:"docker-compose.yml"`
		// A list of environment(endpoint) variables used during stack deployment
		Env []portainer.Pair `json:"Env"`
		//
		ResourceControl *portainer.ResourceControl `json:"ResourceControl"`
		// Stack status (1 - active, 2 - inactive)
		Status portainer.StackStatus `json:"Status" example:"1"`
		// Path on disk to the repository hosting the Stack file
		ProjectPath string `example:"/data/compose/myStack_jpofkc0i9uo9wtx1zesuk649w"`
		// The date in unix time when stack was created
		CreationDate int64 `example:"1587399600"`
		// The username which created this stack
		CreatedBy string `example:"admin"`
		// The date in unix time when stack was last updated
		UpdateDate int64 `example:"1587399600"`
		// The username which last updated this stack
		UpdatedBy string `example:"bob"`
		// Only applies when deploying stack with multiple files
		AdditionalFiles []string `json:"AdditionalFiles"`
		// The GitOps update settings of a git stack
		AutoUpdate *portainer.AutoUpdateSettings `json:"AutoUpdate"`
		// The stack deployment option
		Option *portainer.StackOption `json:"Option"`
		// The git configuration of a git stack
		GitConfig *gittypes.RepoConfig
		// Whether the stack is from a app template
		FromAppTemplate bool `example:"false"`
		// Kubernetes namespace if stack is a kube application
		Namespace string `example:"default"`
		// IsComposeFormat indicates if the Kubernetes stack is created from a Docker Compose file
		IsComposeFormat bool `example:"false"`
		// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
		Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
		// If stack support relative path volume
		SupportRelativePath bool `example:"false"`
		// Network(Swarm) or local(Standalone) filesystem path
		FilesystemPath string `example:"/tmp"`
		// StackFileVersion indicates the stack file version, such as yaml, hcl, and manifest
		StackFileVersion int `example:"1"`
		// The previous deployment info of the stack
		PreviousDeploymentInfo *portainer.StackDeploymentInfo `json:"PreviousDeploymentInfo"`
		// Whether the stack is detached from git
		IsDetachedFromGit bool `example:"false"`
	}

	// TunnelDetails represents information associated to a tunnel
	TunnelDetails struct {
		Status       string
		LastActivity time.Time
		Port         int
		Jobs         []portainer.EdgeJob
		Credentials  string
	}

	// User represents a user account
	User struct {
		// User Identifier
		ID       portainer.UserID `json:"Id" example:"1"`
		Username string           `json:"Username" example:"bob"`
		Password string           `json:"Password,omitempty" swaggerignore:"true"`
		// User role (1 for administrator account and 2 for regular account)
		Role                    portainer.UserRole               `json:"Role" example:"1"`
		TokenIssueAt            int64                            `json:"TokenIssueAt" example:"1639408208"`
		PortainerAuthorizations portainer.Authorizations         `json:"PortainerAuthorizations"`
		EndpointAuthorizations  portainer.EndpointAuthorizations `json:"EndpointAuthorizations"`
		ThemeSettings           UserThemeSettings

		// OpenAI integration parameters
		OpenAIApiKey string `json:"OpenAIApiKey" example:"sk-1234567890"`

		// Deprecated fields

		// Deprecated
		UserTheme string `json:"UserTheme,omitempty" example:"dark"`
	}

	// AuthActivityLog represents a log entry for user authentication activities

	UserActivityLogBase struct {
		ID        int    `json:"id" storm:"increment"`
		Timestamp int64  `json:"timestamp" storm:"index"`
		Username  string `json:"username" storm:"index"`
	}

	UserActivityLogBaseQuery struct {
		Limit           int
		Offset          int
		BeforeTimestamp int64
		AfterTimestamp  int64
		SortBy          string
		SortDesc        bool
		Keyword         string
	}

	UserActivityLog struct {
		UserActivityLogBase `storm:"inline"`
		Context             string `json:"context" storm:"index"`
		Action              string `json:"action" storm:"index"`
		Payload             []byte `json:"payload"`
	}

	// UserThemeSettings represents the theme settings for a user
	UserThemeSettings struct {
		// Color represents the color theme of the UI
		Color string `json:"color" example:"dark" enums:"dark,light,highcontrast,auto"`
		// SubtleUpgradeButton indicates if the upgrade banner should be displayed in a subtle way
		SubtleUpgradeButton bool `json:"subtleUpgradeButton"`
	}

	Snapshot struct {
		EndpointID portainer.EndpointID          `json:"EndpointId"`
		Docker     *portainer.DockerSnapshot     `json:"Docker"`
		Kubernetes *portainer.KubernetesSnapshot `json:"Kubernetes"`
		Nomad      *NomadSnapshot                `json:"Nomad"`
	}

	// AuthEventHandler represents an handler for an auth event
	AuthEventHandler interface {
		HandleUsersAuthUpdate()
		HandleUserAuthDelete(userID portainer.UserID)
		HandleEndpointAuthUpdate(endpointID portainer.EndpointID)
	}

	// CLIService represents a service for managing CLI
	CLIService interface {
		ParseFlags(version string) (*CLIFlags, error)
		ValidateFlags(flags *CLIFlags) error
	}

	// ComposeStackManager represents a service to manage Compose stacks
	ComposeStackManager interface {
		ComposeSyntaxMaxVersion() string
		NormalizeStackName(name string) string
		Up(ctx context.Context, stack *Stack, endpoint *Endpoint, forceRecreate bool) error
		Down(ctx context.Context, stack *Stack, endpoint *Endpoint) error
		Pull(ctx context.Context, stack *Stack, endpoint *Endpoint) error
	}

	// DockerSnapshotter represents a service used to create Docker environment(endpoint) snapshots
	DockerSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*portainer.DockerSnapshot, error)
	}

	// FileService represents a service for managing files
	FileService interface {
		portainer.FileService

		GetKaasFolder() string
		StoreSSLClientCert(certData []byte) error
		GetSSLClientCertPath() string
		StoreEdgeConfigFile(ID EdgeConfigID, path string, r io.Reader) error
		GetEdgeConfigFilepaths(ID EdgeConfigID, version EdgeConfigVersion) (basePath string, filepaths []string, err error)
		GetEdgeConfigDirEntries(edgeConfig *EdgeConfig, edgeID string, version EdgeConfigVersion) (dirEntries []filesystem.DirEntry, err error)
		RotateEdgeConfigs(ID EdgeConfigID) error
	}

	// OpenAMTService represents a service for managing OpenAMT
	OpenAMTService interface {
		Configure(configuration portainer.OpenAMTConfiguration) error
		DeviceInformation(configuration portainer.OpenAMTConfiguration, deviceGUID string) (*OpenAMTDeviceInformation, error)
		EnableDeviceFeatures(configuration portainer.OpenAMTConfiguration, deviceGUID string, features portainer.OpenAMTDeviceEnabledFeatures) (string, error)
		ExecuteDeviceAction(configuration portainer.OpenAMTConfiguration, deviceGUID string, action string) error
	}

	// JWTService represents a service for managing JWT tokens
	JWTService interface {
		GenerateToken(data *portainer.TokenData) (string, error)
		GenerateTokenForOAuth(data *portainer.TokenData, expiryTime *time.Time) (string, error)
		GenerateTokenForKubeconfig(data *portainer.TokenData) (string, error)
		ParseAndVerifyToken(token string) (*portainer.TokenData, error)
		SetUserSessionDuration(userSessionDuration time.Duration)
	}

	// KubeClient represents a service used to query a Kubernetes environment(endpoint)
	KubeClient interface {
		SetupUserServiceAccount(
			user User,
			endpointRoleID portainer.RoleID,
			namespaces map[string]K8sNamespaceInfo,
			namespaceRoles map[string]Role,
			clusterConfig KubernetesConfiguration,
		) error
		IsRBACEnabled() (bool, error)
		GetServiceAccount(tokendata *portainer.TokenData) (*v1.ServiceAccount, error)
		GetServiceAccounts(namespace string) ([]kubeModels.K8sServiceAccount, error)
		DeleteServiceAccounts(reqs kubeModels.K8sServiceAccountDeleteRequests) error
		GetServiceAccountBearerToken(userID int) (string, error)
		CreateUserShellPod(ctx context.Context, serviceAccountName, shellPodImage string) (*portainer.KubernetesShellPod, error)
		StartExecProcess(token string, useAdminToken bool, namespace, podName, containerName string, command []string, stdin io.Reader, stdout io.Writer, errChan chan error)
		CreateNamespace(info kubeModels.K8sNamespaceDetails) error
		UpdateNamespace(info kubeModels.K8sNamespaceDetails) error
		GetNamespaces() (map[string]K8sNamespaceInfo, error)
		GetNamespace(string) (K8sNamespaceInfo, error)
		DeleteNamespace(namespace string) error
		GetConfigMapsAndSecrets(namespace string) ([]kubeModels.K8sConfigMapOrSecret, error)
		GetApplications(namespace, kind string) ([]kubeModels.K8sApplication, error)
		GetApplication(namespace, kind, name string) (kubeModels.K8sApplication, error)
		CreateIngress(namespace string, info kubeModels.K8sIngressInfo, owner string) error
		UpdateIngress(namespace string, info kubeModels.K8sIngressInfo) error
		GetIngresses(namespace string) ([]kubeModels.K8sIngressInfo, error)
		DeleteIngresses(reqs kubeModels.K8sIngressDeleteRequests) error
		GetIngressControllers() (kubeModels.K8sIngressControllers, error)
		GetMetrics() (kubeModels.K8sMetrics, error)
		GetStorage() ([]portainer.KubernetesStorageClassConfig, error)
		CreateService(namespace string, service kubeModels.K8sServiceInfo) error
		UpdateService(namespace string, service kubeModels.K8sServiceInfo) error
		GetServices(namespace string, lookupApplications bool) ([]kubeModels.K8sServiceInfo, error)
		DeleteServices(reqs kubeModels.K8sServiceDeleteRequests) error
		GetNodesLimits() (portainer.K8sNodesLimits, error)
		GetMaxResourceLimits(name string, overCommitEnabled bool, resourceOverCommitPercent int) (portainer.K8sNodeLimits, error)
		RemoveUserServiceAccount(userID int) error
		RemoveUserNamespaceBindings(
			userID int,
			namespace string,
		) error
		HasStackName(namespace string, stackName string) (bool, error)
		NamespaceAccessPoliciesDeleteNamespace(namespace string) error
		GetNamespaceAccessPolicies() (map[string]portainer.K8sNamespaceAccessPolicy, error)
		UpdateNamespaceAccessPolicies(accessPolicies map[string]portainer.K8sNamespaceAccessPolicy) error
		DeleteRegistrySecret(registry portainer.RegistryID, namespace string) error
		CreateRegistrySecret(registry *Registry, namespace string) error
		IsRegistrySecret(namespace, secretName string) (bool, error)
		ToggleSystemState(namespace string, isSystem bool) error
		DeployPortainerAgent(useNodePort bool) error
		UpsertPortainerK8sClusterRoles(clusterConfig KubernetesConfiguration) error
		GetPortainerAgentAddress(nodeIPs []string) (string, error)
		CheckRunningPortainerAgentDeployment(nodeIPs []string) error

		GetClusterRoles() ([]kubeModels.K8sClusterRole, error)
		DeleteClusterRoles(kubeModels.K8sClusterRoleDeleteRequests) error
		GetClusterRoleBindings() ([]kubeModels.K8sClusterRoleBinding, error)
		DeleteClusterRoleBindings(kubeModels.K8sClusterRoleBindingDeleteRequests) error

		GetRoles(namespace string) ([]kubeModels.K8sRole, error)
		DeleteRoles(kubeModels.K8sRoleDeleteRequests) error
		GetRoleBindings(namespace string) ([]kubeModels.K8sRoleBinding, error)
		DeleteRoleBindings(kubeModels.K8sRoleBindingDeleteRequests) error
	}

	// NomadClient represents a service used to query a Nomad environment(endpoint)
	NomadClient interface {
		Validate() (valid bool)
		Leader() (string, error)
		ListJobs(namespace string) (jobList []*nomad.JobListStub, err error)
		ListNodes() (nodeList []*nomad.NodeListStub, err error)
		ListAllocations(jobID, namespace string) (allocationsList []*nomad.AllocationListStub, err error)
		DeleteJob(jobID, namespace string) error
		TaskEvents(allocationID, taskName, namespace string) ([]*nomad.TaskEvent, error)
		TaskLogs(refresh bool, allocationID, taskName, namespace, logType, origin string, offset int64) (<-chan *nomad.StreamFrame, <-chan error)
	}

	// KubernetesDeployer represents a service to deploy a manifest inside a Kubernetes environment(endpoint)
	KubernetesDeployer interface {
		Deploy(userID portainer.UserID, endpoint *Endpoint, manifestFiles []string, namespace string) (string, error)
		Restart(userID portainer.UserID, endpoint *Endpoint, resourceList []string, namespace string) (string, error)
		DeployViaKubeConfig(kubeConfig string, clusterID string, manifestFile string) error
		Remove(userID portainer.UserID, endpoint *Endpoint, manifestFiles []string, namespace string) (string, error)
		ConvertCompose(data []byte) ([]byte, error)
	}

	// KubernetesSnapshotter represents a service used to create Kubernetes environment(endpoint) snapshots
	KubernetesSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*portainer.KubernetesSnapshot, error)
	}

	// NomadSnapshotter represents a service used to create Nomad environment(endpoint) snapshots
	NomadSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*NomadSnapshot, error)
	}

	// LDAPService represents a service used to authenticate users against a LDAP/AD
	LDAPService interface {
		AuthenticateUser(username, password string, settings *LDAPSettings) error
		TestConnectivity(settings *LDAPSettings) error
		GetUserGroups(username string, settings *LDAPSettings, useAutoAdminSearchSettings bool) ([]string, error)
		SearchGroups(settings *LDAPSettings) ([]portainer.LDAPUser, error)
		SearchAdminGroups(settings *LDAPSettings) ([]string, error)
		SearchUsers(settings *LDAPSettings) ([]string, error)
	}

	// LicenseService represents a service used to manage licenses
	LicenseService interface {
		AddLicense(licenseKey string, force bool) ([]string, error)
		DeleteLicense(licenseKey string) error
		Info() LicenseInfo
		Licenses() []liblicense.PortainerLicense
		ShouldEnforceOveruse() bool
		SyncLicenses() error
		Start() error
		WillBeEnforcedAt() int64
	}

	// LicenseRepository represents a service used to manage licenses store
	LicenseRepository interface {
		Licenses() ([]liblicense.PortainerLicense, error)
		License(licenseKey string) (*liblicense.PortainerLicense, error)
		AddLicense(licenseKey string, license *liblicense.PortainerLicense) error
		UpdateLicense(licenseKey string, license *liblicense.PortainerLicense) error
		DeleteLicense(licenseKey string) error
	}

	// OAuthService represents a service used to authenticate users using OAuth
	OAuthService interface {
		Authenticate(code string, configuration *OAuthSettings) (*OAuthInfo, error)
	}

	// ReverseTunnelService represents a service used to manage reverse tunnel connections.
	ReverseTunnelService interface {
		StartTunnelServer(addr, port string, snapshotService SnapshotService) error
		StopTunnelServer() error
		GenerateEdgeKey(apiURL, tunnelAddr string, endpointIdentifier int) string
		SetTunnelStatusToActive(endpointID portainer.EndpointID)
		SetTunnelStatusToRequired(endpointID portainer.EndpointID) error
		SetTunnelStatusToIdle(endpointID portainer.EndpointID)
		KeepTunnelAlive(endpointID portainer.EndpointID, ctx context.Context, maxKeepAlive time.Duration)
		GetTunnelDetails(endpointID portainer.EndpointID) TunnelDetails
		GetActiveTunnel(endpoint *Endpoint) (TunnelDetails, error)
		AddEdgeJob(endpoint *Endpoint, edgeJob *portainer.EdgeJob)
		RemoveEdgeJob(edgeJobID portainer.EdgeJobID)
		RemoveEdgeJobFromEndpoint(endpointID portainer.EndpointID, edgeJobID portainer.EdgeJobID)
	}

	// S3BackupService represents a storage service for managing S3 backup settings and status
	S3BackupService interface {
		GetStatus() (S3BackupStatus, error)
		DropStatus() error
		UpdateStatus(status S3BackupStatus) error
		UpdateSettings(settings S3BackupSettings) error
		GetSettings() (S3BackupSettings, error)
	}

	// SnapshotService represents a service for managing environment(endpoint) snapshots
	SnapshotService interface {
		Start()
		SetSnapshotInterval(snapshotInterval string) error
		SnapshotEndpoint(endpoint *Endpoint) error
		FillSnapshotData(endpoint *Endpoint) error
	}

	// SwarmStackManager represents a service to manage Swarm stacks
	SwarmStackManager interface {
		Login(registries []Registry, endpoint *Endpoint) error
		Logout(endpoint *Endpoint) error
		Deploy(stack *Stack, prune bool, pullImage bool, endpoint *Endpoint) error
		Remove(stack *Stack, endpoint *Endpoint) error
		NormalizeStackName(name string) string
	}

	UserActivityService interface {
		LogAuthActivity(username, origin string, context portainer.AuthenticationMethod, activityType AuthenticationActivityType) error
		LogUserActivity(username, context, action string, payload []byte) error
	}

	// UserActivityStore store all logs related to user activity: authentication, actions, ...
	UserActivityStore interface {
		BackupTo(w io.Writer) error
		Close() error

		GetAuthLogs(opts AuthLogsQuery) ([]*AuthActivityLog, int, error)
		StoreAuthLog(authLog *AuthActivityLog) error

		GetUserActivityLogs(opts UserActivityLogBaseQuery) ([]*UserActivityLog, int, error)
		StoreUserActivityLog(userLog *UserActivityLog) error
	}
)

const (
	// APIVersion is the version number of the Portainer API
	APIVersion = "2.20.0"
	// Edition is the edition of the Portainer API
	Edition = PortainerEE
	// ComposeSyntaxMaxVersion is a maximum supported version of the docker compose syntax
	ComposeSyntaxMaxVersion = "3.9"
	// AssetsServerURL represents the URL of the Portainer asset server
	AssetsServerURL = "https://portainer-io-assets.sfo2.digitaloceanspaces.com"
	// MessageOfTheDayURL represents the URL where Portainer EE MOTD message can be retrieved
	MessageOfTheDayURL = AssetsServerURL + "/motd-ee.json"
	// VersionCheckURL represents the URL used to retrieve the latest version of Portainer
	VersionCheckURL = "https://api.github.com/repos/portainer/portainer-ee/releases/latest"
	// PortainerAgentHeader represents the name of the header available in any agent response
	PortainerAgentHeader = "Portainer-Agent"
	// PortainerAgentEdgeIDHeader represent the name of the header containing the Edge ID associated to an agent/agent cluster
	PortainerAgentEdgeIDHeader = "X-PortainerAgent-EdgeID"
	// HTTPResponseAgentPlatform represents the name of the header containing the Agent platform
	HTTPResponseAgentPlatform = "Portainer-Agent-Platform"
	// PortainerAgentTargetHeader represent the name of the header containing the target node name
	PortainerAgentTargetHeader = "X-PortainerAgent-Target"
	// PortainerAgentSignatureHeader represent the name of the header containing the digital signature
	PortainerAgentSignatureHeader = "X-PortainerAgent-Signature"
	// PortainerAgentPublicKeyHeader represent the name of the header containing the public key
	PortainerAgentPublicKeyHeader = "X-PortainerAgent-PublicKey"
	// PortainerAgentKubernetesSATokenHeader represent the name of the header containing a Kubernetes SA token
	PortainerAgentKubernetesSATokenHeader = "X-PortainerAgent-SA-Token"
	// PortainerAgentTimeZoneHeader is the name of the header containing the timezone
	PortainerAgentTimeZoneHeader = "X-PortainerAgent-TimeZone"
	// PortainerAgentEdgeUpdateIDHeader is the name of the header that will have the update ID that started this container
	PortainerAgentEdgeUpdateIDHeader = "X-PortainerAgent-Update-ID"
	// PortainerAgentSignatureMessage represents the message used to create a digital signature
	// to be used when communicating with an agent
	PortainerAgentSignatureMessage = "Portainer-App"
	// DefaultSnapshotInterval represents the default interval between each environment snapshot job
	DefaultSnapshotInterval = "5m"
	// DefaultEdgeAgentCheckinIntervalInSeconds represents the default interval (in seconds) used by Edge agents to checkin with the Portainer instance
	DefaultEdgeAgentCheckinIntervalInSeconds = 5
	// DefaultHelmrepositoryURL represents the URL to the official templates supported by Bitnami
	DefaultHelmRepositoryURL = "https://charts.bitnami.com/bitnami"
	// DefaultUserSessionTimeout represents the default timeout after which the user session is cleared
	DefaultUserSessionTimeout = "8h"
	// DefaultUserSessionTimeout represents the default timeout after which the user session is cleared
	DefaultKubeconfigExpiry = "0"
	// DefaultKubectlShellImage represents the default image and tag for the kubectl shell
	DefaultKubectlShellImage = "portainer/kubectl-shell"
	// WebSocketKeepAlive web socket keep alive for edge environments
	WebSocketKeepAlive = 1 * time.Hour
	// For parsing 24hr time
	TimeFormat24 = "15:04"
	// Date-Time format that we use in Portainer app
	DateTimeFormat = "2006-01-02 15:04:05"
)

type EndpointOperationStatus string

const (
	EndpointOperationStatusProcessing EndpointOperationStatus = "processing"
	EndpointOperationStatusWarning    EndpointOperationStatus = "warning"
	EndpointOperationStatusError      EndpointOperationStatus = "error"
	EndpointOperationStatusDone       EndpointOperationStatus = ""
)

// List of supported features
const (
	FeatureFdo = "fdo"
)

var SupportedFeatureFlags = []featureflags.Feature{
	FeatureFdo,
}

const (
	_ portainer.AuthenticationMethod = iota
	// AuthenticationInternal represents the internal authentication method (authentication against Portainer API)
	AuthenticationInternal
	// AuthenticationLDAP represents the LDAP authentication method (authentication against a LDAP server)
	AuthenticationLDAP
	// AuthenticationOAuth represents the OAuth authentication method (authentication against a portainer.Authorization server)
	AuthenticationOAuth
)

// Represent different types of user activities
const (
	_ AuthenticationActivityType = iota
	AuthenticationActivitySuccess
	AuthenticationActivityFailure
	AuthenticationActivityLogOut
)

const (
	_ portainer.AgentPlatform = iota
	// AgentPlatformDocker represent the Docker platform (Standalone/Swarm)
	AgentPlatformDocker
	// AgentPlatformKubernetes represent the Kubernetes platform
	AgentPlatformKubernetes
	// AgentPlatformPodman represent the Podman platform
	AgentPlatformPodman
	// AgentPlatformNomad represent the Nomad platform
	AgentPlatformNomad
)

const (
	_ portainer.EdgeJobLogsStatus = iota
	// EdgeJobLogsStatusIdle represents an idle log collection job
	EdgeJobLogsStatusIdle
	// EdgeJobLogsStatusPending represents a pending log collection job
	EdgeJobLogsStatusPending
	// EdgeJobLogsStatusCollected represents a completed log collection job
	EdgeJobLogsStatusCollected
)

const (
	_ EdgeStaggerOption = iota
	// EdgeStaggerOptionAllAtOnce represents a staggered deployment where all nodes are updated at once
	EdgeStaggerOptionAllAtOnce
	// EdgeStaggerOptionOneByOne represents a staggered deployment where nodes are updated with parallel setting
	EdgeStaggerOptionParallel
)

const (
	_ EdgeStaggerParallelOption = iota
	// EdgeStaggerParallelOptionFixed represents a staggered deployment where nodes are updated with a fixed number of nodes in parallel
	EdgeStaggerParallelOptionFixed
	// EdgeStaggerParallelOptionIncremental represents a staggered deployment where nodes are updated with an incremental number of nodes in parallel
	EdgeStaggerParallelOptionIncremental
)

const (
	_ EdgeUpdateFailureAction = iota
	// EdgeUpdateFailureActionContinue represents that stagger update will continue regardless of whether the endpoint update status
	EdgeUpdateFailureActionContinue
	// EdgeUpdateFailureActionPause represents that stagger update will pause when the endpoint update status is failed
	EdgeUpdateFailureActionPause
	// EdgeUpdateFailureActionRollback represents that stagger update will rollback as long as one endpoint update status is failed
	EdgeUpdateFailureActionRollback
)

const (
	_ portainer.CustomTemplatePlatform = iota
	// CustomTemplatePlatformLinux represents a custom template for linux
	CustomTemplatePlatformLinux
	// CustomTemplatePlatformWindows represents a custom template for windows
	CustomTemplatePlatformWindows
)

const (
	// EdgeStackDeploymentCompose represent an edge stack deployed using a compose file
	EdgeStackDeploymentCompose portainer.EdgeStackDeploymentType = iota
	// EdgeStackDeploymentKubernetes represent an edge stack deployed using a kubernetes manifest file
	EdgeStackDeploymentKubernetes
	// EdgeStackDeploymentNomad represent an edge stack deployed using a nomad hcl job file
	EdgeStackDeploymentNomad
)

const (
	_ portainer.EndpointStatus = iota
	// EndpointStatusUp is used to represent an available environment(endpoint)
	EndpointStatusUp
	// EndpointStatusDown is used to represent an unavailable environment(endpoint)
	EndpointStatusDown
	// EndpointStatusProvisioning is used to represent an environment which is
	// being provisioned by a cloud provider.
	EndpointStatusProvisioning
	// EndpointStatusError represents a fatal error has occurred in the endpoint
	// and it cannot be recovered.
	EndpointStatusError
)

const (
	_ portainer.EndpointType = iota
	// DockerEnvironment represents an environment(endpoint) connected to a Docker environment(endpoint)
	DockerEnvironment
	// AgentOnDockerEnvironment represents an environment(endpoint) connected to a Portainer agent deployed on a Docker environment(endpoint)
	AgentOnDockerEnvironment
	// AzureEnvironment represents an environment(endpoint) connected to an Azure environment(endpoint)
	AzureEnvironment
	// EdgeAgentOnDockerEnvironment represents an environment(endpoint) connected to an Edge agent deployed on a Docker environment(endpoint)
	EdgeAgentOnDockerEnvironment
	// KubernetesLocalEnvironment represents an environment(endpoint) connected to a local Kubernetes environment(endpoint)
	KubernetesLocalEnvironment
	// AgentOnKubernetesEnvironment represents an environment(endpoint) connected to a Portainer agent deployed on a Kubernetes environment(endpoint)
	AgentOnKubernetesEnvironment
	// EdgeAgentOnKubernetesEnvironment represents an environment(endpoint) connected to an Edge agent deployed on a Kubernetes environment(endpoint)
	EdgeAgentOnKubernetesEnvironment
	// EdgeAgentOnNomadEnvironment represents an environment(endpoint) connected to an Edge agent deployed on a Nomad environment(endpoint)
	EdgeAgentOnNomadEnvironment
	// KubeConfigEnvironment represents an environment(endpoint) connected to a Kubernetes cluster
	// Note: this endpoint type is only being used for validating the request payload but
	// these environments are using `AgentOnKubernetesEnvironment` type when created
	KubeConfigEnvironment
)

const (
	_ portainer.JobType = iota
	// SnapshotJobType is a system job used to create environment(endpoint) snapshots
	SnapshotJobType = 2
)

// LDAPServerType represents the type of LDAP server
const (
	LDAPServerCustom LDAPServerType = iota
	LDAPServerOpenLDAP
	LDAPServerAD
)

const (
	_ portainer.MembershipRole = iota
	// TeamLeader represents a leader role inside a team
	TeamLeader
	// TeamMember represents a member role inside a team
	TeamMember
)

const (
	_ portainer.SoftwareEdition = iota
	// PortainerCE represents the community edition of Portainer
	PortainerCE
	// PortainerBE represents the business edition of Portainer
	PortainerBE
	// PortainerEE represents the business edition of Portainer
	PortainerEE
)

const (
	_ portainer.RegistryType = iota
	// QuayRegistry represents a Quay.io registry
	QuayRegistry
	// AzureRegistry represents an ACR registry
	AzureRegistry
	// CustomRegistry represents a custom registry
	CustomRegistry
	// GitlabRegistry represents a gitlab registry
	GitlabRegistry
	// ProGetRegistry represents a proget registry
	ProGetRegistry
	// DockerHubRegistry represents a dockerhub registry
	DockerHubRegistry
	// EcrRegistry represents an ECR registry
	EcrRegistry
	// Github container registry
	GithubRegistry
)

const (
	_ portainer.ResourceAccessLevel = iota
	// ReadWriteAccessLevel represents an access level with read-write permissions on a resource
	ReadWriteAccessLevel
)

const (
	_ portainer.ResourceControlType = iota
	// ContainerResourceControl represents a resource control associated to a Docker container
	ContainerResourceControl
	// ServiceResourceControl represents a resource control associated to a Docker service
	ServiceResourceControl
	// VolumeResourceControl represents a resource control associated to a Docker volume
	VolumeResourceControl
	// NetworkResourceControl represents a resource control associated to a Docker network
	NetworkResourceControl
	// SecretResourceControl represents a resource control associated to a Docker secret
	SecretResourceControl
	// StackResourceControl represents a resource control associated to a stack composed of Docker services
	StackResourceControl
	// ConfigResourceControl represents a resource control associated to a Docker config
	ConfigResourceControl
	// CustomTemplateResourceControl represents a resource control associated to a custom template
	CustomTemplateResourceControl
	// ContainerGroupResourceControl represents a resource control associated to an Azure container group
	ContainerGroupResourceControl
)

const (
	_ portainer.StackType = iota
	// DockerSwarmStack represents a stack managed via docker stack
	DockerSwarmStack
	// DockerComposeStack represents a stack managed via docker-compose
	DockerComposeStack
	// KubernetesStack represents a stack managed via kubectl
	KubernetesStack
	// NomadStack represents a stack managed via Nomad
	NomadStack
)

// StackStatus represents a status for a stack
const (
	_ portainer.StackStatus = iota
	StackStatusActive
	StackStatusInactive
)

const (
	_ portainer.UserRole = iota
	// AdministratorRole represents an administrator user role
	AdministratorRole
	// StandardUserRole represents a regular user role
	StandardUserRole
)

const (
	_ portainer.RoleID = iota
	// RoleIDEndpointAdmin represents environment(endpoint) admin role id
	RoleIDEndpointAdmin
	// RoleIDHelpdesk represents help desk role id
	RoleIDHelpdesk
	// RoleIDStandardUser represents standard user role id
	RoleIDStandardUser
	// RoleIDReadonly represents readonly role id
	RoleIDReadonly
	// RoleIDOperator represents operator role id
	RoleIDOperator
)

const (
	_ portainer.WebhookType = iota
	// ServiceWebhook is a webhook for restarting a docker service
	ServiceWebhook
	// ContainerWebhook is a webhook for recreating a docker container
	ContainerWebhook
)

const (
	// EdgeAgentIdle represents an idle state for a tunnel connected to an Edge environment(endpoint).
	EdgeAgentIdle string = "IDLE"
	// EdgeAgentManagementRequired represents a required state for a tunnel connected to an Edge environment(endpoint)
	EdgeAgentManagementRequired string = "REQUIRED"
	// EdgeAgentActive represents an active state for a tunnel connected to an Edge environment(endpoint)
	EdgeAgentActive string = "ACTIVE"
)

const (
	EdgeAsyncCommandTypeConfig      EdgeAsyncCommandType = "edgeConfig"
	EdgeAsyncCommandTypeStack       EdgeAsyncCommandType = "edgeStack"
	EdgeAsyncCommandTypeJob         EdgeAsyncCommandType = "edgeJob"
	EdgeAsyncCommandTypeLog         EdgeAsyncCommandType = "edgeLog"
	EdgeAsyncCommandTypeContainer   EdgeAsyncCommandType = "container"
	EdgeAsyncCommandTypeImage       EdgeAsyncCommandType = "image"
	EdgeAsyncCommandTypeVolume      EdgeAsyncCommandType = "volume"
	EdgeAsyncCommandTypeNormalStack EdgeAsyncCommandType = "normalStack"

	EdgeAsyncCommandOpAdd     EdgeAsyncCommandOperation = "add"
	EdgeAsyncCommandOpRemove  EdgeAsyncCommandOperation = "remove"
	EdgeAsyncCommandOpReplace EdgeAsyncCommandOperation = "replace"

	EdgeAsyncContainerOperationStart   EdgeAsyncContainerOperation = "start"
	EdgeAsyncContainerOperationRestart EdgeAsyncContainerOperation = "restart"
	EdgeAsyncContainerOperationStop    EdgeAsyncContainerOperation = "stop"
	EdgeAsyncContainerOperationDelete  EdgeAsyncContainerOperation = "delete"
	EdgeAsyncContainerOperationKill    EdgeAsyncContainerOperation = "kill"

	EdgeAsyncImageOperationDelete EdgeAsyncImageOperation = "delete"

	EdgeAsyncVolumeOperationDelete EdgeAsyncVolumeOperation = "delete"

	EdgeAsyncNormalStackOperationRemove EdgeAsyncNormalStackOperation = "remove"
)

const (
	// K8sRoleClusterAdmin is a built in k8s role
	K8sRoleClusterAdmin K8sRole = "cluster-admin"
	// K8sRolePortainerBasic is a portainer k8s role at cluster level
	K8sRolePortainerBasic K8sRole = "portainer-basic"
	// K8sRolePortainerHelpdesk is a portainer k8s role at cluster level
	K8sRolePortainerHelpdesk K8sRole = "portainer-helpdesk"
	// K8sRolePortainerOperator is a portainer k8s role at cluster level
	K8sRolePortainerOperator K8sRole = "portainer-operator"
	// K8sRolePortainerView is a portainer k8s role at namespace level
	K8sRolePortainerView K8sRole = "portainer-view"
	// K8sRolePortainerEdit is a portainer k8s role at namespace level
	K8sRolePortainerEdit K8sRole = "portainer-edit"
)

// represents an portainer.Authorization type
const (
	OperationDockerContainerArchiveInfo         portainer.Authorization = "DockerContainerArchiveInfo"
	OperationDockerContainerList                portainer.Authorization = "DockerContainerList"
	OperationDockerContainerExport              portainer.Authorization = "DockerContainerExport"
	OperationDockerContainerChanges             portainer.Authorization = "DockerContainerChanges"
	OperationDockerContainerInspect             portainer.Authorization = "DockerContainerInspect"
	OperationDockerContainerTop                 portainer.Authorization = "DockerContainerTop"
	OperationDockerContainerLogs                portainer.Authorization = "DockerContainerLogs"
	OperationDockerContainerStats               portainer.Authorization = "DockerContainerStats"
	OperationDockerContainerAttachWebsocket     portainer.Authorization = "DockerContainerAttachWebsocket"
	OperationDockerContainerArchive             portainer.Authorization = "DockerContainerArchive"
	OperationDockerContainerCreate              portainer.Authorization = "DockerContainerCreate"
	OperationDockerContainerPrune               portainer.Authorization = "DockerContainerPrune"
	OperationDockerContainerKill                portainer.Authorization = "DockerContainerKill"
	OperationDockerContainerPause               portainer.Authorization = "DockerContainerPause"
	OperationDockerContainerUnpause             portainer.Authorization = "DockerContainerUnpause"
	OperationDockerContainerRestart             portainer.Authorization = "DockerContainerRestart"
	OperationDockerContainerStart               portainer.Authorization = "DockerContainerStart"
	OperationDockerContainerStop                portainer.Authorization = "DockerContainerStop"
	OperationDockerContainerWait                portainer.Authorization = "DockerContainerWait"
	OperationDockerContainerResize              portainer.Authorization = "DockerContainerResize"
	OperationDockerContainerAttach              portainer.Authorization = "DockerContainerAttach"
	OperationDockerContainerExec                portainer.Authorization = "DockerContainerExec"
	OperationDockerContainerRename              portainer.Authorization = "DockerContainerRename"
	OperationDockerContainerUpdate              portainer.Authorization = "DockerContainerUpdate"
	OperationDockerContainerPutContainerArchive portainer.Authorization = "DockerContainerPutContainerArchive"
	OperationDockerContainerDelete              portainer.Authorization = "DockerContainerDelete"
	OperationDockerImageList                    portainer.Authorization = "DockerImageList"
	OperationDockerImageSearch                  portainer.Authorization = "DockerImageSearch"
	OperationDockerImageGetAll                  portainer.Authorization = "DockerImageGetAll"
	OperationDockerImageGet                     portainer.Authorization = "DockerImageGet"
	OperationDockerImageHistory                 portainer.Authorization = "DockerImageHistory"
	OperationDockerImageInspect                 portainer.Authorization = "DockerImageInspect"
	OperationDockerImageLoad                    portainer.Authorization = "DockerImageLoad"
	OperationDockerImageCreate                  portainer.Authorization = "DockerImageCreate"
	OperationDockerImagePrune                   portainer.Authorization = "DockerImagePrune"
	OperationDockerImagePush                    portainer.Authorization = "DockerImagePush"
	OperationDockerImageTag                     portainer.Authorization = "DockerImageTag"
	OperationDockerImageDelete                  portainer.Authorization = "DockerImageDelete"
	OperationDockerImageCommit                  portainer.Authorization = "DockerImageCommit"
	OperationDockerImageBuild                   portainer.Authorization = "DockerImageBuild"
	OperationDockerNetworkList                  portainer.Authorization = "DockerNetworkList"
	OperationDockerNetworkInspect               portainer.Authorization = "DockerNetworkInspect"
	OperationDockerNetworkCreate                portainer.Authorization = "DockerNetworkCreate"
	OperationDockerNetworkConnect               portainer.Authorization = "DockerNetworkConnect"
	OperationDockerNetworkDisconnect            portainer.Authorization = "DockerNetworkDisconnect"
	OperationDockerNetworkPrune                 portainer.Authorization = "DockerNetworkPrune"
	OperationDockerNetworkDelete                portainer.Authorization = "DockerNetworkDelete"
	OperationDockerVolumeList                   portainer.Authorization = "DockerVolumeList"
	OperationDockerVolumeInspect                portainer.Authorization = "DockerVolumeInspect"
	OperationDockerVolumeCreate                 portainer.Authorization = "DockerVolumeCreate"
	OperationDockerVolumePrune                  portainer.Authorization = "DockerVolumePrune"
	OperationDockerVolumeDelete                 portainer.Authorization = "DockerVolumeDelete"
	OperationDockerExecInspect                  portainer.Authorization = "DockerExecInspect"
	OperationDockerExecStart                    portainer.Authorization = "DockerExecStart"
	OperationDockerExecResize                   portainer.Authorization = "DockerExecResize"
	OperationDockerSwarmInspect                 portainer.Authorization = "DockerSwarmInspect"
	OperationDockerSwarmUnlockKey               portainer.Authorization = "DockerSwarmUnlockKey"
	OperationDockerSwarmInit                    portainer.Authorization = "DockerSwarmInit"
	OperationDockerSwarmJoin                    portainer.Authorization = "DockerSwarmJoin"
	OperationDockerSwarmLeave                   portainer.Authorization = "DockerSwarmLeave"
	OperationDockerSwarmUpdate                  portainer.Authorization = "DockerSwarmUpdate"
	OperationDockerSwarmUnlock                  portainer.Authorization = "DockerSwarmUnlock"
	OperationDockerNodeList                     portainer.Authorization = "DockerNodeList"
	OperationDockerNodeInspect                  portainer.Authorization = "DockerNodeInspect"
	OperationDockerNodeUpdate                   portainer.Authorization = "DockerNodeUpdate"
	OperationDockerNodeDelete                   portainer.Authorization = "DockerNodeDelete"
	OperationDockerServiceList                  portainer.Authorization = "DockerServiceList"
	OperationDockerServiceInspect               portainer.Authorization = "DockerServiceInspect"
	OperationDockerServiceLogs                  portainer.Authorization = "DockerServiceLogs"
	OperationDockerServiceCreate                portainer.Authorization = "DockerServiceCreate"
	OperationDockerServiceUpdate                portainer.Authorization = "DockerServiceUpdate"
	OperationDockerServiceDelete                portainer.Authorization = "DockerServiceDelete"
	OperationDockerServiceForceUpdateService    portainer.Authorization = "DockerServiceForceUpdateService"
	OperationDockerSecretList                   portainer.Authorization = "DockerSecretList"
	OperationDockerSecretInspect                portainer.Authorization = "DockerSecretInspect"
	OperationDockerSecretCreate                 portainer.Authorization = "DockerSecretCreate"
	OperationDockerSecretUpdate                 portainer.Authorization = "DockerSecretUpdate"
	OperationDockerSecretDelete                 portainer.Authorization = "DockerSecretDelete"
	OperationDockerConfigList                   portainer.Authorization = "DockerConfigList"
	OperationDockerConfigInspect                portainer.Authorization = "DockerConfigInspect"
	OperationDockerConfigCreate                 portainer.Authorization = "DockerConfigCreate"
	OperationDockerConfigUpdate                 portainer.Authorization = "DockerConfigUpdate"
	OperationDockerConfigDelete                 portainer.Authorization = "DockerConfigDelete"
	OperationDockerTaskList                     portainer.Authorization = "DockerTaskList"
	OperationDockerTaskInspect                  portainer.Authorization = "DockerTaskInspect"
	OperationDockerTaskLogs                     portainer.Authorization = "DockerTaskLogs"
	OperationDockerPluginList                   portainer.Authorization = "DockerPluginList"
	OperationDockerPluginPrivileges             portainer.Authorization = "DockerPluginPrivileges"
	OperationDockerPluginInspect                portainer.Authorization = "DockerPluginInspect"
	OperationDockerPluginPull                   portainer.Authorization = "DockerPluginPull"
	OperationDockerPluginCreate                 portainer.Authorization = "DockerPluginCreate"
	OperationDockerPluginEnable                 portainer.Authorization = "DockerPluginEnable"
	OperationDockerPluginDisable                portainer.Authorization = "DockerPluginDisable"
	OperationDockerPluginPush                   portainer.Authorization = "DockerPluginPush"
	OperationDockerPluginUpgrade                portainer.Authorization = "DockerPluginUpgrade"
	OperationDockerPluginSet                    portainer.Authorization = "DockerPluginSet"
	OperationDockerPluginDelete                 portainer.Authorization = "DockerPluginDelete"
	OperationDockerSessionStart                 portainer.Authorization = "DockerSessionStart"
	OperationDockerDistributionInspect          portainer.Authorization = "DockerDistributionInspect"
	OperationDockerBuildPrune                   portainer.Authorization = "DockerBuildPrune"
	OperationDockerBuildCancel                  portainer.Authorization = "DockerBuildCancel"
	OperationDockerPing                         portainer.Authorization = "DockerPing"
	OperationDockerInfo                         portainer.Authorization = "DockerInfo"
	OperationDockerEvents                       portainer.Authorization = "DockerEvents"
	OperationDockerSystem                       portainer.Authorization = "DockerSystem"
	OperationDockerVersion                      portainer.Authorization = "DockerVersion"

	OperationDockerAgentPing         portainer.Authorization = "DockerAgentPing"
	OperationDockerAgentList         portainer.Authorization = "DockerAgentList"
	OperationDockerAgentHostInfo     portainer.Authorization = "DockerAgentHostInfo"
	OperationDockerAgentBrowseDelete portainer.Authorization = "DockerAgentBrowseDelete"
	OperationDockerAgentBrowseGet    portainer.Authorization = "DockerAgentBrowseGet"
	OperationDockerAgentBrowseList   portainer.Authorization = "DockerAgentBrowseList"
	OperationDockerAgentBrowsePut    portainer.Authorization = "DockerAgentBrowsePut"
	OperationDockerAgentBrowseRename portainer.Authorization = "DockerAgentBrowseRename"

	OperationAzureSubscriptionsList    portainer.Authorization = "AzureSubscriptionsList"
	OperationAzureSubscriptionGet      portainer.Authorization = "AzureSubscriptionGet"
	OperationAzureProviderGet          portainer.Authorization = "AzureProviderGet"
	OperationAzureResourceGroupsList   portainer.Authorization = "AzureResourceGroupsList"
	OperationAzureResourceGroupGet     portainer.Authorization = "AzureResourceGroupGet"
	OperationAzureContainerGroupsList  portainer.Authorization = "AzureContainerGroupsList"
	OperationAzureContainerGroupGet    portainer.Authorization = "AzureContainerGroupGet"
	OperationAzureContainerGroupCreate portainer.Authorization = "AzureContainerGroupCreate"
	OperationAzureContainerGroupDelete portainer.Authorization = "AzureContainerGroupDelete"

	OperationPortainerDockerHubInspect       portainer.Authorization = "PortainerDockerHubInspect"
	OperationPortainerDockerHubUpdate        portainer.Authorization = "PortainerDockerHubUpdate"
	OperationPortainerEndpointGroupCreate    portainer.Authorization = "PortainerEndpointGroupCreate"
	OperationPortainerEndpointGroupList      portainer.Authorization = "PortainerEndpointGroupList"
	OperationPortainerEndpointGroupDelete    portainer.Authorization = "PortainerEndpointGroupDelete"
	OperationPortainerEndpointGroupInspect   portainer.Authorization = "PortainerEndpointGroupInspect"
	OperationPortainerEndpointGroupUpdate    portainer.Authorization = "PortainerEndpointGroupEdit"
	OperationPortainerEndpointGroupAccess    portainer.Authorization = "PortainerEndpointGroupAccess "
	OperationPortainerEndpointList           portainer.Authorization = "PortainerEndpointList"
	OperationPortainerEndpointInspect        portainer.Authorization = "PortainerEndpointInspect"
	OperationPortainerEndpointCreate         portainer.Authorization = "PortainerEndpointCreate"
	OperationPortainerEndpointJob            portainer.Authorization = "PortainerEndpointJob"
	OperationPortainerEndpointSnapshots      portainer.Authorization = "PortainerEndpointSnapshots"
	OperationPortainerEndpointSnapshot       portainer.Authorization = "PortainerEndpointSnapshot"
	OperationPortainerEndpointUpdate         portainer.Authorization = "PortainerEndpointUpdate"
	OperationPortainerEndpointUpdateAccess   portainer.Authorization = "PortainerEndpointUpdateAccess"
	OperationPortainerEndpointUpdateSettings portainer.Authorization = "PortainerEndpointUpdateSettings"
	OperationPortainerEndpointDelete         portainer.Authorization = "PortainerEndpointDelete"
	OperationPortainerExtensionList          portainer.Authorization = "PortainerExtensionList"
	OperationPortainerExtensionInspect       portainer.Authorization = "PortainerExtensionInspect"
	OperationPortainerExtensionCreate        portainer.Authorization = "PortainerExtensionCreate"
	OperationPortainerExtensionUpdate        portainer.Authorization = "PortainerExtensionUpdate"
	OperationPortainerExtensionDelete        portainer.Authorization = "PortainerExtensionDelete"
	OperationPortainerMOTD                   portainer.Authorization = "PortainerMOTD"
	OperationPortainerRegistryList           portainer.Authorization = "PortainerRegistryList"
	OperationPortainerRegistryInspect        portainer.Authorization = "PortainerRegistryInspect"
	OperationPortainerRegistryCreate         portainer.Authorization = "PortainerRegistryCreate"
	OperationPortainerRegistryConfigure      portainer.Authorization = "PortainerRegistryConfigure"
	OperationPortainerRegistryUpdate         portainer.Authorization = "PortainerRegistryUpdate"
	OperationPortainerRegistryUpdateAccess   portainer.Authorization = "PortainerRegistryUpdateAccess"
	OperationPortainerRegistryDelete         portainer.Authorization = "PortainerRegistryDelete"
	OperationPortainerRegistryInternalUpdate portainer.Authorization = "PortainerRegistryInternalUpdate"
	OperationPortainerRegistryInternalDelete portainer.Authorization = "PortainerRegistryInternalDelete"
	OperationPortainerResourceControlCreate  portainer.Authorization = "PortainerResourceControlCreate"
	OperationPortainerResourceControlUpdate  portainer.Authorization = "PortainerResourceControlUpdate"
	OperationPortainerResourceControlDelete  portainer.Authorization = "PortainerResourceControlDelete"
	OperationPortainerRoleList               portainer.Authorization = "PortainerRoleList"
	OperationPortainerRoleInspect            portainer.Authorization = "PortainerRoleInspect"
	OperationPortainerRoleCreate             portainer.Authorization = "PortainerRoleCreate"
	OperationPortainerRoleUpdate             portainer.Authorization = "PortainerRoleUpdate"
	OperationPortainerRoleDelete             portainer.Authorization = "PortainerRoleDelete"
	OperationPortainerScheduleList           portainer.Authorization = "PortainerScheduleList"
	OperationPortainerScheduleInspect        portainer.Authorization = "PortainerScheduleInspect"
	OperationPortainerScheduleFile           portainer.Authorization = "PortainerScheduleFile"
	OperationPortainerScheduleTasks          portainer.Authorization = "PortainerScheduleTasks"
	OperationPortainerScheduleCreate         portainer.Authorization = "PortainerScheduleCreate"
	OperationPortainerScheduleUpdate         portainer.Authorization = "PortainerScheduleUpdate"
	OperationPortainerScheduleDelete         portainer.Authorization = "PortainerScheduleDelete"
	OperationPortainerSettingsInspect        portainer.Authorization = "PortainerSettingsInspect"
	OperationPortainerSettingsUpdate         portainer.Authorization = "PortainerSettingsUpdate"
	OperationPortainerSettingsLDAPCheck      portainer.Authorization = "PortainerSettingsLDAPCheck"
	OperationPortainerStackList              portainer.Authorization = "PortainerStackList"
	OperationPortainerStackInspect           portainer.Authorization = "PortainerStackInspect"
	OperationPortainerStackFile              portainer.Authorization = "PortainerStackFile"
	OperationPortainerStackCreate            portainer.Authorization = "PortainerStackCreate"
	OperationPortainerStackMigrate           portainer.Authorization = "PortainerStackMigrate"
	OperationPortainerStackUpdate            portainer.Authorization = "PortainerStackUpdate"
	OperationPortainerStackDelete            portainer.Authorization = "PortainerStackDelete"
	OperationPortainerTagList                portainer.Authorization = "PortainerTagList"
	OperationPortainerTagCreate              portainer.Authorization = "PortainerTagCreate"
	OperationPortainerTagDelete              portainer.Authorization = "PortainerTagDelete"
	OperationPortainerTeamMembershipList     portainer.Authorization = "PortainerTeamMembershipList"
	OperationPortainerTeamMembershipCreate   portainer.Authorization = "PortainerTeamMembershipCreate"
	OperationPortainerTeamMembershipUpdate   portainer.Authorization = "PortainerTeamMembershipUpdate"
	OperationPortainerTeamMembershipDelete   portainer.Authorization = "PortainerTeamMembershipDelete"
	OperationPortainerTeamList               portainer.Authorization = "PortainerTeamList"
	OperationPortainerTeamInspect            portainer.Authorization = "PortainerTeamInspect"
	OperationPortainerTeamMemberships        portainer.Authorization = "PortainerTeamMemberships"
	OperationPortainerTeamCreate             portainer.Authorization = "PortainerTeamCreate"
	OperationPortainerTeamUpdate             portainer.Authorization = "PortainerTeamUpdate"
	OperationPortainerTeamDelete             portainer.Authorization = "PortainerTeamDelete"
	OperationPortainerTemplateList           portainer.Authorization = "PortainerTemplateList"
	OperationPortainerTemplateInspect        portainer.Authorization = "PortainerTemplateInspect"
	OperationPortainerTemplateCreate         portainer.Authorization = "PortainerTemplateCreate"
	OperationPortainerTemplateUpdate         portainer.Authorization = "PortainerTemplateUpdate"
	OperationPortainerTemplateDelete         portainer.Authorization = "PortainerTemplateDelete"
	OperationPortainerUploadTLS              portainer.Authorization = "PortainerUploadTLS"
	OperationPortainerUserList               portainer.Authorization = "PortainerUserList"
	OperationPortainerUserInspect            portainer.Authorization = "PortainerUserInspect"
	OperationPortainerUserMemberships        portainer.Authorization = "PortainerUserMemberships"
	OperationPortainerUserCreate             portainer.Authorization = "PortainerUserCreate"
	OperationPortainerUserListToken          portainer.Authorization = "PortainerUserListToken"
	OperationPortainerUserCreateToken        portainer.Authorization = "PortainerUserCreateToken"
	OperationPortainerUserRevokeToken        portainer.Authorization = "PortainerUserRevokeToken"
	OperationPortainerUserUpdate             portainer.Authorization = "PortainerUserUpdate"
	OperationPortainerUserUpdatePassword     portainer.Authorization = "PortainerUserUpdatePassword"
	OperationPortainerUserDelete             portainer.Authorization = "PortainerUserDelete"

	OperationPortainerUserListGitCredential    portainer.Authorization = "PortainerUserListGitCredential"
	OperationPortainerUserInspectGitCredential portainer.Authorization = "PortainerUserInspectGitCredential"
	OperationPortainerUserCreateGitCredential  portainer.Authorization = "PortainerUserCreateGitCredential"
	OperationPortainerUserUpdateGitCredential  portainer.Authorization = "PortainerUserUpdateGitCredential"
	OperationPortainerUserDeleteGitCredential  portainer.Authorization = "PortainerUserDeleteGitCredential"

	OperationPortainerWebsocketExec portainer.Authorization = "PortainerWebsocketExec"
	OperationPortainerWebhookList   portainer.Authorization = "PortainerWebhookList"
	OperationPortainerWebhookCreate portainer.Authorization = "PortainerWebhookCreate"
	OperationPortainerWebhookDelete portainer.Authorization = "PortainerWebhookDelete"

	OperationDockerUndefined      portainer.Authorization = "DockerUndefined"
	OperationAzureUndefined       portainer.Authorization = "AzureUndefined"
	OperationDockerAgentUndefined portainer.Authorization = "DockerAgentUndefined"
	OperationPortainerUndefined   portainer.Authorization = "PortainerUndefined"

	EndpointResourcesAccess portainer.Authorization = "EndpointResourcesAccess"

	OperationK8sUndefined portainer.Authorization = "K8sUndefined"
	// OperationK8sAccessAllNamespaces allows to access all namespaces across cluster.
	// Setting this flag ignores other namespace specific settings.
	OperationK8sAccessAllNamespaces portainer.Authorization = "K8sAccessAllNamespaces"
	// OperationK8sAccessSystemNamespaces allow to access system namespaces
	// if the namespace is assigned
	OperationK8sAccessSystemNamespaces portainer.Authorization = "K8sAccessSystemNamespaces"
	// OperationK8sAccessUserNamespaces allows to access user namespaces
	OperationK8sAccessUserNamespaces portainer.Authorization = "K8sAccessUserNamespaces"
	// k8s namespace operations
	OperationK8sAccessNamespaceRead  portainer.Authorization = "K8sAccessNamespaceRead"
	OperationK8sAccessNamespaceWrite portainer.Authorization = "K8sAccessNamespaceWrite"

	// OPA/Gatekeeper
	OperationK8sPodSecurityW portainer.Authorization = "OperationK8sPodSecurityW"

	// k8s cluster operations
	OperationK8sResourcePoolsR                   portainer.Authorization = "K8sResourcePoolsR"
	OperationK8sResourcePoolsW                   portainer.Authorization = "K8sResourcePoolsW"
	OperationK8sResourcePoolDetailsR             portainer.Authorization = "K8sResourcePoolDetailsR"
	OperationK8sResourcePoolDetailsW             portainer.Authorization = "K8sResourcePoolDetailsW"
	OperationK8sResourcePoolsAccessManagementRW  portainer.Authorization = "K8sResourcePoolsAccessManagementRW"
	OperationK8sApplicationsR                    portainer.Authorization = "K8sApplicationsR"
	OperationK8sApplicationsW                    portainer.Authorization = "K8sApplicationsW"
	OperationK8sApplicationDetailsR              portainer.Authorization = "K8sApplicationDetailsR"
	OperationK8sApplicationDetailsW              portainer.Authorization = "K8sApplicationDetailsW"
	OperationK8sApplicationP                     portainer.Authorization = "K8sApplicationsP" // Patching gives the ability to rollout restart or rollout undo deployments, daemonsets and statefulsets
	OperationK8sPodDelete                        portainer.Authorization = "K8sPodDelete"
	OperationK8sApplicationConsoleRW             portainer.Authorization = "K8sApplicationConsoleRW"
	OperationK8sApplicationsAdvancedDeploymentRW portainer.Authorization = "K8sApplicationsAdvancedDeploymentRW"
	OperationK8sConfigMapsR                      portainer.Authorization = "K8sConfigMapsR" // ConfigMaps
	OperationK8sConfigMapsW                      portainer.Authorization = "K8sConfigMapsW" // ConfigMaps
	OperationK8sSecretsR                         portainer.Authorization = "K8sSecretsR"    // Secrets
	OperationK8sSecretsW                         portainer.Authorization = "K8sSecretsW"    // Secrets
	OperationK8sRegistrySecretList               portainer.Authorization = "K8sRegistrySecretList"
	OperationK8sRegistrySecretInspect            portainer.Authorization = "K8sRegistrySecretInspect"
	OperationK8sRegistrySecretUpdate             portainer.Authorization = "K8sRegistrySecretUpdate"
	OperationK8sRegistrySecretDelete             portainer.Authorization = "K8sRegistrySecretDelete"
	OperationK8sVolumesR                         portainer.Authorization = "K8sVolumesR"
	OperationK8sVolumesW                         portainer.Authorization = "K8sVolumesW"
	OperationK8sVolumeDetailsR                   portainer.Authorization = "K8sVolumeDetailsR"
	OperationK8sVolumeDetailsW                   portainer.Authorization = "K8sVolumeDetailsW"
	OperationK8sClusterR                         portainer.Authorization = "K8sClusterR"
	OperationK8sClusterW                         portainer.Authorization = "K8sClusterW"
	OperationK8sClusterNodeR                     portainer.Authorization = "K8sClusterNodeR"
	OperationK8sClusterNodeW                     portainer.Authorization = "K8sClusterNodeW"
	OperationK8sClusterSetupRW                   portainer.Authorization = "K8sClusterSetupRW"
	OperationK8sApplicationErrorDetailsR         portainer.Authorization = "K8sApplicationErrorDetailsR"
	OperationK8sStorageClassDisabledR            portainer.Authorization = "K8sStorageClassDisabledR"

	OperationK8sIngressesR portainer.Authorization = "K8sIngressesR"
	OperationK8sIngressesW portainer.Authorization = "K8sIngressesW"

	OperationK8sServiceAccountsW     portainer.Authorization = "K8sServiceAccountsW"
	OperationK8sServiceAccountsR     portainer.Authorization = "K8sServiceAccountsR"
	OperationK8sClusterRolesW        portainer.Authorization = "K8sClusterRolesW"
	OperationK8sClusterRolesR        portainer.Authorization = "K8sClusterRolesR"
	OperationK8sClusterRoleBindingsW portainer.Authorization = "K8sClusterRoleBindingsW"
	OperationK8sClusterRoleBindingsR portainer.Authorization = "K8sClusterRoleBindingsR"
	OperationK8sRolesW               portainer.Authorization = "K8sRolesW"
	OperationK8sRolesR               portainer.Authorization = "K8sRolesR"
	OperationK8sRoleBindingsW        portainer.Authorization = "K8sRoleBindingsW"
	OperationK8sRoleBindingsR        portainer.Authorization = "K8sRoleBindingsR"

	OperationK8sYAMLR portainer.Authorization = "K8sYAMLR"
	OperationK8sYAMLW portainer.Authorization = "K8sYAMLW"

	OperationK8sServicesR portainer.Authorization = "K8sServicesR"
	OperationK8sServicesW portainer.Authorization = "K8sServicesW"

	// Helm operations
	OperationHelmRepoList       portainer.Authorization = "HelmRepoList"
	OperationHelmRepoCreate     portainer.Authorization = "HelmRepoCreate"
	OperationHelmInstallChart   portainer.Authorization = "HelmInstallChart"
	OperationHelmUninstallChart portainer.Authorization = "HelmUninstallChart"

	// Deprecated operations
	OperationPortainerEndpointExtensionRemove portainer.Authorization = "PortainerEndpointExtensionRemove"
	OperationPortainerEndpointExtensionAdd    portainer.Authorization = "PortainerEndpointExtensionAdd"
)

const (
	AzurePathContainerGroups = "/subscriptions/*/providers/Microsoft.ContainerInstance/containerGroups"
	AzurePathContainerGroup  = "/subscriptions/*/resourceGroups/*/providers/Microsoft.ContainerInstance/containerGroups/*"
)

const (
	CloudProviderCivo              = "civo"
	CloudProviderDigitalOcean      = "digitalocean"
	CloudProviderLinode            = "linode"
	CloudProviderGKE               = "gke"
	CloudProviderKubeConfig        = "kubeconfig"
	CloudProviderAzure             = "azure"
	CloudProviderAmazon            = "amazon"
	CloudProviderMicrok8s          = "microk8s"
	CloudProviderPreinstalledAgent = "agent"
)

// Edge Configs
type (
	EdgeConfigID        int
	EdgeConfigType      int
	EdgeConfigStateType int
	EdgeConfigCategory  string

	EdgeConfigProgress struct {
		Success int `json:"success"`
		Total   int `json:"total"`
	}

	EdgeConfig struct {
		ID           EdgeConfigID            `json:"id"`
		Name         string                  `json:"name"`
		Type         EdgeConfigType          `json:"type"`
		Category     EdgeConfigCategory      `json:"category"`
		State        EdgeConfigStateType     `json:"state"`
		EdgeGroupIDs []portainer.EdgeGroupID `json:"edgeGroupIDs"`
		BaseDir      string                  `json:"baseDir"`
		Created      int64                   `json:"created"`
		CreatedBy    portainer.UserID        `json:"createdBy"`
		Updated      int64                   `json:"updated"`
		UpdatedBy    portainer.UserID        `json:"updatedBy"`
		Progress     EdgeConfigProgress      `json:"progress"`
		Prev         *EdgeConfig             `json:"prev"`
	}

	EdgeConfigState struct {
		EndpointID portainer.EndpointID                 `json:"endpointID"`
		States     map[EdgeConfigID]EdgeConfigStateType `json:"states"`
	}

	EdgeConfigVersion string
)

const (
	EdgeConfigIdleState EdgeConfigStateType = iota
	EdgeConfigFailureState
	EdgeConfigSavingState
	EdgeConfigDeletingState
	EdgeConfigUpdatingState

	EdgeConfigCurrent  EdgeConfigVersion = "cur"
	EdgeConfigPrevious EdgeConfigVersion = "prev"

	EdgeConfigCategoryConfig EdgeConfigCategory = "configuration"
	EdgeConfigCategorySecret EdgeConfigCategory = "secret"
)

func (e EdgeConfigStateType) String() string {
	switch e {
	case EdgeConfigIdleState:
		return "Idle"
	case EdgeConfigFailureState:
		return "Failure"
	case EdgeConfigSavingState:
		return "Saving"
	case EdgeConfigDeletingState:
		return "Deleting"
	case EdgeConfigUpdatingState:
		return "Updating"
	}

	return "N/A"
}

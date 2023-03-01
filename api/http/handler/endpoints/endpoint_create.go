package endpoints

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"gopkg.in/yaml.v3"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/agent"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
)

type SelfContainedConfigExecError struct{}

func (e *SelfContainedConfigExecError) Error() string {
	return "The kubeconfig uses a binary authentication plugin and is therefore not self-contained. Only self-contained kubeconfigs are currently supported."
}

type SelfContainedConfigFileError struct {
	File string
}

func (e *SelfContainedConfigFileError) Error() string {
	return "The kubeconfig is not self-contained: found local file reference to " + e.File +
		": kubeconfig should be created with `kubectl config view --flatten=true --minify=true`"
}

type endpointCreatePayload struct {
	Name                   string
	URL                    string
	EndpointCreationType   endpointCreationEnum
	PublicURL              string
	Gpus                   []portaineree.Pair
	GroupID                int
	TLS                    bool
	TLSSkipVerify          bool
	TLSSkipClientVerify    bool
	TLSCACertFile          []byte
	TLSCertFile            []byte
	TLSKeyFile             []byte
	AzureApplicationID     string
	AzureTenantID          string
	AzureAuthenticationKey string
	TagIDs                 []portaineree.TagID
	EdgeCheckinInterval    int
	IsEdgeDevice           bool
	Edge                   struct {
		AsyncMode           bool
		PingInterval        int
		SnapshotInterval    int
		CommandInterval     int
		TunnelServerAddress string
	}

	KubeConfig string
}

type endpointCreationEnum int

const (
	_ endpointCreationEnum = iota
	localDockerEnvironment
	agentEnvironment
	azureEnvironment
	edgeAgentEnvironment
	localKubernetesEnvironment
	kubeConfigEnvironment
)

func (payload *endpointCreatePayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("invalid environment name")
	}
	payload.Name = name

	endpointCreationType, err := request.RetrieveNumericMultiPartFormValue(r, "EndpointCreationType", false)
	if err != nil || endpointCreationType == 0 {
		return errors.New("invalid environment type value. Value must be one of: 1 (Docker environment), 2 (Agent environment), 3 (Azure environment), 4 (Edge Agent environment) or 5 (Local Kubernetes environment)")
	}
	payload.EndpointCreationType = endpointCreationEnum(endpointCreationType)

	groupID, _ := request.RetrieveNumericMultiPartFormValue(r, "GroupID", true)
	if groupID == 0 {
		groupID = 1
	}
	payload.GroupID = groupID

	var tagIDs []portaineree.TagID
	err = request.RetrieveMultiPartFormJSONValue(r, "TagIds", &tagIDs, true)
	if err != nil {
		return errors.New("invalid TagIds parameter")
	}
	payload.TagIDs = tagIDs
	if payload.TagIDs == nil {
		payload.TagIDs = make([]portaineree.TagID, 0)
	}

	useTLS, _ := request.RetrieveBooleanMultiPartFormValue(r, "TLS", true)
	payload.TLS = useTLS

	if payload.TLS {
		skipTLSServerVerification, _ := request.RetrieveBooleanMultiPartFormValue(r, "TLSSkipVerify", true)
		payload.TLSSkipVerify = skipTLSServerVerification
		skipTLSClientVerification, _ := request.RetrieveBooleanMultiPartFormValue(r, "TLSSkipClientVerify", true)
		payload.TLSSkipClientVerify = skipTLSClientVerification

		if !payload.TLSSkipVerify {
			caCert, _, err := request.RetrieveMultiPartFormFile(r, "TLSCACertFile")
			if err != nil {
				return errors.New("invalid CA certificate file. Ensure that the file is uploaded correctly")
			}
			payload.TLSCACertFile = caCert
		}

		if !payload.TLSSkipClientVerify {
			cert, _, err := request.RetrieveMultiPartFormFile(r, "TLSCertFile")
			if err != nil {
				return errors.New("invalid certificate file. Ensure that the file is uploaded correctly")
			}
			payload.TLSCertFile = cert

			key, _, err := request.RetrieveMultiPartFormFile(r, "TLSKeyFile")
			if err != nil {
				return errors.New("invalid key file. Ensure that the file is uploaded correctly")
			}
			payload.TLSKeyFile = key
		}
	}

	switch payload.EndpointCreationType {
	case kubeConfigEnvironment:
		c, err := validateKubeConfigEnvironment(r)
		if err != nil {
			return err
		}
		payload.KubeConfig = c

	case azureEnvironment:
		azureApplicationID, err := request.RetrieveMultiPartFormValue(r, "AzureApplicationID", false)
		if err != nil {
			return errors.New("invalid Azure application ID")
		}
		payload.AzureApplicationID = azureApplicationID

		azureTenantID, err := request.RetrieveMultiPartFormValue(r, "AzureTenantID", false)
		if err != nil {
			return errors.New("invalid Azure tenant ID")
		}
		payload.AzureTenantID = azureTenantID

		azureAuthenticationKey, err := request.RetrieveMultiPartFormValue(r, "AzureAuthenticationKey", false)
		if err != nil {
			return errors.New("invalid Azure authentication key")
		}
		payload.AzureAuthenticationKey = azureAuthenticationKey

	default:
		endpointURL, err := request.RetrieveMultiPartFormValue(r, "URL", true)
		if err != nil {
			return errors.New("invalid environment URL")
		}
		payload.URL = endpointURL

		publicURL, _ := request.RetrieveMultiPartFormValue(r, "PublicURL", true)
		payload.PublicURL = publicURL
	}

	gpus := make([]portaineree.Pair, 0)
	err = request.RetrieveMultiPartFormJSONValue(r, "Gpus", &gpus, true)
	if err != nil {
		return errors.New("invalid Gpus parameter")
	}
	payload.Gpus = gpus

	checkinInterval, _ := request.RetrieveNumericMultiPartFormValue(r, "CheckinInterval", true)
	payload.EdgeCheckinInterval = checkinInterval

	asyncMode, _ := request.RetrieveBooleanMultiPartFormValue(r, "EdgeAsyncMode", true)
	payload.Edge.AsyncMode = asyncMode

	payload.Edge.PingInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "EdgePingInterval", true)
	payload.Edge.SnapshotInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "EdgeSnapshotInterval", true)
	payload.Edge.CommandInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "EdgeCommandInterval", true)
	payload.Edge.TunnelServerAddress, _ = request.RetrieveMultiPartFormValue(r, "EdgeTunnelServerAddress", true)

	return nil
}

func validateKubeConfigEnvironment(r *http.Request) (string, error) {
	encoded, err := request.RetrieveMultiPartFormValue(r, "KubeConfig", true)
	if err != nil {
		return "", fmt.Errorf("Invalid kubeconfig: %w", err)
	}

	if encoded == "" {
		return "", fmt.Errorf("Missing or invalid kubeconfig")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("KubeConfig could not be decoded")
	}

	// Parse the config as yaml so we can check for the existence of several
	// fields which indicate a dependency on external files which (very
	// likely) do not exist inside the portainer environment.
	type Cluster struct {
		CertificateAuthority string `yaml:"certificate-authority"`
	}

	type Clusters struct {
		Cluster Cluster `yaml:"cluster"`
		Name    string  `yaml:"name"`
	}

	type AuthConfig struct {
		CmdArgs string `yaml:"cmd-args"`
		CmdPath string `yaml:"cmd-path"`
	}

	type AuthProvider struct {
		AuthConfig AuthConfig `yaml:"config"`
	}

	type User struct {
		ClientCertificate string       `yaml:"client-certificate"`
		ClientKey         string       `yaml:"client-key"`
		AuthProvider      AuthProvider `yaml:"auth-provider"`
	}

	type Users struct {
		Name string `yaml:"name"`
		User User   `yaml:"user"`
	}

	type Context struct {
		Cluster string `yaml:"cluster"`
		User    string `yaml:"user"`
	}

	type Contexts struct {
		Context Context `yaml:"context,omitempty"`
		Name    string  `yaml:"name"`
	}

	type Config struct {
		Clusters       []Clusters `yaml:"clusters"`
		Contexts       []Contexts `yaml:"contexts"`
		CurrentContext string     `yaml:"current-context"`
		Users          []Users    `yaml:"users"`
	}

	var config Config
	err = yaml.Unmarshal([]byte(decoded), &config)
	if err != nil {
		return "", errors.New("KubeConfig could not be parsed as yaml")
	}

	// Check if any of the invalid fields are present. These fields refer to
	// local filepaths on the system which uploaded the Kubeconfig which
	// indicates that the config is not self contained and will fail. The user
	// should instead get their kubeconfig with the following:
	//
	// kubectl config view --flatten=true --minify=true
	var ctxUser, ctxCluster string
	for _, v := range config.Contexts {
		if config.CurrentContext == v.Name {
			ctxUser = v.Context.User
			ctxCluster = v.Context.Cluster
		}
	}

	for _, v := range config.Users {
		if v.Name != ctxUser && config.CurrentContext != "" {
			continue
		}

		if v.User.ClientCertificate != "" {
			return "", &SelfContainedConfigFileError{
				File: v.User.ClientCertificate,
			}
		}
		if v.User.ClientKey != "" {
			return "", &SelfContainedConfigFileError{
				File: v.User.ClientKey,
			}
		}
		if v.User.AuthProvider.AuthConfig.CmdArgs != "" {
			return "", &SelfContainedConfigExecError{}
		}
		if v.User.AuthProvider.AuthConfig.CmdPath != "" {
			return "", &SelfContainedConfigExecError{}
		}
	}

	for _, v := range config.Clusters {
		if v.Name != ctxCluster && config.CurrentContext != "" {
			continue
		}

		if v.Cluster.CertificateAuthority != "" {
			return "", &SelfContainedConfigFileError{
				File: v.Cluster.CertificateAuthority,
			}
		}
	}

	return encoded, nil
}

// @id EndpointCreate
// @summary Create a new environment(endpoint)
// @description  Create a new environment(endpoint) that will be used to manage an environment(endpoint).
// @description **Access policy**: administrator
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @accept multipart/form-data
// @produce json
// @param Name formData string true "Name that will be used to identify this environment(endpoint) (example: my-environment)"
// @param EndpointCreationType formData integer true "Environment(Endpoint) type. Value must be one of: 1 (Local Docker environment), 2 (Agent environment), 3 (Azure environment), 4 (Edge agent environment) or 5 (Local Kubernetes Environment" Enum(1,2,3,4,5)
// @param URL formData string false "URL or IP address of a Docker host (example: docker.mydomain.tld:2375). Defaults to local if not specified (Linux: /var/run/docker.sock, Windows: //./pipe/docker_engine)". Cannot be empty if EndpointCreationType is set to 4 (Edge agent environment)
// @param PublicURL formData string false "URL or IP address where exposed containers will be reachable. Defaults to URL if not specified (example: docker.mydomain.tld:2375)"
// @param GroupID formData int false "Environment(Endpoint) group identifier. If not specified will default to 1 (unassigned)."
// @param TLS formData bool false "Require TLS to connect against this environment(endpoint)"
// @param TLSSkipVerify formData bool false "Skip server verification when using TLS"
// @param TLSSkipClientVerify formData bool false "Skip client verification when using TLS"
// @param TLSCACertFile formData file false "TLS CA certificate file"
// @param TLSCertFile formData file false "TLS client certificate file"
// @param TLSKeyFile formData file false "TLS client key file"
// @param AzureApplicationID formData string false "Azure application ID. Required if environment(endpoint) type is set to 3"
// @param AzureTenantID formData string false "Azure tenant ID. Required if environment(endpoint) type is set to 3"
// @param AzureAuthenticationKey formData string false "Azure authentication key. Required if environment(endpoint) type is set to 3"
// @param TagIDs formData []int false "List of tag identifiers to which this environment(endpoint) is associated"
// @param EdgeCheckinInterval formData int false "The check in interval for edge agent (in seconds)"
// @param EdgeTunnelServerAddress formData string true "URL or IP address that will be used to establish a reverse tunnel"
// @param EdgeAsyncMode formData bool false "Enable async mode for edge agent"
// @param Gpus formData array false "List of GPUs"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /endpoints [post]
func (handler *Handler) endpointCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	payload := &endpointCreatePayload{}
	err := payload.Validate(r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	isUnique, err := handler.isNameUnique(payload.Name, 0)
	if err != nil {
		return httperror.InternalServerError("Unable to check if name is unique", err)
	}

	if !isUnique {
		return httperror.NewError(http.StatusConflict, "Name is not unique", nil)
	}

	endpoint, endpointCreationError := handler.createEndpoint(payload)
	if endpointCreationError != nil {
		return endpointCreationError
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge stacks from the database", err)
	}

	relationObject := &portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
		EdgeStacks: map[portaineree.EdgeStackID]bool{},
	}

	if endpointutils.IsEdgeEndpoint(endpoint) {
		endpointGroup, err := handler.DataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
		if err != nil {
			return httperror.InternalServerError("Unable to find an environment group inside the database", err)
		}

		edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve edge groups from the database", err)
		}

		relatedEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
		for _, stackID := range relatedEdgeStacks {
			relationObject.EdgeStacks[stackID] = true

			edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(stackID)
			if err != nil {
				return httperror.InternalServerError("Unable to retrieve edge stack from the database", err)
			}

			err = handler.edgeService.AddStackCommand(endpoint, stackID, edgeStack.ScheduledTime)
			if err != nil {
				return httperror.InternalServerError("Unable to store edge async command into the database", err)
			}
		}
	} else if endpointutils.IsKubernetesEndpoint(endpoint) {
		endpointutils.InitialIngressClassDetection(
			endpoint,
			handler.DataStore.Endpoint(),
			handler.K8sClientFactory,
		)
		endpointutils.InitialMetricsDetection(
			endpoint,
			handler.DataStore.Endpoint(),
			handler.K8sClientFactory,
		)
		endpointutils.InitialStorageDetection(
			endpoint,
			handler.DataStore.Endpoint(),
			handler.K8sClientFactory,
		)
	}

	err = handler.DataStore.EndpointRelation().Create(relationObject)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the relation object inside the database", err)
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	return response.JSON(w, endpoint)
}

func (handler *Handler) createEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	var err error

	switch payload.EndpointCreationType {
	case azureEnvironment:
		return handler.createAzureEndpoint(payload)

	case edgeAgentEnvironment:
		return handler.createEdgeAgentEndpoint(payload)

	case localKubernetesEnvironment:
		return handler.createKubernetesEndpoint(payload)

	case kubeConfigEnvironment:
		return handler.createKubeConfigEndpoint(payload)
	}

	endpointType := portaineree.DockerEnvironment
	var agentVersion string
	if payload.EndpointCreationType == agentEnvironment {
		var tlsConfig *tls.Config
		if payload.TLS {
			tlsConfig, err = crypto.CreateTLSConfigurationFromBytes(payload.TLSCACertFile, payload.TLSCertFile, payload.TLSKeyFile, payload.TLSSkipVerify, payload.TLSSkipClientVerify)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to create TLS configuration", err)
			}
		}

		agentPlatform, version, err := agent.GetAgentVersionAndPlatform(payload.URL, tlsConfig)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to get environment type", err)
		}

		agentVersion = version

		if agentPlatform == portaineree.AgentPlatformDocker {
			endpointType = portaineree.AgentOnDockerEnvironment
		} else if agentPlatform == portaineree.AgentPlatformKubernetes {
			endpointType = portaineree.AgentOnKubernetesEnvironment
			payload.URL = strings.TrimPrefix(payload.URL, "tcp://")
		}
	}

	if payload.TLS {
		return handler.createTLSSecuredEndpoint(payload, endpointType, agentVersion)
	}
	return handler.createUnsecuredEndpoint(payload)
}

func (handler *Handler) createAzureEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	credentials := portaineree.AzureCredentials{
		ApplicationID:     payload.AzureApplicationID,
		TenantID:          payload.AzureTenantID,
		AuthenticationKey: payload.AzureAuthenticationKey,
	}

	httpClient := client.NewHTTPClient()
	_, err := httpClient.ExecuteAzureAuthenticationRequest(&credentials)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to authenticate against Azure", err)
	}

	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:                 portaineree.EndpointID(endpointID),
		Name:               payload.Name,
		URL:                "https://management.azure.com",
		Type:               portaineree.AzureEnvironment,
		GroupID:            portaineree.EndpointGroupID(payload.GroupID),
		PublicURL:          payload.PublicURL,
		Gpus:               payload.Gpus,
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		AzureCredentials:   credentials,
		TagIDs:             payload.TagIDs,
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err = handler.saveEndpointAndUpdateAuthorizations(endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	return endpoint, nil
}

func (handler *Handler) createEdgeAgentEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	apiServerURL := payload.URL
	if apiServerURL == "" {
		if settings.EdgePortainerURL == "" {
			return nil, httperror.InternalServerError("API server URL not set in Edge Compute settings", err)
		}

		apiServerURL = settings.EdgePortainerURL
	}

	tunnelServerAddr := payload.Edge.TunnelServerAddress
	if tunnelServerAddr == "" {
		if settings.Edge.TunnelServerAddress == "" {
			return nil, httperror.InternalServerError("Tunnel server address not set in Edge Compute settings", err)
		}

		tunnelServerAddr = settings.Edge.TunnelServerAddress
	}

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(apiServerURL, tunnelServerAddr, endpointID)

	endpoint := &portaineree.Endpoint{
		ID:      portaineree.EndpointID(endpointID),
		Name:    payload.Name,
		URL:     apiServerURL,
		Type:    portaineree.EdgeAgentOnDockerEnvironment,
		GroupID: portaineree.EndpointGroupID(payload.GroupID),
		Gpus:    payload.Gpus,
		TLSConfig: portaineree.TLSConfiguration{
			TLS: payload.TLS,
		},
		AuthorizedUsers:     []portaineree.UserID{},
		AuthorizedTeams:     []portaineree.TeamID{},
		UserAccessPolicies:  portaineree.UserAccessPolicies{},
		TeamAccessPolicies:  portaineree.TeamAccessPolicies{},
		TagIDs:              payload.TagIDs,
		Status:              portaineree.EndpointStatusUp,
		Snapshots:           []portainer.DockerSnapshot{},
		EdgeKey:             edgeKey,
		EdgeCheckinInterval: payload.EdgeCheckinInterval,
		Kubernetes:          portaineree.KubernetesDefault(),
		IsEdgeDevice:        payload.IsEdgeDevice,
		UserTrusted:         true,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	if payload.Edge.AsyncMode {
		endpoint.Edge.AsyncMode = true
		endpoint.Edge.PingInterval = payload.Edge.PingInterval
		endpoint.Edge.SnapshotInterval = payload.Edge.SnapshotInterval
		endpoint.Edge.CommandInterval = payload.Edge.CommandInterval
	}

	if settings.EnforceEdgeID {
		edgeID, err := uuid.NewV4()
		if err != nil {
			return nil, httperror.InternalServerError("Cannot generate the Edge ID", err)
		}

		endpoint.EdgeID = edgeID.String()
	}

	err = handler.saveEndpointAndUpdateAuthorizations(endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	err = handler.DataStore.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID})
	if err != nil {
		return nil, httperror.InternalServerError("Unable to create a snapshot object for the environment", err)
	}

	return endpoint, nil
}

func (handler *Handler) createUnsecuredEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointType := portaineree.DockerEnvironment

	if payload.URL == "" {
		payload.URL = "unix:///var/run/docker.sock"
		if runtime.GOOS == "windows" {
			payload.URL = "npipe:////./pipe/docker_engine"
		}
	}

	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:        portaineree.EndpointID(endpointID),
		Name:      payload.Name,
		URL:       payload.URL,
		Type:      endpointType,
		GroupID:   portaineree.EndpointGroupID(payload.GroupID),
		PublicURL: payload.PublicURL,
		Gpus:      payload.Gpus,
		TLSConfig: portaineree.TLSConfiguration{
			TLS: false,
		},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             payload.TagIDs,
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),
		IsEdgeDevice:       payload.IsEdgeDevice,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err := handler.snapshotAndPersistEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *Handler) createKubernetesEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	if payload.URL == "" {
		payload.URL = "https://kubernetes.default.svc"
	}

	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()

	endpoint := &portaineree.Endpoint{
		ID:        portaineree.EndpointID(endpointID),
		Name:      payload.Name,
		URL:       payload.URL,
		Type:      portaineree.KubernetesLocalEnvironment,
		GroupID:   portaineree.EndpointGroupID(payload.GroupID),
		PublicURL: payload.PublicURL,
		Gpus:      payload.Gpus,
		TLSConfig: portaineree.TLSConfiguration{
			TLS:           payload.TLS,
			TLSSkipVerify: payload.TLSSkipVerify,
		},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             payload.TagIDs,
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err := handler.snapshotAndPersistEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *Handler) createKubeConfigEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	if payload.URL == "" {
		payload.URL = "https://kubernetes.default.svc"
	}

	// store kubeconfig as secret
	credentials := models.CloudCredential{
		Name:     "kubeconfig",
		Provider: portaineree.CloudProviderKubeConfig,
		Credentials: models.CloudCredentialMap{
			"kubeconfig": payload.KubeConfig,
		},
	}
	err := handler.DataStore.CloudCredential().Create(&credentials)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to create kubeconfig environment", err)
	}

	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:      portaineree.EndpointID(endpointID),
		Name:    payload.Name,
		Type:    portaineree.AgentOnKubernetesEnvironment,
		GroupID: portaineree.EndpointGroupID(payload.GroupID),
		TLSConfig: portaineree.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             payload.TagIDs,
		Status:             portaineree.EndpointStatusProvisioning,
		StatusMessage: portaineree.EndpointStatusMessage{
			Summary: "Importing KubeConfig",
			Detail:  "Importing KubeConfig",
		},
		Snapshots:  []portainer.DockerSnapshot{},
		Kubernetes: portaineree.KubernetesDefault(),

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},

		CloudProvider: &portaineree.CloudProvider{
			CredentialID: credentials.ID,
		},
	}

	err = handler.saveEndpointAndUpdateAuthorizations(endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to create kubeconfig environment", err)
	}

	request := portaineree.CloudProvisioningRequest{
		EndpointID:    endpoint.ID,
		CredentialID:  credentials.ID,
		StartingState: int(cloud.ProvisioningStateWaitingForCluster),
		Provider:      portaineree.CloudProviderKubeConfig,
	}

	handler.cloudClusterSetupService.Request(&request)

	return endpoint, nil
}

func (handler *Handler) createTLSSecuredEndpoint(payload *endpointCreatePayload, endpointType portaineree.EndpointType, agentVersion string) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:        portaineree.EndpointID(endpointID),
		Name:      payload.Name,
		URL:       payload.URL,
		Type:      endpointType,
		GroupID:   portaineree.EndpointGroupID(payload.GroupID),
		PublicURL: payload.PublicURL,
		Gpus:      payload.Gpus,
		TLSConfig: portaineree.TLSConfiguration{
			TLS:           payload.TLS,
			TLSSkipVerify: payload.TLSSkipVerify,
		},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             payload.TagIDs,
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),
		IsEdgeDevice:       payload.IsEdgeDevice,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	endpoint.Agent.Version = agentVersion

	err := handler.storeTLSFiles(endpoint, payload)
	if err != nil {
		return nil, err
	}

	err = handler.snapshotAndPersistEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *Handler) snapshotAndPersistEndpoint(endpoint *portaineree.Endpoint) *httperror.HandlerError {
	err := handler.SnapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		if (endpoint.Type == portaineree.AgentOnDockerEnvironment && strings.Contains(err.Error(), "Invalid request signature")) ||
			(endpoint.Type == portaineree.AgentOnKubernetesEnvironment && strings.Contains(err.Error(), "unknown")) {
			err = errors.New("agent already paired with another Portainer instance")
		}
		return httperror.InternalServerError("Unable to initiate communications with environment", err)
	}

	err = handler.saveEndpointAndUpdateAuthorizations(endpoint)
	if err != nil {
		return httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	return nil
}

func (handler *Handler) saveEndpointAndUpdateAuthorizations(endpoint *portaineree.Endpoint) error {
	endpoint.SecuritySettings = portaineree.EndpointSecuritySettings{
		AllowVolumeBrowserForRegularUsers: false,
		EnableHostManagementFeatures:      false,

		AllowSysctlSettingForRegularUsers:         true,
		AllowBindMountsForRegularUsers:            true,
		AllowPrivilegedModeForRegularUsers:        true,
		AllowHostNamespaceForRegularUsers:         true,
		AllowContainerCapabilitiesForRegularUsers: true,
		AllowDeviceMappingForRegularUsers:         true,
		AllowStackManagementForRegularUsers:       true,
	}

	err := handler.DataStore.Endpoint().Create(endpoint)
	if err != nil {
		return err
	}

	group, err := handler.DataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	if len(group.UserAccessPolicies) > 0 || len(group.TeamAccessPolicies) > 0 {
		err = handler.AuthorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return err
		}
	}

	for _, tagID := range endpoint.TagIDs {
		err = handler.DataStore.Tag().UpdateTagFunc(tagID, func(tag *portaineree.Tag) {
			tag.Endpoints[endpoint.ID] = true
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) storeTLSFiles(endpoint *portaineree.Endpoint, payload *endpointCreatePayload) *httperror.HandlerError {
	folder := strconv.Itoa(int(endpoint.ID))

	if !payload.TLSSkipVerify {
		caCertPath, err := handler.FileService.StoreTLSFileFromBytes(folder, portainer.TLSFileCA, payload.TLSCACertFile)
		if err != nil {
			return httperror.InternalServerError("Unable to persist TLS CA certificate file on disk", err)
		}
		endpoint.TLSConfig.TLSCACertPath = caCertPath
	}

	if !payload.TLSSkipClientVerify {
		certPath, err := handler.FileService.StoreTLSFileFromBytes(folder, portainer.TLSFileCert, payload.TLSCertFile)
		if err != nil {
			return httperror.InternalServerError("Unable to persist TLS certificate file on disk", err)
		}
		endpoint.TLSConfig.TLSCertPath = certPath

		keyPath, err := handler.FileService.StoreTLSFileFromBytes(folder, portainer.TLSFileKey, payload.TLSKeyFile)
		if err != nil {
			return httperror.InternalServerError("Unable to persist TLS key file on disk", err)
		}
		endpoint.TLSConfig.TLSKeyPath = keyPath
	}

	return nil
}

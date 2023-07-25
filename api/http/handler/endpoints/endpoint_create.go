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

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/agent"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/unique"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"

	"github.com/gofrs/uuid"
	"gopkg.in/yaml.v3"
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
	Edge                   struct {
		AsyncMode           bool
		PingInterval        int
		SnapshotInterval    int
		CommandInterval     int
		TunnelServerAddress string
	}
	InitialStatus        portaineree.EndpointStatus
	InitialStatusMessage portaineree.EndpointStatusMessage

	CustomTemplateID      portaineree.CustomTemplateID
	CustomTemplateContent string

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

	edgeCheckinInterval, _ := request.RetrieveNumericMultiPartFormValue(r, "EdgeCheckinInterval", true)
	if edgeCheckinInterval == 0 {
		// deprecated CheckinInterval
		edgeCheckinInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "CheckinInterval", true)
	}
	payload.EdgeCheckinInterval = edgeCheckinInterval

	asyncMode, _ := request.RetrieveBooleanMultiPartFormValue(r, "EdgeAsyncMode", true)
	payload.Edge.AsyncMode = asyncMode

	customTemplateID, _ := request.RetrieveNumericMultiPartFormValue(r, "CustomTemplateID", true)
	payload.CustomTemplateID = portaineree.CustomTemplateID(customTemplateID)

	customTemplateContent, _ := request.RetrieveMultiPartFormValue(r, "CustomTemplateContent", true)
	payload.CustomTemplateContent = customTemplateContent

	payload.Edge.PingInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "EdgePingInterval", true)
	payload.Edge.SnapshotInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "EdgeSnapshotInterval", true)
	payload.Edge.CommandInterval, _ = request.RetrieveNumericMultiPartFormValue(r, "EdgeCommandInterval", true)
	payload.Edge.TunnelServerAddress, _ = request.RetrieveMultiPartFormValue(r, "EdgeTunnelServerAddress", true)

	return nil
}

func validateKubeConfigEnvironment(r *http.Request) (string, error) {
	encoded, err := request.RetrieveMultiPartFormValue(r, "KubeConfig", true)
	if err != nil {
		return "", fmt.Errorf("invalid kubeconfig: %w", err)
	}

	if encoded == "" {
		return "", fmt.Errorf("missing or invalid kubeconfig")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("kubeConfig could not be decoded")
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
// @param EndpointCreationType formData integer true "Environment(Endpoint) type. Value must be one of: 1 (Local Docker environment), 2 (Agent environment), 3 (Azure environment), 4 (Edge agent environment) or 5 (Local Kubernetes Environment)" Enum(1,2,3,4,5)
// @param URL formData string false "URL or IP address of a Docker host (example: docker.mydomain.tld:2375). Defaults to local if not specified (Linux: /var/run/docker.sock, Windows: //./pipe/docker_engine). Cannot be empty if EndpointCreationType is set to 4 (Edge agent environment)"
// @param PublicURL formData string false "URL or IP address where exposed containers will be reachable. Defaults to URL if not specified (example: docker.mydomain.tld:2375)"
// @param GroupID formData int false "Environment(Endpoint) group identifier. If not specified will default to 1 (unassigned)."
// @param TLS formData bool false "Require TLS to connect against this environment(endpoint). Must be true if EndpointCreationType is set to 2 (Agent environment)"
// @param TLSSkipVerify formData bool false "Skip server verification when using TLS. Must be true if EndpointCreationType is set to 2 (Agent environment)"
// @param TLSSkipClientVerify formData bool false "Skip client verification when using TLS. Must be true if EndpointCreationType is set to 2 (Agent environment)"
// @param TLSCACertFile formData file false "TLS CA certificate file"
// @param TLSCertFile formData file false "TLS client certificate file"
// @param TLSKeyFile formData file false "TLS client key file"
// @param AzureApplicationID formData string false "Azure application ID. Required if environment(endpoint) type is set to 3"
// @param AzureTenantID formData string false "Azure tenant ID. Required if environment(endpoint) type is set to 3"
// @param AzureAuthenticationKey formData string false "Azure authentication key. Required if environment(endpoint) type is set to 3"
// @param TagIds formData []int false "List of tag identifiers to which this environment(endpoint) is associated"
// @param EdgeCheckinInterval formData int false "The check in interval for edge agent (in seconds)"
// @param EdgeTunnelServerAddress formData string false "URL or IP address that will be used to establish a reverse tunnel. Required when settings.EnableEdgeComputeFeatures is set to false or when settings.Edge.TunnelServerAddress is not set"
// @param EdgeAsyncMode formData bool false "Enable async mode for edge agent"
// @param Gpus formData string false "List of GPUs - json stringified array of {name, value} structs"
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

	payload.InitialStatus = portaineree.EndpointStatusUp
	payload.InitialStatusMessage = portaineree.EndpointStatusMessage{}
	if payload.CustomTemplateID != 0 {
		payload.InitialStatus = portaineree.EndpointStatusProvisioning
		payload.InitialStatusMessage = portaineree.EndpointStatusMessage{
			Summary: "Deploying Custom Template",
		}
	}

	endpoint, endpointCreationError := handler.createEndpoint(handler.DataStore, payload)
	if endpointCreationError != nil {
		return endpointCreationError
	}

	// Use the KaaS provisioning loop for installing custom templates if required.
	if payload.CustomTemplateID != 0 {
		kaasRequest := &portaineree.CloudProvisioningRequest{
			EndpointID:            endpoint.ID,
			Provider:              portaineree.CloudProviderPreinstalledAgent,
			Name:                  payload.Name,
			CustomTemplateID:      payload.CustomTemplateID,
			CustomTemplateContent: payload.CustomTemplateContent,
		}
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			return httperror.InternalServerError(
				"Unable to retrieve user details from authentication token",
				err,
			)
		}
		kaasRequest.CreatedByUserID = tokenData.ID

		handler.cloudManagementService.SubmitRequest(kaasRequest)
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
		endpointGroup, err := handler.DataStore.EndpointGroup().Read(endpoint.GroupID)
		if err != nil {
			return httperror.InternalServerError("Unable to find an environment group inside the database", err)
		}

		edgeGroups, err := handler.DataStore.EdgeGroup().ReadAll()
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

func (handler *Handler) createEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	var err error

	switch payload.EndpointCreationType {
	case azureEnvironment:
		return handler.createAzureEndpoint(tx, payload)

	case edgeAgentEnvironment:
		return handler.createEdgeAgentEndpoint(tx, payload)

	case localKubernetesEnvironment:
		return handler.createKubernetesEndpoint(tx, payload)

	case kubeConfigEnvironment:
		return handler.createKubeConfigEndpoint(tx, payload)
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
		return handler.createTLSSecuredEndpoint(tx, payload, endpointType, agentVersion)
	}

	return handler.createUnsecuredEndpoint(tx, payload)
}

func (handler *Handler) createAzureEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
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

	endpointID := tx.Endpoint().GetNextIdentifier()
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
		Status:             payload.InitialStatus,
		StatusMessage:      payload.InitialStatusMessage,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err = handler.saveEndpointAndUpdateAuthorizations(tx, endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	return endpoint, nil
}

func (handler *Handler) createEdgeAgentEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := tx.Endpoint().GetNextIdentifier()

	settings, err := tx.Settings().Settings()
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
		Status:              payload.InitialStatus,
		StatusMessage:       payload.InitialStatusMessage,
		Snapshots:           []portainer.DockerSnapshot{},
		EdgeKey:             edgeKey,
		EdgeCheckinInterval: payload.EdgeCheckinInterval,
		Kubernetes:          portaineree.KubernetesDefault(),
		UserTrusted:         true,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		EnableImageNotification: true,
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

	err = handler.saveEndpointAndUpdateAuthorizations(tx, endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	err = tx.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID})
	if err != nil {
		return nil, httperror.InternalServerError("Unable to create a snapshot object for the environment", err)
	}

	if err = handler.createEdgeConfigs(tx, endpoint); err != nil {
		return nil, httperror.InternalServerError("Unable to create edge configs", err)
	}

	return endpoint, nil
}

func (handler *Handler) createUnsecuredEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointType := portaineree.DockerEnvironment

	if payload.URL == "" {
		payload.URL = "unix:///var/run/docker.sock"
		if runtime.GOOS == "windows" {
			payload.URL = "npipe:////./pipe/docker_engine"
		}
	}

	endpointID := tx.Endpoint().GetNextIdentifier()
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
		Status:             payload.InitialStatus,
		StatusMessage:      payload.InitialStatusMessage,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		EnableImageNotification: true,
	}

	err := handler.snapshotAndPersistEndpoint(tx, endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *Handler) createKubernetesEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	if payload.URL == "" {
		payload.URL = "https://kubernetes.default.svc"
	}

	endpointID := tx.Endpoint().GetNextIdentifier()

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
		Status:             payload.InitialStatus,
		StatusMessage:      payload.InitialStatusMessage,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err := handler.snapshotAndPersistEndpoint(tx, endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *Handler) createKubeConfigEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
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

	err := tx.CloudCredential().Create(&credentials)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to create kubeconfig environment", err)
	}

	endpointID := tx.Endpoint().GetNextIdentifier()
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
			Summary: "Importing Kubeconfig",
			Detail:  "",
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

	err = handler.saveEndpointAndUpdateAuthorizations(tx, endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to create kubeconfig environment", err)
	}

	request := portaineree.CloudProvisioningRequest{
		EndpointID:    endpoint.ID,
		CredentialID:  credentials.ID,
		StartingState: int(cloud.ProvisioningStateWaitingForCluster),
		Provider:      portaineree.CloudProviderKubeConfig,
	}

	handler.cloudManagementService.SubmitRequest(&request)

	return endpoint, nil
}

func (handler *Handler) createTLSSecuredEndpoint(tx dataservices.DataStoreTx, payload *endpointCreatePayload, endpointType portaineree.EndpointType, agentVersion string) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := tx.Endpoint().GetNextIdentifier()
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
		Status:             payload.InitialStatus,
		StatusMessage:      payload.InitialStatusMessage,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		EnableImageNotification: true,
	}

	endpoint.Agent.Version = agentVersion

	err := handler.storeTLSFiles(endpoint, payload)
	if err != nil {
		return nil, err
	}

	err = handler.snapshotAndPersistEndpoint(tx, endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *Handler) snapshotAndPersistEndpoint(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	err := handler.SnapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		if (endpoint.Type == portaineree.AgentOnDockerEnvironment && strings.Contains(err.Error(), "Invalid request signature")) ||
			(endpoint.Type == portaineree.AgentOnKubernetesEnvironment && strings.Contains(err.Error(), "unknown")) {
			err = errors.New("agent already paired with another Portainer instance")
		}
		return httperror.InternalServerError("Unable to initiate communications with environment", err)
	}

	err = handler.saveEndpointAndUpdateAuthorizations(tx, endpoint)
	if err != nil {
		return httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	return nil
}

func (handler *Handler) saveEndpointAndUpdateAuthorizations(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) error {
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

	err := tx.Endpoint().Create(endpoint)
	if err != nil {
		return err
	}

	group, err := tx.EndpointGroup().Read(endpoint.GroupID)
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

func (handler *Handler) createEdgeConfigs(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) error {
	edgeConfigs, err := tx.EdgeConfig().ReadAll()
	if err != nil {
		return err
	}

	endpointGroups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return err
	}

	edgeGroupsToEdgeConfigs := make(map[portaineree.EdgeGroupID][]portaineree.EdgeConfigID)
	for _, edgeConfig := range edgeConfigs {
		for _, edgeGroupID := range edgeConfig.EdgeGroupIDs {
			edgeGroupsToEdgeConfigs[edgeGroupID] = append(edgeGroupsToEdgeConfigs[edgeGroupID], edgeConfig.ID)
		}
	}

	endpoints := []portaineree.Endpoint{*endpoint}

	var edgeConfigsToCreate []portaineree.EdgeConfigID

	for edgeGroupID, edgeConfigIDs := range edgeGroupsToEdgeConfigs {
		edgeGroup, err := tx.EdgeGroup().Read(edgeGroupID)
		if err != nil {
			return err
		}

		relatedEndpointIDs := edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)

		if len(relatedEndpointIDs) < 1 {
			continue
		}

		edgeConfigsToCreate = append(edgeConfigsToCreate, edgeConfigIDs...)
	}

	edgeConfigsToCreate = unique.Unique(edgeConfigsToCreate)

	for _, edgeConfigID := range edgeConfigsToCreate {
		// Update the Edge Config
		edgeConfig, err := tx.EdgeConfig().Read(edgeConfigID)
		if err != nil {
			return err
		}

		switch edgeConfig.State {
		case portaineree.EdgeConfigFailureState, portaineree.EdgeConfigDeletingState:
			continue
		}

		edgeConfig.Progress.Total++

		if err := tx.EdgeConfig().Update(edgeConfigID, edgeConfig); err != nil {
			return err
		}

		// Update or create the Edge Config State
		edgeConfigState, err := tx.EdgeConfigState().Read(endpoint.ID)
		if err != nil {
			edgeConfigState = &portaineree.EdgeConfigState{
				EndpointID: endpoint.ID,
				States:     make(map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType),
			}

			if err := tx.EdgeConfigState().Create(edgeConfigState); err != nil {
				return err
			}
		}

		edgeConfigState.States[edgeConfigID] = portaineree.EdgeConfigSavingState

		if err := tx.EdgeConfigState().Update(edgeConfigState.EndpointID, edgeConfigState); err != nil {
			return err
		}

		dirEntries, err := handler.FileService.GetEdgeConfigDirEntries(edgeConfig, endpoint.EdgeID, portaineree.EdgeConfigCurrent)
		if err != nil {
			return httperror.InternalServerError("Unable to process the files for the edge configuration", err)
		}

		if err = handler.edgeService.AddConfigCommandTx(tx, endpoint.ID, edgeConfig, dirEntries); err != nil {
			return httperror.InternalServerError("Unable to persist the edge configuration command inside the database", err)
		}
	}

	return nil
}

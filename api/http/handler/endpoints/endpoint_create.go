package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/portainer/portainer-ee/api/internal/endpointutils"

	"github.com/gofrs/uuid"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/internal/edge"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
)

type endpointCreatePayload struct {
	Name                   string
	URL                    string
	EndpointCreationType   endpointCreationEnum
	PublicURL              string
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
}

type endpointCreationEnum int

const (
	_ endpointCreationEnum = iota
	localDockerEnvironment
	agentEnvironment
	azureEnvironment
	edgeAgentEnvironment
	localKubernetesEnvironment
)

func (payload *endpointCreatePayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("Invalid environment name")
	}
	payload.Name = name

	endpointCreationType, err := request.RetrieveNumericMultiPartFormValue(r, "EndpointCreationType", false)
	if err != nil || endpointCreationType == 0 {
		return errors.New("Invalid environment type value. Value must be one of: 1 (Docker environment), 2 (Agent environment), 3 (Azure environment), 4 (Edge Agent environment) or 5 (Local Kubernetes environment)")
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
		return errors.New("Invalid TagIds parameter")
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
				return errors.New("Invalid CA certificate file. Ensure that the file is uploaded correctly")
			}
			payload.TLSCACertFile = caCert
		}

		if !payload.TLSSkipClientVerify {
			cert, _, err := request.RetrieveMultiPartFormFile(r, "TLSCertFile")
			if err != nil {
				return errors.New("Invalid certificate file. Ensure that the file is uploaded correctly")
			}
			payload.TLSCertFile = cert

			key, _, err := request.RetrieveMultiPartFormFile(r, "TLSKeyFile")
			if err != nil {
				return errors.New("Invalid key file. Ensure that the file is uploaded correctly")
			}
			payload.TLSKeyFile = key
		}
	}

	switch payload.EndpointCreationType {
	case azureEnvironment:
		azureApplicationID, err := request.RetrieveMultiPartFormValue(r, "AzureApplicationID", false)
		if err != nil {
			return errors.New("Invalid Azure application ID")
		}
		payload.AzureApplicationID = azureApplicationID

		azureTenantID, err := request.RetrieveMultiPartFormValue(r, "AzureTenantID", false)
		if err != nil {
			return errors.New("Invalid Azure tenant ID")
		}
		payload.AzureTenantID = azureTenantID

		azureAuthenticationKey, err := request.RetrieveMultiPartFormValue(r, "AzureAuthenticationKey", false)
		if err != nil {
			return errors.New("Invalid Azure authentication key")
		}
		payload.AzureAuthenticationKey = azureAuthenticationKey
	default:
		endpointURL, err := request.RetrieveMultiPartFormValue(r, "URL", true)
		if err != nil {
			return errors.New("Invalid environment URL")
		}
		payload.URL = endpointURL

		publicURL, _ := request.RetrieveMultiPartFormValue(r, "PublicURL", true)
		payload.PublicURL = publicURL
	}

	checkinInterval, _ := request.RetrieveNumericMultiPartFormValue(r, "CheckinInterval", true)
	payload.EdgeCheckinInterval = checkinInterval

	isEdgeDevice, _ := request.RetrieveBooleanMultiPartFormValue(r, "IsEdgeDevice", true)
	payload.IsEdgeDevice = isEdgeDevice

	return nil
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
// @param URL formData string false "URL or IP address of a Docker host (example: docker.mydomain.tld:2375). Defaults to local if not specified (Linux: /var/run/docker.sock, Windows: //./pipe/docker_engine)"
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
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /endpoints [post]
func (handler *Handler) endpointCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	payload := &endpointCreatePayload{}
	err := payload.Validate(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
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

	endpointGroup, err := handler.dataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment group inside the database", err}
	}

	edgeGroups, err := handler.dataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge groups from the database", err}
	}

	edgeStacks, err := handler.dataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge stacks from the database", err}
	}

	relationObject := &portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
		EdgeStacks: map[portaineree.EdgeStackID]bool{},
	}

	if endpointutils.IsEdgeEndpoint(endpoint) {
		relatedEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
		for _, stackID := range relatedEdgeStacks {
			relationObject.EdgeStacks[stackID] = true

			err = handler.edgeService.AddStackCommand(endpoint, stackID)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to store edge async command into the database", err}
			}
		}
	}

	err = handler.dataStore.EndpointRelation().Create(relationObject)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the relation object inside the database", err}
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	return response.JSON(w, endpoint)
}

func (handler *Handler) createEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	switch payload.EndpointCreationType {
	case azureEnvironment:
		return handler.createAzureEndpoint(payload)

	case edgeAgentEnvironment:
		return handler.createEdgeAgentEndpoint(payload)

	case localKubernetesEnvironment:
		return handler.createKubernetesEndpoint(payload)
	}

	endpointType := portaineree.DockerEnvironment
	if payload.EndpointCreationType == agentEnvironment {
		agentPlatform, err := handler.pingAndCheckPlatform(payload)
		if err != nil {
			return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to get environment type", err}
		}

		if agentPlatform == portaineree.AgentPlatformDocker {
			endpointType = portaineree.AgentOnDockerEnvironment
		} else if agentPlatform == portaineree.AgentPlatformKubernetes {
			endpointType = portaineree.AgentOnKubernetesEnvironment
			payload.URL = strings.TrimPrefix(payload.URL, "tcp://")
		}
	}

	if payload.TLS {
		return handler.createTLSSecuredEndpoint(payload, endpointType)
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
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to authenticate against Azure", err}
	}

	endpointID := handler.dataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:                 portaineree.EndpointID(endpointID),
		Name:               payload.Name,
		URL:                "https://management.azure.com",
		Type:               portaineree.AzureEnvironment,
		GroupID:            portaineree.EndpointGroupID(payload.GroupID),
		PublicURL:          payload.PublicURL,
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
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "An error occurred while trying to create the environment", err}
	}

	return endpoint, nil
}

func (handler *Handler) createEdgeAgentEndpoint(payload *endpointCreatePayload) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := handler.dataStore.Endpoint().GetNextIdentifier()

	portainerHost, err := edge.ParseHostForEdge(payload.URL)
	if err != nil {
		return nil, httperror.BadRequest("Unable to parse host", err)
	}

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(payload.URL, portainerHost, endpointID)

	endpoint := &portaineree.Endpoint{
		ID:      portaineree.EndpointID(endpointID),
		Name:    payload.Name,
		URL:     portainerHost,
		Type:    portaineree.EdgeAgentOnDockerEnvironment,
		GroupID: portaineree.EndpointGroupID(payload.GroupID),
		TLSConfig: portaineree.TLSConfiguration{
			TLS: false,
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

	settings, err := handler.dataStore.Settings().Settings()
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve the settings from the database", err}
	}

	if settings.EnforceEdgeID {
		edgeID, err := uuid.NewV4()
		if err != nil {
			return nil, &httperror.HandlerError{http.StatusInternalServerError, "Cannot generate the Edge ID", err}
		}

		endpoint.EdgeID = edgeID.String()
	}

	err = handler.saveEndpointAndUpdateAuthorizations(endpoint)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "An error occured while trying to create the environment", err}
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

	endpointID := handler.dataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:        portaineree.EndpointID(endpointID),
		Name:      payload.Name,
		URL:       payload.URL,
		Type:      endpointType,
		GroupID:   portaineree.EndpointGroupID(payload.GroupID),
		PublicURL: payload.PublicURL,
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

	endpointID := handler.dataStore.Endpoint().GetNextIdentifier()

	endpoint := &portaineree.Endpoint{
		ID:        portaineree.EndpointID(endpointID),
		Name:      payload.Name,
		URL:       payload.URL,
		Type:      portaineree.KubernetesLocalEnvironment,
		GroupID:   portaineree.EndpointGroupID(payload.GroupID),
		PublicURL: payload.PublicURL,
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

func (handler *Handler) createTLSSecuredEndpoint(payload *endpointCreatePayload, endpointType portaineree.EndpointType) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := handler.dataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:        portaineree.EndpointID(endpointID),
		Name:      payload.Name,
		URL:       payload.URL,
		Type:      endpointType,
		GroupID:   portaineree.EndpointGroupID(payload.GroupID),
		PublicURL: payload.PublicURL,
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
		if strings.Contains(err.Error(), "Invalid request signature") || strings.Contains(err.Error(), "unknown") {
			err = errors.New("agent already paired with another Portainer instance")
		}
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to initiate communications with environment", err}
	}

	err = handler.saveEndpointAndUpdateAuthorizations(endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "An error occured while trying to create the environment", err}
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

	err := handler.dataStore.Endpoint().Create(endpoint)
	if err != nil {
		return err
	}

	group, err := handler.dataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
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
		tag, err := handler.dataStore.Tag().Tag(tagID)
		if err != nil {
			return err
		}

		tag.Endpoints[endpoint.ID] = true

		err = handler.dataStore.Tag().UpdateTag(tagID, tag)
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
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist TLS CA certificate file on disk", err}
		}
		endpoint.TLSConfig.TLSCACertPath = caCertPath
	}

	if !payload.TLSSkipClientVerify {
		certPath, err := handler.FileService.StoreTLSFileFromBytes(folder, portainer.TLSFileCert, payload.TLSCertFile)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist TLS certificate file on disk", err}
		}
		endpoint.TLSConfig.TLSCertPath = certPath

		keyPath, err := handler.FileService.StoreTLSFileFromBytes(folder, portainer.TLSFileKey, payload.TLSKeyFile)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist TLS key file on disk", err}
		}
		endpoint.TLSConfig.TLSKeyPath = keyPath
	}

	return nil
}

func (handler *Handler) pingAndCheckPlatform(payload *endpointCreatePayload) (portaineree.AgentPlatform, error) {
	httpCli := &http.Client{
		Timeout: 3 * time.Second,
	}

	if payload.TLS {
		tlsConfig, err := crypto.CreateTLSConfigurationFromBytes(payload.TLSCACertFile, payload.TLSCertFile, payload.TLSKeyFile, payload.TLSSkipVerify, payload.TLSSkipClientVerify)
		if err != nil {
			return 0, err
		}

		httpCli.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	url, err := url.Parse(fmt.Sprintf("%s/ping", payload.URL))
	if err != nil {
		return 0, err
	}

	url.Scheme = "https"

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return 0, err
	}

	resp, err := httpCli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return 0, fmt.Errorf("Failed request with status %d", resp.StatusCode)
	}

	agentPlatformHeader := resp.Header.Get(portaineree.HTTPResponseAgentPlatform)
	if agentPlatformHeader == "" {
		return 0, errors.New("Agent Platform Header is missing")
	}

	agentPlatformNumber, err := strconv.Atoi(agentPlatformHeader)
	if err != nil {
		return 0, err
	}

	if agentPlatformNumber == 0 {
		return 0, errors.New("Agent platform is invalid")
	}

	return portaineree.AgentPlatform(agentPlatformNumber), nil
}

package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/utils"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	ru "github.com/portainer/portainer-ee/api/http/utils"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainerErrors "github.com/portainer/portainer/api/dataservices/errors"
)

var apiVersionRe = regexp.MustCompile(`(/v[0-9]\.[0-9]*)?`)

type (
	// Transport is a custom transport for Docker API reverse proxy. It allows
	// interception of requests and rewriting of responses.
	Transport struct {
		HTTPTransport        *http.Transport
		endpoint             *portaineree.Endpoint
		dataStore            dataservices.DataStore
		signatureService     portaineree.DigitalSignatureService
		reverseTunnelService portaineree.ReverseTunnelService
		dockerClientFactory  *docker.ClientFactory
		userActivityService  portaineree.UserActivityService
	}

	// TransportParameters is used to create a new Transport
	TransportParameters struct {
		Endpoint             *portaineree.Endpoint
		DataStore            dataservices.DataStore
		SignatureService     portaineree.DigitalSignatureService
		ReverseTunnelService portaineree.ReverseTunnelService
		DockerClientFactory  *docker.ClientFactory
		UserActivityService  portaineree.UserActivityService
	}

	restrictedDockerOperationContext struct {
		isAdmin                bool
		endpointResourceAccess bool
		userID                 portaineree.UserID
		userTeamIDs            []portaineree.TeamID
		resourceControls       []portaineree.ResourceControl
	}

	operationExecutor struct {
		operationContext *restrictedDockerOperationContext
		labelBlackList   []portaineree.Pair
	}
	restrictedOperationRequest func(*http.Response, *operationExecutor) error
	operationRequest           func(*http.Request) error
)

// NewTransport returns a pointer to a new Transport instance.
func NewTransport(parameters *TransportParameters, httpTransport *http.Transport) (*Transport, error) {
	transport := &Transport{
		endpoint:             parameters.Endpoint,
		dataStore:            parameters.DataStore,
		signatureService:     parameters.SignatureService,
		reverseTunnelService: parameters.ReverseTunnelService,
		dockerClientFactory:  parameters.DockerClientFactory,
		HTTPTransport:        httpTransport,
		userActivityService:  parameters.UserActivityService,
	}

	return transport, nil
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport.ProxyDockerRequest(request)
}

// ProxyDockerRequest intercepts a Docker API request and apply logic based
// on the requested operation.
func (transport *Transport) ProxyDockerRequest(request *http.Request) (*http.Response, error) {
	requestPath := apiVersionRe.ReplaceAllString(request.URL.Path, "")
	request.URL.Path = requestPath

	if transport.endpoint.Type == portaineree.AgentOnDockerEnvironment || transport.endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment {
		signature, err := transport.signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
		if err != nil {
			return nil, err
		}

		request.Header.Set(portaineree.PortainerAgentPublicKeyHeader, transport.signatureService.EncodedPublicKey())
		request.Header.Set(portaineree.PortainerAgentSignatureHeader, signature)
	}

	switch {
	case strings.HasPrefix(requestPath, "/configs"):
		return transport.proxyConfigRequest(request)
	case strings.HasPrefix(requestPath, "/containers"):
		return transport.proxyContainerRequest(request)
	case strings.HasPrefix(requestPath, "/services"):
		return transport.proxyServiceRequest(request)
	case strings.HasPrefix(requestPath, "/volumes"):
		return transport.proxyVolumeRequest(request)
	case strings.HasPrefix(requestPath, "/networks"):
		return transport.proxyNetworkRequest(request)
	case strings.HasPrefix(requestPath, "/secrets"):
		return transport.proxySecretRequest(request)
	case strings.HasPrefix(requestPath, "/swarm"):
		return transport.proxySwarmRequest(request)
	case strings.HasPrefix(requestPath, "/nodes"):
		return transport.proxyNodeRequest(request)
	case strings.HasPrefix(requestPath, "/tasks"):
		return transport.proxyTaskRequest(request)
	case strings.HasPrefix(requestPath, "/build"):
		return transport.proxyBuildRequest(request)
	case strings.HasPrefix(requestPath, "/images"):
		return transport.proxyImageRequest(request)
	case strings.HasPrefix(requestPath, "/v2"):
		return transport.proxyAgentRequest(request)
	default:
		return transport.executeDockerRequest(request)
	}
}

func (transport *Transport) executeDockerRequest(request *http.Request) (*http.Response, error) {
	// need a copy of the request body to preserve the original
	body := ru.CopyRequestBody(request)

	response, err := transport.HTTPTransport.RoundTrip(request)
	if err == nil {
		useractivity.LogProxiedActivity(transport.userActivityService, transport.endpoint, response.StatusCode, body, request)
	}

	if transport.endpoint.Type != portaineree.EdgeAgentOnDockerEnvironment {
		return response, err
	}

	if err == nil {
		transport.reverseTunnelService.SetTunnelStatusToActive(transport.endpoint.ID)
	} else {
		transport.reverseTunnelService.SetTunnelStatusToIdle(transport.endpoint.ID)
	}

	return response, err
}

func (transport *Transport) proxyAgentRequest(r *http.Request) (*http.Response, error) {
	requestPath := strings.TrimPrefix(r.URL.Path, "/v2")

	switch {
	case strings.HasPrefix(requestPath, "/browse"):
		// host file browser request
		volumeIDParameter, found := r.URL.Query()["volumeID"]
		if !found || len(volumeIDParameter) < 1 {
			return transport.administratorOperation(r)
		}

		volumeName := volumeIDParameter[0]

		resourceID, err := transport.getVolumeResourceID(volumeName)
		if err != nil {
			return nil, err
		}

		// volume browser request
		return transport.restrictedResourceOperation(r, resourceID, volumeName, portaineree.VolumeResourceControl, true)
	case strings.HasPrefix(requestPath, "/dockerhub"):
		requestPath, registryIdString := path.Split(r.URL.Path)

		registryID, err := strconv.Atoi(registryIdString)
		if err != nil {
			return nil, fmt.Errorf("missing registry id: %w", err)
		}

		r.URL.Path = strings.TrimSuffix(requestPath, "/")

		registry := &portaineree.Registry{
			Type: portaineree.DockerHubRegistry,
		}

		if registryID != 0 {
			registry, err = transport.dataStore.Registry().Registry(portaineree.RegistryID(registryID))
			if err != nil {
				return nil, fmt.Errorf("failed fetching registry: %w", err)
			}
		}

		if registry.Type != portaineree.DockerHubRegistry {
			return nil, errors.New("invalid registry type")
		}

		newBody, err := json.Marshal(registry)
		if err != nil {
			return nil, err
		}

		r.Method = http.MethodPost

		r.Body = ioutil.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

	}

	return transport.executeDockerRequest(r)
}

func (transport *Transport) proxyConfigRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/configs/create":
		return transport.decorateGenericResourceCreationOperation(request, configObjectIdentifier, portaineree.ConfigResourceControl)

	case "/configs":
		return transport.rewriteOperation(request, transport.configListOperation)

	default:
		// assume /configs/{id}
		configID := path.Base(requestPath)

		if request.Method == http.MethodGet {
			return transport.rewriteOperation(request, transport.configInspectOperation)
		} else if request.Method == http.MethodDelete {
			return transport.executeGenericResourceDeletionOperation(request, configID, configID, portaineree.ConfigResourceControl)
		}

		return transport.restrictedResourceOperation(request, configID, configID, portaineree.ConfigResourceControl, false)
	}
}

func (transport *Transport) proxyContainerRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/containers/create":
		return transport.decorateContainerCreationOperation(request, containerObjectIdentifier, portaineree.ContainerResourceControl)

	case "/containers/prune":
		return transport.administratorOperation(request)

	case "/containers/json":
		return transport.rewriteOperationWithLabelFiltering(request, transport.containerListOperation)

	default:
		// This section assumes /containers/**
		if match, _ := path.Match("/containers/*/*", requestPath); match {
			// Handle /containers/{id}/{action} requests
			containerID := path.Base(path.Dir(requestPath))
			action := path.Base(requestPath)

			if action == "json" {
				return transport.rewriteOperation(request, transport.containerInspectOperation)
			}
			return transport.restrictedResourceOperation(request, containerID, containerID, portaineree.ContainerResourceControl, false)
		} else if match, _ := path.Match("/containers/*", requestPath); match {
			// Handle /containers/{id} requests
			containerID := path.Base(requestPath)

			if request.Method == http.MethodDelete {
				return transport.executeGenericResourceDeletionOperation(request, containerID, containerID, portaineree.ContainerResourceControl)
			}

			return transport.restrictedResourceOperation(request, containerID, containerID, portaineree.ContainerResourceControl, false)
		}
		return transport.executeDockerRequest(request)
	}
}

func (transport *Transport) proxyServiceRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/services/create":
		return transport.decorateServiceCreationOperation(request)

	case "/services":
		return transport.rewriteOperation(request, transport.serviceListOperation)

	default:
		// This section assumes /services/**
		if match, _ := path.Match("/services/*/*", requestPath); match {
			// Handle /services/{id}/{action} requests
			serviceID := path.Base(path.Dir(requestPath))
			transport.decorateRegistryAuthenticationHeader(request)
			return transport.restrictedResourceOperation(request, serviceID, serviceID, portaineree.ServiceResourceControl, false)
		} else if match, _ := path.Match("/services/*", requestPath); match {
			// Handle /services/{id} requests
			serviceID := path.Base(requestPath)

			switch request.Method {
			case http.MethodGet:
				return transport.rewriteOperation(request, transport.serviceInspectOperation)
			case http.MethodDelete:
				return transport.executeGenericResourceDeletionOperation(request, serviceID, serviceID, portaineree.ServiceResourceControl)
			}
			return transport.restrictedResourceOperation(request, serviceID, serviceID, portaineree.ServiceResourceControl, false)
		}
		return transport.executeDockerRequest(request)
	}
}

func (transport *Transport) proxyVolumeRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/volumes/create":
		return transport.decorateVolumeResourceCreationOperation(request, portaineree.VolumeResourceControl)

	case "/volumes/prune":
		return transport.administratorOperation(request)

	case "/volumes":
		return transport.rewriteOperation(request, transport.volumeListOperation)

	default:
		// assume /volumes/{name}
		return transport.restrictedVolumeOperation(requestPath, request)
	}
}

func (transport *Transport) proxyNetworkRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/networks/create":
		return transport.decorateGenericResourceCreationOperation(request, networkObjectIdentifier, portaineree.NetworkResourceControl)

	case "/networks":
		return transport.rewriteOperation(request, transport.networkListOperation)

	default:
		// assume /networks/{id}
		networkID := path.Base(requestPath)

		if request.Method == http.MethodGet {
			return transport.rewriteOperation(request, transport.networkInspectOperation)
		} else if request.Method == http.MethodDelete {
			return transport.executeGenericResourceDeletionOperation(request, networkID, networkID, portaineree.NetworkResourceControl)
		}
		return transport.restrictedResourceOperation(request, networkID, networkID, portaineree.NetworkResourceControl, false)
	}
}

func (transport *Transport) proxySecretRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/secrets/create":
		return transport.decorateGenericResourceCreationOperation(request, secretObjectIdentifier, portaineree.SecretResourceControl)

	case "/secrets":
		return transport.rewriteOperation(request, transport.secretListOperation)

	default:
		// assume /secrets/{id}
		secretID := path.Base(requestPath)

		if request.Method == http.MethodGet {
			return transport.rewriteOperation(request, transport.secretInspectOperation)
		} else if request.Method == http.MethodDelete {
			return transport.executeGenericResourceDeletionOperation(request, secretID, secretID, portaineree.SecretResourceControl)
		}
		return transport.restrictedResourceOperation(request, secretID, secretID, portaineree.SecretResourceControl, false)
	}
}

func (transport *Transport) proxyNodeRequest(request *http.Request) (*http.Response, error) {
	requestPath := request.URL.Path

	// assume /nodes/{id}
	if path.Base(requestPath) != "nodes" {
		return transport.administratorOperation(request)
	}

	return transport.executeDockerRequest(request)
}

func (transport *Transport) proxySwarmRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/swarm":
		return transport.rewriteOperation(request, swarmInspectOperation)
	default:
		// assume /swarm/{action}
		return transport.administratorOperation(request)
	}
}

func (transport *Transport) proxyTaskRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/tasks":
		return transport.rewriteOperation(request, transport.taskListOperation)
	default:
		// assume /tasks/{id}
		return transport.executeDockerRequest(request)
	}
}

func (transport *Transport) proxyBuildRequest(request *http.Request) (*http.Response, error) {
	return transport.interceptAndRewriteRequest(request, buildOperation)
}

func (transport *Transport) proxyImageRequest(request *http.Request) (*http.Response, error) {
	switch requestPath := request.URL.Path; requestPath {
	case "/images/create":
		return transport.replaceRegistryAuthenticationHeader(request)
	default:
		if path.Base(requestPath) == "push" && request.Method == http.MethodPost {
			return transport.replaceRegistryAuthenticationHeader(request)
		}
		return transport.executeDockerRequest(request)
	}
}
func (transport *Transport) decorateRegistryAuthenticationHeader(request *http.Request) error {
	accessContext, err := transport.createRegistryAccessContext(request)
	if err != nil {
		return err
	}

	originalHeader := request.Header.Get("X-Registry-Auth")

	if originalHeader != "" {

		decodedHeaderData, err := base64.StdEncoding.DecodeString(originalHeader)
		if err != nil {
			return err
		}

		var originalHeaderData portainerRegistryAuthenticationHeader
		err = json.Unmarshal(decodedHeaderData, &originalHeaderData)
		if err != nil {
			return err
		}

		authenticationHeader, err := createRegistryAuthenticationHeader(transport.dataStore, originalHeaderData.RegistryId, accessContext)
		if err != nil {
			return err
		}

		headerData, err := json.Marshal(authenticationHeader)
		if err != nil {
			return err
		}

		header := base64.StdEncoding.EncodeToString(headerData)

		request.Header.Set("X-Registry-Auth", header)
	}

	return nil
}
func (transport *Transport) replaceRegistryAuthenticationHeader(request *http.Request) (*http.Response, error) {
	transport.decorateRegistryAuthenticationHeader(request)

	return transport.decorateGenericResourceCreationOperation(request, serviceObjectIdentifier, portaineree.ServiceResourceControl)
}

func (transport *Transport) restrictedResourceOperation(request *http.Request, resourceID string, dockerResourceID string, resourceType portaineree.ResourceControlType, volumeBrowseRestrictionCheck bool) (*http.Response, error) {
	var err error
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	if tokenData.Role != portaineree.AdministratorRole {
		user, err := transport.dataStore.User().User(tokenData.ID)
		if err != nil {
			return nil, err
		}

		if volumeBrowseRestrictionCheck {
			securitySettings, err := transport.fetchEndpointSecuritySettings()
			if err != nil {
				return nil, err
			}

			if !securitySettings.AllowVolumeBrowserForRegularUsers {
				// Return access denied for all roles except environment(endpoint)-administrator
				_, userCanBrowse := user.EndpointAuthorizations[transport.endpoint.ID][portaineree.OperationDockerAgentBrowseList]
				if !userCanBrowse {
					return utils.WriteAccessDeniedResponse()
				}
			}
		}

		_, endpointResourceAccess := user.EndpointAuthorizations[transport.endpoint.ID][portaineree.EndpointResourcesAccess]

		if endpointResourceAccess {
			return transport.executeDockerRequest(request)
		}

		teamMemberships, err := transport.dataStore.TeamMembership().TeamMembershipsByUserID(tokenData.ID)
		if err != nil {
			return nil, err
		}

		userTeamIDs := make([]portaineree.TeamID, 0)
		for _, membership := range teamMemberships {
			userTeamIDs = append(userTeamIDs, membership.TeamID)
		}

		resourceControls, err := transport.dataStore.ResourceControl().ResourceControls()
		if err != nil {
			return nil, err
		}

		resourceControl := authorization.GetResourceControlByResourceIDAndType(resourceID, resourceType, resourceControls)
		if resourceControl == nil {
			agentTargetHeader := request.Header.Get(portaineree.PortainerAgentTargetHeader)

			if dockerResourceID == "" {
				dockerResourceID = resourceID
			}

			// This resource was created outside of portainer,
			// is part of a Docker service or part of a Docker Swarm/Compose stack.
			inheritedResourceControl, err := transport.getInheritedResourceControlFromServiceOrStack(dockerResourceID, agentTargetHeader, resourceType, resourceControls)
			if err != nil {
				return nil, err
			}

			if inheritedResourceControl == nil || !authorization.UserCanAccessResource(tokenData.ID, userTeamIDs, inheritedResourceControl) {
				return utils.WriteAccessDeniedResponse()
			}
		}

		if resourceControl != nil && !authorization.UserCanAccessResource(tokenData.ID, userTeamIDs, resourceControl) {
			return utils.WriteAccessDeniedResponse()
		}
	}
	return transport.executeDockerRequest(request)
}

// rewriteOperationWithLabelFiltering will create a new operation context with data that will be used
// to decorate the original request's response as well as retrieve all the black listed labels
// to filter the resources.
func (transport *Transport) rewriteOperationWithLabelFiltering(request *http.Request, operation restrictedOperationRequest) (*http.Response, error) {
	operationContext, err := transport.createOperationContext(request)
	if err != nil {
		return nil, err
	}

	settings, err := transport.dataStore.Settings().Settings()
	if err != nil {
		return nil, err
	}

	executor := &operationExecutor{
		operationContext: operationContext,
		labelBlackList:   settings.BlackListedLabels,
	}

	return transport.executeRequestAndRewriteResponse(request, operation, executor)
}

// rewriteOperation will create a new operation context with data that will be used
// to decorate the original request's response.
func (transport *Transport) rewriteOperation(request *http.Request, operation restrictedOperationRequest) (*http.Response, error) {
	operationContext, err := transport.createOperationContext(request)
	if err != nil {
		return nil, err
	}

	executor := &operationExecutor{
		operationContext: operationContext,
	}

	return transport.executeRequestAndRewriteResponse(request, operation, executor)
}

func (transport *Transport) interceptAndRewriteRequest(request *http.Request, operation operationRequest) (*http.Response, error) {
	err := operation(request)
	if err != nil {
		return nil, err
	}

	return transport.executeDockerRequest(request)
}

// decorateGenericResourceCreationResponse extracts the response as a JSON object, extracts the resource identifier from that object based
// on the resourceIdentifierAttribute parameter then generate a new resource control associated to that resource
// with a random token and rewrites the response by decorating the original response with a ResourceControl object.
// The generic Docker API response format is JSON object:
// https://docs.docker.com/engine/api/v1.37/#operation/ContainerCreate
// https://docs.docker.com/engine/api/v1.37/#operation/NetworkCreate
// https://docs.docker.com/engine/api/v1.37/#operation/VolumeCreate
// https://docs.docker.com/engine/api/v1.37/#operation/ServiceCreate
// https://docs.docker.com/engine/api/v1.37/#operation/SecretCreate
// https://docs.docker.com/engine/api/v1.37/#operation/ConfigCreate
func (transport *Transport) decorateGenericResourceCreationResponse(response *http.Response, resourceIdentifierAttribute string, resourceType portaineree.ResourceControlType, userID portaineree.UserID) error {
	responseObject, err := utils.GetResponseAsJSONObject(response)
	if err != nil {
		return err
	}

	if responseObject[resourceIdentifierAttribute] == nil {
		log.Printf("[ERROR] [proxy,docker]")
		return errors.New("missing identifier in Docker resource creation response")
	}

	resourceID := responseObject[resourceIdentifierAttribute].(string)

	resourceControl, err := transport.createPrivateResourceControl(resourceID, resourceType, userID)
	if err != nil {
		return err
	}

	responseObject = decorateObject(responseObject, resourceControl)

	return utils.RewriteResponse(response, responseObject, http.StatusOK)
}

func (transport *Transport) decorateGenericResourceCreationOperation(request *http.Request, resourceIdentifierAttribute string, resourceType portaineree.ResourceControlType) (*http.Response, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	response, err := transport.executeDockerRequest(request)
	if err != nil {
		return response, err
	}

	if response.StatusCode == http.StatusCreated {
		err = transport.decorateGenericResourceCreationResponse(response, resourceIdentifierAttribute, resourceType, tokenData.ID)
	}

	return response, err
}

func (transport *Transport) executeGenericResourceDeletionOperation(request *http.Request, resourceIdentifierAttribute string, volumeName string, resourceType portaineree.ResourceControlType) (*http.Response, error) {
	response, err := transport.restrictedResourceOperation(request, resourceIdentifierAttribute, volumeName, resourceType, false)
	if err != nil {
		return response, err
	}

	if response.StatusCode == http.StatusNoContent || response.StatusCode == http.StatusOK {
		resourceControl, err := transport.dataStore.ResourceControl().ResourceControlByResourceIDAndType(resourceIdentifierAttribute, resourceType)
		if err != nil {
			if err == portainerErrors.ErrObjectNotFound {
				return response, nil
			}

			return response, err
		}

		if resourceControl != nil {
			err = transport.dataStore.ResourceControl().DeleteResourceControl(resourceControl.ID)
			if err != nil {
				return response, err
			}
		}
	}

	return response, err
}

func (transport *Transport) executeRequestAndRewriteResponse(request *http.Request, operation restrictedOperationRequest, executor *operationExecutor) (*http.Response, error) {
	response, err := transport.executeDockerRequest(request)
	if err != nil {
		return response, err
	}

	err = operation(response, executor)
	return response, err
}

// administratorOperation ensures that the user has administrator privileges
// before executing the original request.
func (transport *Transport) administratorOperation(request *http.Request) (*http.Response, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	if tokenData.Role != portaineree.AdministratorRole {
		return utils.WriteAccessDeniedResponse()
	}

	return transport.executeDockerRequest(request)
}

func (transport *Transport) createRegistryAccessContext(request *http.Request) (*registryAccessContext, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	accessContext := &registryAccessContext{
		isAdmin:    true,
		endpointID: transport.endpoint.ID,
	}

	user, err := transport.dataStore.User().User(tokenData.ID)
	if err != nil {
		return nil, err
	}
	accessContext.user = user

	registries, err := transport.dataStore.Registry().Registries()
	if err != nil {
		return nil, err
	}
	accessContext.registries = registries

	if user.Role != portaineree.AdministratorRole {
		accessContext.isAdmin = false

		teamMemberships, err := transport.dataStore.TeamMembership().TeamMembershipsByUserID(tokenData.ID)
		if err != nil {
			return nil, err
		}

		accessContext.teamMemberships = teamMemberships
	}

	return accessContext, nil
}

func (transport *Transport) createOperationContext(request *http.Request) (*restrictedDockerOperationContext, error) {
	var err error
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	resourceControls, err := transport.dataStore.ResourceControl().ResourceControls()
	if err != nil {
		return nil, err
	}

	operationContext := &restrictedDockerOperationContext{
		isAdmin:                true,
		userID:                 tokenData.ID,
		resourceControls:       resourceControls,
		endpointResourceAccess: false,
	}

	if tokenData.Role != portaineree.AdministratorRole {
		operationContext.isAdmin = false

		user, err := transport.dataStore.User().User(operationContext.userID)
		if err != nil {
			return nil, err
		}

		_, ok := user.EndpointAuthorizations[transport.endpoint.ID][portaineree.EndpointResourcesAccess]
		if ok {
			operationContext.endpointResourceAccess = true
		}

		teamMemberships, err := transport.dataStore.TeamMembership().TeamMembershipsByUserID(tokenData.ID)
		if err != nil {
			return nil, err
		}

		userTeamIDs := make([]portaineree.TeamID, 0)
		for _, membership := range teamMemberships {
			userTeamIDs = append(userTeamIDs, membership.TeamID)
		}
		operationContext.userTeamIDs = userTeamIDs
	}

	return operationContext, nil
}

func (transport *Transport) isAdminOrEndpointAdmin(request *http.Request) (bool, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return false, err
	}

	if tokenData.Role == portaineree.AdministratorRole {
		return true, nil
	}

	user, err := transport.dataStore.User().User(tokenData.ID)
	if err != nil {
		return false, err
	}

	_, endpointResourceAccess := user.EndpointAuthorizations[portaineree.EndpointID(transport.endpoint.ID)][portaineree.EndpointResourcesAccess]

	return endpointResourceAccess, nil
}

func (transport *Transport) fetchEndpointSecuritySettings() (*portaineree.EndpointSecuritySettings, error) {
	endpoint, err := transport.dataStore.Endpoint().Endpoint(portaineree.EndpointID(transport.endpoint.ID))
	if err != nil {
		return nil, err
	}

	return &endpoint.SecuritySettings, nil
}

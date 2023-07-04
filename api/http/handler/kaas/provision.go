package kaas

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/providers"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
)

// @id provisionClusterAzure
// @summary Provision a new KaaS cluster on azure and create an environment
// @description Provision a new KaaS cluster and create an environment.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body providers.AzureProvisionPayload true "KaaS cluster provisioning details"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/azure/provision [post]

// @id provisionClusterGKE
// @summary Provision a new KaaS cluster on GKE and create an environment
// @description Provision a new KaaS cluster and create an environment.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body providers.GKEProvisionPayload true "KaaS cluster provisioning details"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/gke/provision [post]

// @id provisionCluster
// @summary Provision a new KaaS cluster and create an environment
// @description Provision a new KaaS cluster and create an environment.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body providers.DefaultProvisionPayload true "KaaS cluster provisioning details"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/civo/provision [post]
// @router /cloud/digitalocean/provision [post]
// @router /cloud/linode/provision [post]

func (handler *Handler) provisionCluster(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	provider, err := request.RetrieveRouteVariableValue(r, "provider")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	var cloudProvider *portaineree.CloudProvider
	var payload providers.Providers
	switch provider {
	case portaineree.CloudProviderAzure:
		var p providers.AzureProvisionPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)
		payload = &p
	case portaineree.CloudProviderGKE:
		var p providers.GKEProvisionPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)
		payload = &p
	case portaineree.CloudProviderAmazon:
		var p providers.AmazonProvisionPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)
		payload = &p
	case portaineree.CloudProviderCivo, portaineree.CloudProviderDigitalOcean, portaineree.CloudProviderLinode:
		var p providers.DefaultProvisionPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)
		payload = &p
	case portaineree.CloudProviderMicrok8s:
		var p providers.Microk8sProvisionPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)
		payload = &p

	default:
		return httperror.BadRequest("Invalid request payload", fmt.Errorf("invalid cloud provider: %s", provider))
	}

	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	cloudProvider, err = payload.GetCloudProvider(provider)
	if err != nil {
		return httperror.InternalServerError("Unable to create environment", err)
	}

	endpoint, err := handler.createEndpoint(payload.GetEndpointName(), *cloudProvider, payload.GetEnvironmentMetadata())
	if err != nil {
		return httperror.InternalServerError("Unable to create environment", err)
	}

	// Prepare a new CloudProvisioningRequest
	request := payload.GetCloudProvisioningRequest(endpoint.ID, provider)
	request.CreatedByUserID = getUserID(r)

	handler.cloudManagementService.SubmitRequest(request)
	return response.JSON(w, endpoint)
}

func getUserID(r *http.Request) portaineree.UserID {
	tokenData, _ := security.RetrieveTokenData(r)
	return tokenData.ID
}

func (handler *Handler) createEndpoint(name string, provider portaineree.CloudProvider, metadata types.EnvironmentMetadata) (*portaineree.Endpoint, error) {
	endpointID := handler.dataStore.Endpoint().GetNextIdentifier()

	summaryMessage := "Waiting for cloud provider"
	if provider.Name == types.CloudProvidersMap[portaineree.CloudProviderMicrok8s].Name {
		summaryMessage = "Waiting for nodes"
	}

	endpoint := &portaineree.Endpoint{
		ID:      portaineree.EndpointID(endpointID),
		Name:    name,
		Type:    portaineree.AgentOnKubernetesEnvironment,
		GroupID: metadata.GroupId,
		TLSConfig: portaineree.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             metadata.TagIds,
		Status:             portaineree.EndpointStatusProvisioning,
		StatusMessage: portaineree.EndpointStatusMessage{
			Summary: summaryMessage,
		},
		CloudProvider: &provider,
		Snapshots:     []portainer.DockerSnapshot{},
		Kubernetes:    portaineree.KubernetesDefault(),
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err := handler.dataStore.Endpoint().Create(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

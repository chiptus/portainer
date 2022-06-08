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
	portainer "github.com/portainer/portainer/api"
)

// @id provisionKaaSCluster
// @summary Provision a new KaaS cluster and create an environment
// @description Provision a new KaaS cluster and create an environment.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body kaasClusterProvisionPayload true "Kaas cluster provisioning details"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud [post]
func (handler *Handler) provisionKaaSCluster(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	provider, err := request.RetrieveRouteVariableValue(r, "provider")
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid user identifier route variable",
			Err:        err,
		}
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
	default:
		return &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request payload",
			Err:        fmt.Errorf("Invalid cloud provider: %s", provider),
		}
	}
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request payload",
			Err:        err,
		}
	}

	cloudProvider, err = payload.GetCloudProvider(provider)
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to create environment",
			Err:        err,
		}
	}

	endpoint, err := handler.createEndpoint(payload.GetEndpointName(), *cloudProvider, payload.GetEnvironmentMetadata())
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to create environment",
			Err:        err,
		}
	}

	// Prepare a new CloudProvisioningRequest
	request := payload.GetCloudProvisioningRequest(endpoint.ID, provider)

	handler.cloudClusterSetupService.Request(request)
	return response.JSON(w, endpoint)
}

func (handler *Handler) createEndpoint(name string, provider portaineree.CloudProvider, metadata types.EnvironmentMetadata) (*portaineree.Endpoint, error) {
	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()

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
			Summary: "Waiting for cloud provider",
		},
		CloudProvider: &provider,
		Snapshots:     []portainer.DockerSnapshot{},
		Kubernetes:    portaineree.KubernetesDefault(),
		IsEdgeDevice:  false,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err := handler.DataStore.Endpoint().Create(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

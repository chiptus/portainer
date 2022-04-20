package kaas

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type kaasClusterProvisionPayload struct {
	Name              string `validate:"required" example:"myDevCluster"`
	NodeSize          string `validate:"required" example:"g3.small"`
	NodeCount         int    `validate:"required" example:"3"`
	Region            string `validate:"required" example:"NYC1"`
	NetworkID         string `example:"8465fb26-632e-4fa3-bb9b-21c449629026"`
	KubernetesVersion string `validate:"required" example:"1.23"`
}

func (payload *kaasClusterProvisionPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid cluster name")
	}
	if govalidator.IsNull(payload.NodeSize) {
		return errors.New("Invalid node size")
	}
	if payload.NodeCount <= 0 {
		return errors.New("Invalid node count")
	}
	if govalidator.IsNull(payload.Region) {
		return errors.New("Invalid region")
	}
	if govalidator.IsNull(payload.KubernetesVersion) {
		return errors.New("Invalid Kubernetes version")
	}
	return nil
}

// @id KaaSProvisionCluster
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

	var payload kaasClusterProvisionPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request payload",
			Err:        err,
		}
	}

	var cloudProvider portaineree.CloudProvider
	switch provider {
	case portaineree.CloudProviderCivo:
		cloudProvider.Name = "Civo"
		cloudProvider.URL = "https://www.civo.com/login"
	case portaineree.CloudProviderDigitalOcean:
		cloudProvider.Name = "DigitalOcean"
		cloudProvider.URL = "https://cloud.digitalocean.com/login"
	case portaineree.CloudProviderLinode:
		cloudProvider.Name = "Linode"
		cloudProvider.URL = "https://login.linode.com/"
	default:
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to provision Kaas cluster",
			Err:        errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode"),
		}
	}

	cloudProvider.Region = payload.Region
	cloudProvider.Size = payload.NodeSize
	cloudProvider.NodeCount = payload.NodeCount
	cloudProvider.NetworkID = &payload.NetworkID

	endpointID, err := handler.createEndpoint(payload.Name, cloudProvider)
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to create environment",
			Err:        err,
		}
	}

	// Prepare a new CloudProvisioningRequest
	request := portaineree.CloudProvisioningRequest{
		EndpointID:        endpointID,
		Provider:          provider,
		Region:            payload.Region,
		Name:              payload.Name,
		NodeSize:          payload.NodeSize,
		NetworkID:         payload.NetworkID,
		NodeCount:         payload.NodeCount,
		KubernetesVersion: payload.KubernetesVersion,
	}

	handler.cloudClusterSetupService.Request(&request)
	return response.Empty(w)
}

func (handler *Handler) createEndpoint(name string, provider portaineree.CloudProvider) (portaineree.EndpointID, error) {
	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()

	endpoint := &portaineree.Endpoint{
		ID:      portaineree.EndpointID(endpointID),
		Name:    name,
		Type:    portaineree.AgentOnKubernetesEnvironment,
		GroupID: portaineree.EndpointGroupID(1),
		TLSConfig: portaineree.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             []portaineree.TagID{},
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
		return 0, err
	}

	return endpoint.ID, nil
}

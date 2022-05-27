package kaas

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	portainer "github.com/portainer/portainer/api"
)

type EnvironmentMetadata struct {
	GroupId portaineree.EndpointGroupID
	TagIds  []portaineree.TagID
}

type kaasClusterProvisionPayload struct {
	Name              string                   `validate:"required" example:"myDevCluster"`
	NodeSize          string                   `example:"g3.small"`
	NodeCount         int                      `validate:"required" example:"3"`
	CPU               int                      `example:"2"`
	RAM               float64                  `example:"4"`
	HDD               int                      `example:"100"`
	Region            string                   `validate:"required" example:"NYC1"`
	NetworkID         string                   `example:"8465fb26-632e-4fa3-bb9b-21c449629026"`
	KubernetesVersion string                   `validate:"required" example:"1.23"`
	CredentialID      models.CloudCredentialID `validate:"required" example:"1"`
	Meta              EnvironmentMetadata
}

func (payload *kaasClusterProvisionPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid cluster name")
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
	if govalidator.IsNonPositive(float64(payload.CredentialID)) {
		return errors.New("Invalid Credentials")
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

	cloudProvider, ok := CloudProvidersMap[CloudProviderShortName(provider)]
	if !ok {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to provision Kaas cluster",
			Err:        errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode, gke"),
		}
	}

	cloudProvider.Region = payload.Region
	cloudProvider.Size = &payload.NodeSize
	cloudProvider.NodeCount = payload.NodeCount
	cloudProvider.CPU = &payload.CPU
	cloudProvider.RAM = &payload.RAM
	cloudProvider.HDD = &payload.HDD
	cloudProvider.NetworkID = &payload.NetworkID
	cloudProvider.CredentialID = payload.CredentialID

	endpoint, err := handler.createEndpoint(payload.Name, cloudProvider, payload.Meta)
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to create environment",
			Err:        err,
		}
	}

	// Prepare a new CloudProvisioningRequest
	request := portaineree.CloudProvisioningRequest{
		EndpointID:        endpoint.ID,
		Provider:          provider,
		Region:            payload.Region,
		Name:              payload.Name,
		NodeSize:          payload.NodeSize,
		NetworkID:         payload.NetworkID,
		NodeCount:         payload.NodeCount,
		CPU:               payload.CPU,
		RAM:               payload.RAM,
		HDD:               payload.HDD,
		KubernetesVersion: payload.KubernetesVersion,
		CredentialID:      payload.CredentialID,
	}

	handler.cloudClusterSetupService.Request(&request)
	return response.JSON(w, endpoint)
}

func (handler *Handler) createEndpoint(name string, provider portaineree.CloudProvider, metadata EnvironmentMetadata) (*portaineree.Endpoint, error) {
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

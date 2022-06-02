package providers

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
)

type DefaultProvisionPayload struct {
	Name              string                   `validate:"required" example:"myDevCluster"`
	NodeSize          string                   `validate:"required" example:"g3.small"`
	NodeCount         int                      `validate:"required" example:"3"`
	Region            string                   `validate:"required" example:"NYC1"`
	NetworkID         string                   `example:"8465fb26-632e-4fa3-bb9b-21c449629026"`
	KubernetesVersion string                   `validate:"required" example:"1.23"`
	CredentialID      models.CloudCredentialID `validate:"required" example:"1"`

	Meta types.EnvironmentMetadata `json:"meta"`
}

func (payload *DefaultProvisionPayload) Validate(r *http.Request) error {
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
	if govalidator.IsNonPositive(float64(payload.CredentialID)) {
		return errors.New("Invalid Credentials")
	}
	return nil
}

func (payload *DefaultProvisionPayload) GetCloudProvider(provider string) (*portaineree.CloudProvider, error) {
	cloudProvider, ok := types.CloudProvidersMap[types.CloudProviderShortName(provider)]
	if !ok {
		return nil, errors.New("Invalid cloud provider")
	}

	cloudProvider.Region = payload.Region
	cloudProvider.Size = &payload.NodeSize
	cloudProvider.NodeCount = payload.NodeCount
	cloudProvider.NetworkID = &payload.NetworkID
	cloudProvider.CredentialID = payload.CredentialID

	return &cloudProvider, nil
}

func (payload *DefaultProvisionPayload) GetEndpointName() string {
	return payload.Name
}

func (payload *DefaultProvisionPayload) GetEnvironmentMetadata() types.EnvironmentMetadata {
	return payload.Meta
}

func (payload *DefaultProvisionPayload) GetCloudProvisioningRequest(endpointID portaineree.EndpointID, provider string) *portaineree.CloudProvisioningRequest {
	return &portaineree.CloudProvisioningRequest{
		EndpointID:        endpointID,
		Provider:          provider,
		Region:            payload.Region,
		Name:              payload.Name,
		NodeSize:          payload.NodeSize,
		NetworkID:         payload.NetworkID,
		NodeCount:         payload.NodeCount,
		KubernetesVersion: payload.KubernetesVersion,
		CredentialID:      payload.CredentialID,
	}
}

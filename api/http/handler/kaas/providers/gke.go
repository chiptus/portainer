package providers

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/sirupsen/logrus"
)

type (
	GKEProvisionPayload struct {
		DefaultProvisionPayload

		CPU int     `example:"2"`
		RAM float64 `example:"4"`
		HDD int     `example:"100"`
	}
)

func (payload *GKEProvisionPayload) Validate(r *http.Request) error {

	if err := payload.DefaultProvisionPayload.Validate(r); err != nil {
		return err
	}

	return nil
}

func (payload *GKEProvisionPayload) GetCloudProvider(_ string) (*portaineree.CloudProvider, error) {
	cloudProvider, ok := types.CloudProvidersMap[types.CloudProviderShortName(portaineree.CloudProviderGKE)]
	if !ok {
		return nil, errors.New("Invalid cloud provider")
	}

	logrus.Infof("Cloud provider: %s", cloudProvider.Name)

	cloudProvider.Region = payload.Region
	cloudProvider.Size = &payload.NodeSize
	cloudProvider.NodeCount = payload.NodeCount
	cloudProvider.NetworkID = &payload.NetworkID
	cloudProvider.CredentialID = payload.CredentialID

	cloudProvider.CPU = &payload.CPU
	cloudProvider.RAM = &payload.RAM
	cloudProvider.HDD = &payload.HDD

	cloudProvider.KubernetesVersion = payload.KubernetesVersion

	return &cloudProvider, nil
}

func (payload *GKEProvisionPayload) GetEndpointName() string {
	return payload.Name
}

func (payload *GKEProvisionPayload) GetEnvironmentMetadata() types.EnvironmentMetadata {
	return payload.Meta
}

func (payload *GKEProvisionPayload) GetCloudProvisioningRequest(endpointID portaineree.EndpointID, _ string) *portaineree.CloudProvisioningRequest {
	return &portaineree.CloudProvisioningRequest{
		EndpointID:        endpointID,
		Provider:          portaineree.CloudProviderGKE,
		Region:            payload.Region,
		Name:              payload.Name,
		NodeSize:          payload.NodeSize,
		NetworkID:         payload.NetworkID,
		NodeCount:         payload.NodeCount,
		KubernetesVersion: payload.KubernetesVersion,
		CredentialID:      payload.CredentialID,
	}
}

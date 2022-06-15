package providers

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
)

// this default is the same as what eksctl uses.  But the amazon console chooses 20GB by default.
const defaultNodeVolumeSize = 80

type AmazonProvisionPayload struct {
	DefaultProvisionPayload

	AmiType        string `validate:"required" example:"BOTTLEROCKET_x86_64"`
	InstanceType   string `validate:"required" example:"m5.large"`
	NodeVolumeSize *int   `example:"20"`
}

func (payload *AmazonProvisionPayload) Validate(r *http.Request) error {

	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid cluster name")
	}
	if payload.NodeCount < 1 {
		return errors.New("Invalid node count")
	}
	if govalidator.IsNull(payload.Region) {
		return errors.New("Invalid region")
	}
	if govalidator.IsNonPositive(float64(payload.CredentialID)) {
		return errors.New("Invalid credentials")
	}
	if govalidator.IsNull(payload.AmiType) {
		return errors.New("Invalid AMI type")
	}
	if govalidator.IsNull(payload.InstanceType) {
		return errors.New("Invalid instance type")
	}
	if payload.NodeVolumeSize != nil && !govalidator.InRange(float64(*payload.NodeVolumeSize), 1, 16384) {
		return errors.New("Node volume size out of range")
	}

	return nil
}

func (payload *AmazonProvisionPayload) GetCloudProvider(_ string) (*portaineree.CloudProvider, error) {
	cloudProvider, ok := types.CloudProvidersMap[types.CloudProviderShortName(portaineree.CloudProviderAmazon)]
	if !ok {
		return nil, errors.New("Invalid cloud provider")
	}

	cloudProvider.Region = payload.Region
	cloudProvider.NodeCount = payload.NodeCount
	cloudProvider.CredentialID = payload.CredentialID
	cloudProvider.AmiType = &payload.AmiType
	cloudProvider.InstanceType = &payload.InstanceType
	cloudProvider.NodeVolumeSize = payload.NodeVolumeSize
	return &cloudProvider, nil
}

func (payload *AmazonProvisionPayload) GetEndpointName() string {
	return payload.Name
}

func (payload *AmazonProvisionPayload) GetCloudProvisioningRequest(endpointID portaineree.EndpointID, _ string) *portaineree.CloudProvisioningRequest {
	request := &portaineree.CloudProvisioningRequest{
		EndpointID:        endpointID,
		Provider:          portaineree.CloudProviderAmazon,
		Region:            payload.Region,
		Name:              payload.Name,
		NodeCount:         payload.NodeCount,
		KubernetesVersion: payload.KubernetesVersion,
		CredentialID:      payload.CredentialID,
		AmiType:           payload.AmiType,
		InstanceType:      payload.InstanceType,
	}

	if payload.NodeVolumeSize == nil {
		request.NodeVolumeSize = defaultNodeVolumeSize
	} else {
		request.NodeVolumeSize = *payload.NodeVolumeSize
	}

	return request
}

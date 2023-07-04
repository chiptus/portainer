package providers

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/rs/zerolog/log"
)

type (
	AzureProvisionPayload struct {
		DefaultProvisionPayload

		// Azure specific fields
		ResourceGroup     string   `json:"resourceGroup"`
		ResourceGroupName string   `json:"resourceGroupName"`
		Tier              string   `json:"tier"`
		PoolName          string   `json:"poolName"`
		DNSPrefix         string   `json:"dnsPrefix"`
		AvailabilityZones []string `json:"availabilityZones"`
	}
)

func (payload *AzureProvisionPayload) Validate(r *http.Request) error {

	if err := payload.DefaultProvisionPayload.Validate(r); err != nil {
		return err
	}

	if govalidator.IsNull(payload.Tier) {
		return errors.New("invalid resource tier")
	}
	if govalidator.IsNull(payload.PoolName) {
		return errors.New("invalid pool name")
	}
	if govalidator.IsNull(payload.DNSPrefix) {
		return errors.New("invalid DNS prefix")
	}
	if govalidator.IsNull(payload.ResourceGroupName) && govalidator.IsNull(payload.ResourceGroup) {
		return errors.New("either choose a resource group or a resource group name")
	}

	return nil
}

func (payload *AzureProvisionPayload) GetCloudProvider(_ string) (*portaineree.CloudProvider, error) {
	cloudProvider, ok := types.CloudProvidersMap[types.CloudProviderShortName(portaineree.CloudProviderAzure)]
	if !ok {
		return nil, errors.New("invalid cloud provider")
	}

	log.Info().Str("provider", cloudProvider.Name).Msg("cloud provider")

	cloudProvider.Region = payload.Region
	cloudProvider.Size = &payload.NodeSize
	cloudProvider.NodeCount = payload.NodeCount
	cloudProvider.NetworkID = &payload.NetworkID
	cloudProvider.CredentialID = payload.CredentialID

	// Azure specific fields
	cloudProvider.ResourceGroup = payload.ResourceGroup
	if payload.ResourceGroupName != "" {
		cloudProvider.ResourceGroup = payload.ResourceGroupName
	}
	cloudProvider.Tier = payload.Tier
	cloudProvider.PoolName = payload.PoolName
	cloudProvider.DNSPrefix = payload.DNSPrefix
	cloudProvider.KubernetesVersion = payload.KubernetesVersion

	return &cloudProvider, nil
}

func (payload *AzureProvisionPayload) GetEndpointName() string {
	return payload.Name
}

func (payload *AzureProvisionPayload) GetEnvironmentMetadata() types.EnvironmentMetadata {
	return payload.Meta
}

func (payload *AzureProvisionPayload) GetCloudProvisioningRequest(endpointID portaineree.EndpointID, _ string) *portaineree.CloudProvisioningRequest {
	return &portaineree.CloudProvisioningRequest{
		EndpointID:        endpointID,
		Provider:          portaineree.CloudProviderAzure,
		Region:            payload.Region,
		Name:              payload.Name,
		NodeSize:          payload.NodeSize,
		NetworkID:         payload.NetworkID,
		NodeCount:         payload.NodeCount,
		KubernetesVersion: payload.KubernetesVersion,
		CredentialID:      payload.CredentialID,
		AvailabilityZones: payload.AvailabilityZones,

		ResourceGroup:     payload.ResourceGroup,
		ResourceGroupName: payload.ResourceGroupName,
		Tier:              payload.Tier,
		PoolName:          payload.PoolName,
		DNSPrefix:         payload.DNSPrefix,

		CustomTemplateID:      payload.Meta.CustomTemplateID,
		CustomTemplateContent: payload.Meta.CustomTemplateContent,
	}
}

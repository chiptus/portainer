package providers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
)

type Microk8sProvisionPayload struct {
	NodeIPs          []string                     `json:"nodeIPs"`
	Addons           []string                     `json:"addons"`
	CustomTemplateID portaineree.CustomTemplateID `json:"customTemplateID"`

	DefaultProvisionPayload
}

func (payload *Microk8sProvisionPayload) Validate(r *http.Request) error {

	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid cluster name")
	}
	// TODO: REVIEW-POC-MICROK8S
	// Only supported node count for the scope of the POC
	if payload.NodeCount != 1 && payload.NodeCount != 3 {
		return errors.New("Invalid node count")
	}
	if len(payload.NodeIPs) != payload.NodeCount {
		return errors.New("Invalid count of node IPs")
	}
	if govalidator.IsNonPositive(float64(payload.CredentialID)) {
		return errors.New("Invalid credentials")
	}

	return nil
}

func (payload *Microk8sProvisionPayload) GetCloudProvider(_ string) (*portaineree.CloudProvider, error) {
	cloudProvider, ok := types.CloudProvidersMap[types.CloudProviderShortName(portaineree.CloudProviderMicrok8s)]
	if !ok {
		return nil, errors.New("Invalid cloud provider")
	}

	cloudProvider.CredentialID = payload.CredentialID
	cloudProvider.NodeCount = payload.NodeCount
	if payload.Addons != nil {
		addons := strings.Join(payload.Addons, ", ")
		cloudProvider.Addons = &addons
	}
	if payload.NodeIPs != nil {
		addons := strings.Join(payload.NodeIPs, ", ")
		cloudProvider.NodeIPs = &addons
	}
	return &cloudProvider, nil
}

func (payload *Microk8sProvisionPayload) GetCloudProvisioningRequest(endpointID portaineree.EndpointID, _ string) *portaineree.CloudProvisioningRequest {
	request := &portaineree.CloudProvisioningRequest{
		EndpointID:       endpointID,
		Provider:         portaineree.CloudProviderMicrok8s,
		Name:             payload.Name,
		CredentialID:     payload.CredentialID,
		NodeCount:        payload.NodeCount,
		NodeIPs:          payload.NodeIPs,
		Addons:           payload.Addons,
		CustomTemplateID: payload.CustomTemplateID,
	}

	return request
}

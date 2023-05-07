package providers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/portainer/portainer-ee/api/internal/iprange"
)

const maxNodeLimit = 100

type Microk8sTestSSHPayload struct {
	NodeIPs      []string                 `validate:"required" json:"nodeIPs"`
	CredentialID models.CloudCredentialID `validate:"required" example:"1"`
}

type Microk8sProvisionPayload struct {
	NodeIPs           []string `validate:"required" json:"nodeIPs"`
	KubernetesVersion string   `validate:"required" json:"kubernetesVersion"`
	Addons            []string `json:"addons"`

	DefaultProvisionPayload
}

func (payload *Microk8sTestSSHPayload) Validate(r *http.Request) error {
	ips, err := validateNodeIPs(payload.NodeIPs)
	payload.NodeIPs = ips
	return err
}

func validateNodeIPs(ipRanges []string) (nodeIPs []string, err error) {
	if ipRanges != nil && len(ipRanges) == 0 {
		return nil, errors.New("invalid count of node IPs")
	}

	for _, entry := range ipRanges {
		ipr, err := iprange.Parse(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid ip or ip range '%s' error: %w", entry, err)
		}

		if ipr.Validate() != nil {
			return nil, err
		}

		ips := ipr.Expand()

		// Return an error if there are duplicates
		for _, ip := range ips {
			if strings.Contains(strings.Join(nodeIPs, ","), ip) {
				return nil, fmt.Errorf("duplicate ip '%s' found", ip)
			}
		}

		nodeIPs = append(nodeIPs, ips...)
	}

	if len(nodeIPs) > maxNodeLimit {
		return nil, fmt.Errorf("too many nodes. Max %d nodes allowed", maxNodeLimit)
	}

	return nodeIPs, nil
}

func (payload *Microk8sProvisionPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid cluster name")
	}

	if payload.NodeIPs != nil && len(payload.NodeIPs) == 0 {
		return errors.New("Invalid count of node IPs")
	}

	nodeips, err := validateNodeIPs(payload.NodeIPs)
	if err != nil {
		return err
	}

	payload.NodeIPs = nodeips
	payload.NodeCount = len(nodeips)

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
		EndpointID:            endpointID,
		Provider:              portaineree.CloudProviderMicrok8s,
		Name:                  payload.Name,
		CredentialID:          payload.CredentialID,
		NodeCount:             payload.NodeCount,
		NodeIPs:               payload.NodeIPs,
		Addons:                payload.Addons,
		CustomTemplateID:      payload.Meta.CustomTemplateID,
		CustomTemplateContent: payload.Meta.CustomTemplateContent,
		KubernetesVersion:     payload.KubernetesVersion,
	}

	return request
}

package providers

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	portainer "github.com/portainer/portainer/api"
)

type Providers interface {
	GetCloudProvider(string) (*portaineree.CloudProvider, error)
	Validate(r *http.Request) error
	GetEndpointName() string
	GetCloudProvisioningRequest(portainer.EndpointID, string) *portaineree.CloudProvisioningRequest
	GetEnvironmentMetadata() types.EnvironmentMetadata
}

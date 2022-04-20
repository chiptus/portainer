package clientFactory

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/nomad/clientFactory/client"
)

func (factory *ClientFactory) createClient(endpoint *portaineree.Endpoint) (portaineree.NomadClient, error) {
	return client.NewClient(endpoint, factory.reverseTunnelService, factory.signatureService)
}

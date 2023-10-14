package clientFactory

import (
	cmap "github.com/orcaman/concurrent-map"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type (
	// ClientFactory is used to create and cache Nomad clients
	ClientFactory struct {
		dataStore            dataservices.DataStore
		reverseTunnelService portaineree.ReverseTunnelService
		signatureService     portainer.DigitalSignatureService
		instanceID           string
		clientsMap           cmap.ConcurrentMap
	}
)

// NewClientFactory returns a new instance of a ClientFactory
func NewClientFactory(signatureService portainer.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID string) *ClientFactory {
	return &ClientFactory{
		dataStore:            dataStore,
		signatureService:     signatureService,
		reverseTunnelService: reverseTunnelService,
		instanceID:           instanceID,
		clientsMap:           cmap.New(),
	}
}

func (factory *ClientFactory) GetInstanceID() (instanceID string) {
	return factory.instanceID
}

package clientFactory

import (
	cmap "github.com/orcaman/concurrent-map"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type (
	// ClientFactory is used to create and cache Nomad clients
	ClientFactory struct {
		dataStore            dataservices.DataStore
		reverseTunnelService portaineree.ReverseTunnelService
		signatureService     portaineree.DigitalSignatureService
		instanceID           string
		clientsMap           cmap.ConcurrentMap
	}
)

// NewClientFactory returns a new instance of a ClientFactory
func NewClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID string) *ClientFactory {
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

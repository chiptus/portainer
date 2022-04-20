package clientFactory

import (
	"errors"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
)

// GetClient checks if an existing client is already registered for the environment(endpoint) and returns it if one is found.
// If no client is registered, it will create a new client, register it, and returns it.
func (factory *ClientFactory) GetClient(endpoint *portaineree.Endpoint) (portaineree.NomadClient, error) {
	key := strconv.Itoa(int(endpoint.ID))
	nomadClientCache, ok := factory.clientsMap.Get(key)

	if ok {
		nomadClient, ok := nomadClientCache.(portaineree.NomadClient)
		if !ok {
			return nil, errors.New("invalid nomad client cache")
		}

		// if the tunnel of the client is closed, remove it
		ok = nomadClient.Validate()
		if !ok {
			factory.RemoveClient(endpoint.ID)
		}

		return nomadClient, nil
	}

	nomadClient, err := factory.createClient(endpoint)
	if err != nil {
		return nil, err
	}

	factory.clientsMap.Set(key, nomadClient)

	return nomadClient, nil
}

// RemoveClient Remove the cached Nomad client so a new one can be created
func (factory *ClientFactory) RemoveClient(endpointID portaineree.EndpointID) {
	factory.clientsMap.Remove(strconv.Itoa(int(endpointID)))
}

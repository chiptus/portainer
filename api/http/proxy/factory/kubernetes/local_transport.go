package kubernetes

import (
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

type localTransport struct {
	*baseTransport
}

// NewLocalTransport returns a new transport that can be used to send requests to the local Kubernetes API
func NewLocalTransport(tokenManager *tokenManager, endpoint *portainer.Endpoint, userActivityStore portainer.UserActivityStore, k8sClientFactory *cli.ClientFactory, dataStore portainer.DataStore) (*localTransport, error) {
	config, err := crypto.CreateTLSConfigurationFromBytes(nil, nil, nil, true, true)
	if err != nil {
		return nil, err
	}

	transport := &localTransport{
		baseTransport: newBaseTransport(
			&http.Transport{
				TLSClientConfig: config,
			},
			tokenManager,
			endpoint,
			userActivityStore,
			k8sClientFactory,
			dataStore,
		),
	}

	return transport, nil
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *localTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	err := transport.prepareRoundTrip(request)
	if err != nil {
		return nil, err
	}

	return transport.baseTransport.RoundTrip(request)
}

package kubernetes

import (
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

type edgeTransport struct {
	*baseTransport
	signatureService     portainer.DigitalSignatureService
	reverseTunnelService portainer.ReverseTunnelService
}

// NewAgentTransport returns a new transport that can be used to send signed requests to a Portainer Edge agent
func NewEdgeTransport(dataStore portainer.DataStore, signatureService portainer.DigitalSignatureService, reverseTunnelService portainer.ReverseTunnelService, endpoint *portainer.Endpoint, tokenManager *tokenManager, userActivityStore portainer.UserActivityStore, k8sClientFactory *cli.ClientFactory) *edgeTransport {
	transport := &edgeTransport{
		reverseTunnelService: reverseTunnelService,
		signatureService:     signatureService,
		baseTransport: newBaseTransport(
			&http.Transport{},
			tokenManager,
			endpoint,
			userActivityStore,
			k8sClientFactory,
			dataStore,
		),
	}

	return transport
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *edgeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	token, err := getRoundTripToken(request, transport.tokenManager, transport.endpoint.ID)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portainer.PortainerAgentKubernetesSATokenHeader, token)

	if strings.HasPrefix(request.URL.Path, "/v2") {
		decorateAgentRequest(request, transport.dataStore)
	}

	signature, err := transport.signatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portainer.PortainerAgentPublicKeyHeader, transport.signatureService.EncodedPublicKey())
	request.Header.Set(portainer.PortainerAgentSignatureHeader, signature)

	response, err := transport.baseTransport.RoundTrip(request)

	if err == nil {
		transport.reverseTunnelService.SetTunnelStatusToActive(transport.endpoint.ID)
	} else {
		transport.reverseTunnelService.SetTunnelStatusToIdle(transport.endpoint.ID)
	}

	return response, err
}

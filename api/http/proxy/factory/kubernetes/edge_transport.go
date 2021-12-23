package kubernetes

import (
	"net/http"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

type edgeTransport struct {
	*baseTransport
	signatureService     portaineree.DigitalSignatureService
	reverseTunnelService portaineree.ReverseTunnelService
}

// NewAgentTransport returns a new transport that can be used to send signed requests to a Portainer Edge agent
func NewEdgeTransport(dataStore portaineree.DataStore, signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, endpoint *portaineree.Endpoint, tokenManager *tokenManager, userActivityService portaineree.UserActivityService, k8sClientFactory *cli.ClientFactory) *edgeTransport {
	transport := &edgeTransport{
		reverseTunnelService: reverseTunnelService,
		signatureService:     signatureService,
		baseTransport: newBaseTransport(
			&http.Transport{},
			tokenManager,
			endpoint,
			userActivityService,
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

	request.Header.Set(portaineree.PortainerAgentKubernetesSATokenHeader, token)

	if strings.HasPrefix(request.URL.Path, "/v2") {
		decorateAgentRequest(request, transport.dataStore)
	}

	signature, err := transport.signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portaineree.PortainerAgentPublicKeyHeader, transport.signatureService.EncodedPublicKey())
	request.Header.Set(portaineree.PortainerAgentSignatureHeader, signature)

	response, err := transport.baseTransport.RoundTrip(request)

	if err == nil {
		transport.reverseTunnelService.SetTunnelStatusToActive(transport.endpoint.ID)
	} else {
		transport.reverseTunnelService.SetTunnelStatusToIdle(transport.endpoint.ID)
	}

	return response, err
}

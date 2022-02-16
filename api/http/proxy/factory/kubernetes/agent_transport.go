package kubernetes

import (
	"crypto/tls"
	"net/http"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

type agentTransport struct {
	*baseTransport
	signatureService portaineree.DigitalSignatureService
}

// NewAgentTransport returns a new transport that can be used to send signed requests to a Portainer agent
func NewAgentTransport(dataStore dataservices.DataStore, signatureService portaineree.DigitalSignatureService, tlsConfig *tls.Config, tokenManager *tokenManager, endpoint *portaineree.Endpoint, userActivityService portaineree.UserActivityService, k8sClientFactory *cli.ClientFactory) *agentTransport {
	transport := &agentTransport{
		signatureService: signatureService,
		baseTransport: newBaseTransport(
			&http.Transport{
				TLSClientConfig: tlsConfig,
			},
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
func (transport *agentTransport) RoundTrip(request *http.Request) (*http.Response, error) {
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

	return transport.baseTransport.RoundTrip(request)
}

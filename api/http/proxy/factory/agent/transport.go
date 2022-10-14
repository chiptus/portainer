package agent

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
)

// Transport is an http.Transport wrapper that adds custom http headers to communicate to an Agent
type Transport struct {
	httpTransport    *http.Transport
	signatureService portaineree.DigitalSignatureService
}

// NewTransport returns a new transport that can be used to send signed requests to a Portainer agent
func NewTransport(signatureService portaineree.DigitalSignatureService, httpTransport *http.Transport) *Transport {
	transport := &Transport{
		httpTransport:    httpTransport,
		signatureService: signatureService,
	}

	return transport
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	signature, err := transport.signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portaineree.PortainerAgentPublicKeyHeader, transport.signatureService.EncodedPublicKey())
	request.Header.Set(portaineree.PortainerAgentSignatureHeader, signature)

	return transport.httpTransport.RoundTrip(request)
}

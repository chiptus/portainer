package client

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
)

type NomadClientTransport struct {
	signatureHeader string
	publicKeyHeader string
	tunnelAddress   string
}

func (transport *NomadClientTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Add(portaineree.PortainerAgentPublicKeyHeader, transport.publicKeyHeader)
	request.Header.Add(portaineree.PortainerAgentSignatureHeader, transport.signatureHeader)

	request.URL.Path = "/nomad" + request.URL.Path
	request.URL.Host = transport.tunnelAddress
	return http.DefaultTransport.RoundTrip(request)
}

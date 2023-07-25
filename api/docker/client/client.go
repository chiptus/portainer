package client

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/client"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/crypto"
)

var errUnsupportedEnvironmentType = errors.New("Environment not supported")

const defaultDockerRequestTimeout = 60 * time.Second

// ClientFactory is used to create Docker clients
type ClientFactory struct {
	signatureService     portaineree.DigitalSignatureService
	reverseTunnelService portaineree.ReverseTunnelService
}

// NewClientFactory returns a new instance of a ClientFactory
func NewClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService) *ClientFactory {
	return &ClientFactory{
		signatureService:     signatureService,
		reverseTunnelService: reverseTunnelService,
	}
}

// CreateClient is a generic function to create a Docker client based on
// a specific environment(endpoint) configuration. The nodeName parameter can be used
// with an agent enabled environment(endpoint) to target a specific node in an agent cluster.
func (factory *ClientFactory) CreateClient(endpoint *portaineree.Endpoint, nodeName string, timeout *time.Duration) (*client.Client, error) {
	switch endpoint.Type {
	case portaineree.AzureEnvironment:
		return nil, errUnsupportedEnvironmentType
	case portaineree.AgentOnDockerEnvironment:
		return createAgentClient(endpoint, factory.signatureService, nodeName, timeout)
	case portaineree.EdgeAgentOnDockerEnvironment:
		return createEdgeClient(endpoint, factory.signatureService, factory.reverseTunnelService, nodeName, timeout)
	}

	if strings.HasPrefix(endpoint.URL, "unix://") || strings.HasPrefix(endpoint.URL, "npipe://") {
		return createLocalClient(endpoint)
	}

	return createTCPClient(endpoint, timeout)
}

func createLocalClient(endpoint *portaineree.Endpoint) (*client.Client, error) {
	return client.NewClientWithOpts(
		client.WithHost(endpoint.URL),
		client.WithAPIVersionNegotiation(),
	)
}

func createTCPClient(endpoint *portaineree.Endpoint, timeout *time.Duration) (*client.Client, error) {
	httpCli, err := httpClient(endpoint, timeout)
	if err != nil {
		return nil, err
	}

	return client.NewClientWithOpts(
		client.WithHost(endpoint.URL),
		client.WithAPIVersionNegotiation(),
		client.WithHTTPClient(httpCli),
	)
}

func createEdgeClient(endpoint *portaineree.Endpoint, signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, nodeName string, timeout *time.Duration) (*client.Client, error) {
	httpCli, err := httpClient(endpoint, timeout)
	if err != nil {
		return nil, err
	}

	signature, err := signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		portaineree.PortainerAgentPublicKeyHeader: signatureService.EncodedPublicKey(),
		portaineree.PortainerAgentSignatureHeader: signature,
	}

	if nodeName != "" {
		headers[portaineree.PortainerAgentTargetHeader] = nodeName
	}

	tunnel, err := reverseTunnelService.GetActiveTunnel(endpoint)
	if err != nil {
		return nil, err
	}

	endpointURL := fmt.Sprintf("http://127.0.0.1:%d", tunnel.Port)

	return client.NewClientWithOpts(
		client.WithHost(endpointURL),
		client.WithAPIVersionNegotiation(),
		client.WithHTTPClient(httpCli),
		client.WithHTTPHeaders(headers),
	)
}

func createAgentClient(endpoint *portaineree.Endpoint, signatureService portaineree.DigitalSignatureService, nodeName string, timeout *time.Duration) (*client.Client, error) {
	httpCli, err := httpClient(endpoint, timeout)
	if err != nil {
		return nil, err
	}

	signature, err := signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		portaineree.PortainerAgentPublicKeyHeader: signatureService.EncodedPublicKey(),
		portaineree.PortainerAgentSignatureHeader: signature,
	}

	if nodeName != "" {
		headers[portaineree.PortainerAgentTargetHeader] = nodeName
	}

	return client.NewClientWithOpts(
		client.WithHost(endpoint.URL),
		client.WithAPIVersionNegotiation(),
		client.WithHTTPClient(httpCli),
		client.WithHTTPHeaders(headers),
	)
}

func httpClient(endpoint *portaineree.Endpoint, timeout *time.Duration) (*http.Client, error) {
	transport := &http.Transport{}

	if endpoint.TLSConfig.TLS {
		tlsConfig, err := crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsConfig
	}

	clientTimeout := defaultDockerRequestTimeout
	if timeout != nil {
		clientTimeout = *timeout
	}

	return &http.Client{
		Transport: transport,
		Timeout:   clientTimeout,
	}, nil
}

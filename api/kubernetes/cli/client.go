package cli

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type (
	// ClientFactory is used to create Kubernetes clients
	ClientFactory struct {
		dataStore            dataservices.DataStore
		reverseTunnelService portaineree.ReverseTunnelService
		signatureService     portaineree.DigitalSignatureService
		instanceID           string
		endpointClients      cmap.ConcurrentMap
		endpointProxyClients *cache.Cache
		AddrHTTPS            string
	}

	// KubeClient represent a service used to execute Kubernetes operations
	KubeClient struct {
		cli        kubernetes.Interface
		instanceID string
		lock       *sync.Mutex
	}
)

// NewClientFactory returns a new instance of a ClientFactory
func NewClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID, addrHTTPS, userSessionTimeout string) (*ClientFactory, error) {
	if userSessionTimeout == "" {
		userSessionTimeout = portaineree.DefaultUserSessionTimeout
	}
	timeout, err := time.ParseDuration(userSessionTimeout)
	if err != nil {
		return nil, err
	}

	return &ClientFactory{
		dataStore:            dataStore,
		signatureService:     signatureService,
		reverseTunnelService: reverseTunnelService,
		instanceID:           instanceID,
		endpointClients:      cmap.New(),
		endpointProxyClients: cache.New(timeout, timeout),
		AddrHTTPS:            addrHTTPS,
	}, nil
}

func (factory *ClientFactory) GetInstanceID() (instanceID string) {
	return factory.instanceID
}

// Remove the cached kube client so a new one can be created
func (factory *ClientFactory) RemoveKubeClient(endpointID portaineree.EndpointID) {
	factory.endpointClients.Remove(strconv.Itoa(int(endpointID)))
}

// GetKubeClient checks if an existing client is already registered for the environment(endpoint) and returns it if one is found.
// If no client is registered, it will create a new client, register it, and returns it.
func (factory *ClientFactory) GetKubeClient(endpoint *portaineree.Endpoint) (portaineree.KubeClient, error) {
	key := strconv.Itoa(int(endpoint.ID))
	client, ok := factory.endpointClients.Get(key)
	if !ok {
		client, err := factory.createCachedAdminKubeClient(endpoint)
		if err != nil {
			return nil, err
		}

		factory.endpointClients.Set(key, client)
		return client, nil
	}

	return client.(portaineree.KubeClient), nil
}

// GetProxyKubeClient retrieves a KubeClient from the cache. You should be
// calling SetProxyKubeClient before first. It is normally, called the
// kubernetes middleware.
func (factory *ClientFactory) GetProxyKubeClient(endpointID, token string) (portaineree.KubeClient, bool) {
	client, ok := factory.endpointProxyClients.Get(endpointID + "." + token)
	if !ok {
		return nil, false
	}
	return client.(portaineree.KubeClient), true
}

// SetProxyKubeClient stores a kubeclient in the cache.
func (factory *ClientFactory) SetProxyKubeClient(endpointID, token string, cli portaineree.KubeClient) {
	factory.endpointProxyClients.Set(endpointID+"."+token, cli, 0)
}

// CreateKubeClientFromKubeConfig creates a KubeClient from a clusterID, and
// Kubernetes config.
func (factory *ClientFactory) CreateKubeClientFromKubeConfig(clusterID string, kubeConfig []byte) (portaineree.KubeClient, error) {
	config, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}
	cliConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	cli, err := kubernetes.NewForConfig(cliConfig)
	if err != nil {
		return nil, err
	}

	kubecli := &KubeClient{
		cli:        cli,
		instanceID: factory.instanceID,
		lock:       &sync.Mutex{},
	}

	return kubecli, nil
}

func (factory *ClientFactory) createCachedAdminKubeClient(endpoint *portaineree.Endpoint) (portaineree.KubeClient, error) {
	cli, err := factory.CreateClient(endpoint)
	if err != nil {
		return nil, err
	}

	kubecli := &KubeClient{
		cli:        cli,
		instanceID: factory.instanceID,
		lock:       &sync.Mutex{},
	}

	return kubecli, nil
}

// CreateClient returns a pointer to a new Clientset instance
func (factory *ClientFactory) CreateClient(endpoint *portaineree.Endpoint) (*kubernetes.Clientset, error) {
	switch endpoint.Type {
	case portaineree.KubernetesLocalEnvironment:
		return buildLocalClient()
	case portaineree.AgentOnKubernetesEnvironment:
		return factory.buildAgentClient(endpoint)
	case portaineree.EdgeAgentOnKubernetesEnvironment:
		return factory.buildEdgeClient(endpoint)
	}

	return nil, errors.New("unsupported environment type")
}

type agentHeaderRoundTripper struct {
	signatureHeader string
	publicKeyHeader string

	roundTripper http.RoundTripper
}

// RoundTrip is the implementation of the http.RoundTripper interface.
// It decorates the request with specific agent headers
func (rt *agentHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(portaineree.PortainerAgentPublicKeyHeader, rt.publicKeyHeader)
	req.Header.Add(portaineree.PortainerAgentSignatureHeader, rt.signatureHeader)

	return rt.roundTripper.RoundTrip(req)
}

func (factory *ClientFactory) buildAgentClient(endpoint *portaineree.Endpoint) (*kubernetes.Clientset, error) {
	endpointURL := fmt.Sprintf("https://%s/kubernetes", endpoint.URL)

	return factory.createRemoteClient(endpointURL)
}

func (factory *ClientFactory) buildEdgeClient(endpoint *portaineree.Endpoint) (*kubernetes.Clientset, error) {
	tunnel, err := factory.reverseTunnelService.GetActiveTunnel(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed activating tunnel")
	}
	endpointURL := fmt.Sprintf("http://127.0.0.1:%d/kubernetes", tunnel.Port)

	return factory.createRemoteClient(endpointURL)
}

func (factory *ClientFactory) createRemoteClient(endpointURL string) (*kubernetes.Clientset, error) {
	signature, err := factory.signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags(endpointURL, "")
	if err != nil {
		return nil, err
	}
	config.Insecure = true

	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return &agentHeaderRoundTripper{
			signatureHeader: signature,
			publicKeyHeader: factory.signatureService.EncodedPublicKey(),
			roundTripper:    rt,
		}
	})

	return kubernetes.NewForConfig(config)
}

func buildLocalClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

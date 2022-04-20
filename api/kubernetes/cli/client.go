package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	cmap "github.com/orcaman/concurrent-map"

	"github.com/pkg/errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer/api/filesystem"
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
	}

	// KubeClient represent a service used to execute Kubernetes operations
	KubeClient struct {
		cli        kubernetes.Interface
		instanceID string
		lock       *sync.Mutex
	}
)

// NewClientFactory returns a new instance of a ClientFactory
func NewClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID string) *ClientFactory {
	return &ClientFactory{
		dataStore:            dataStore,
		signatureService:     signatureService,
		reverseTunnelService: reverseTunnelService,
		instanceID:           instanceID,
		endpointClients:      cmap.New(),
	}
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
		client, err := factory.createKubeClient(endpoint)
		if err != nil {
			return nil, err
		}

		factory.endpointClients.Set(key, client)
		return client, nil
	}

	return client.(portaineree.KubeClient), nil
}

// CreateKubeClientFromKubeConfig creates a KubeClient from a clusterID, and
// Kubernetes config.
func (factory *ClientFactory) CreateKubeClientFromKubeConfig(clusterID string, kubeConfig string) (portaineree.KubeClient, error) {
	// We must store the kubeConfig in a temp file because
	// clientcmd.BuildConfigFromFlags takes a filepath an input.
	kubeConfigPath := filepath.Join(os.TempDir(), clusterID)
	err := filesystem.WriteToFile(kubeConfigPath, []byte(kubeConfig))
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	cli, err := kubernetes.NewForConfig(config)
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

func (factory *ClientFactory) createKubeClient(endpoint *portaineree.Endpoint) (portaineree.KubeClient, error) {
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

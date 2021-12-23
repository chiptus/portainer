package proxy

import (
	"fmt"
	"net/http"

	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/proxy/factory"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

// TODO: contain code related to legacy extension management

type (
	// Manager represents a service used to manage proxies to environments(endpoints) and extensions.
	Manager struct {
		proxyFactory           *factory.ProxyFactory
		endpointProxies        cmap.ConcurrentMap
		legacyExtensionProxies cmap.ConcurrentMap
		k8sClientFactory       *cli.ClientFactory
	}
)

// NewManager initializes a new proxy Service
func NewManager(
	dataStore portaineree.DataStore,
	signatureService portaineree.DigitalSignatureService,
	tunnelService portaineree.ReverseTunnelService,
	clientFactory *docker.ClientFactory,
	kubernetesClientFactory *cli.ClientFactory,
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager,
	authService *authorization.Service,
	userActivityService portaineree.UserActivityService,
) *Manager {
	return &Manager{
		endpointProxies:        cmap.New(),
		legacyExtensionProxies: cmap.New(),
		k8sClientFactory:       kubernetesClientFactory,
		proxyFactory: factory.NewProxyFactory(
			dataStore,
			signatureService,
			tunnelService,
			clientFactory,
			kubernetesClientFactory,
			kubernetesTokenCacheManager,
			authService,
			userActivityService,
		),
	}
}

// CreateAndRegisterEndpointProxy creates a new HTTP reverse proxy based on environment(endpoint) properties and and adds it to the registered proxies.
// It can also be used to create a new HTTP reverse proxy and replace an already registered proxy.
func (manager *Manager) CreateAndRegisterEndpointProxy(endpoint *portaineree.Endpoint) (http.Handler, error) {
	proxy, err := manager.proxyFactory.NewEndpointProxy(endpoint)
	if err != nil {
		return nil, err
	}

	manager.endpointProxies.Set(fmt.Sprint(endpoint.ID), proxy)
	return proxy, nil
}

// CreateAgentProxyServer creates a new HTTP reverse proxy based on environment(endpoint) properties and and adds it to the registered proxies.
// It can also be used to create a new HTTP reverse proxy and replace an already registered proxy.
func (manager *Manager) CreateAgentProxyServer(endpoint *portaineree.Endpoint) (*factory.ProxyServer, error) {
	return manager.proxyFactory.NewAgentProxy(endpoint)
}

// GetEndpointProxy returns the proxy associated to a key
func (manager *Manager) GetEndpointProxy(endpoint *portaineree.Endpoint) http.Handler {
	proxy, ok := manager.endpointProxies.Get(fmt.Sprint(endpoint.ID))
	if !ok {
		return nil
	}

	return proxy.(http.Handler)
}

// DeleteEndpointProxy deletes the proxy associated to a key
// and cleans the k8s environment(endpoint) client cache. DeleteEndpointProxy
// is currently only called for edge connection clean up.
func (manager *Manager) DeleteEndpointProxy(endpointID portaineree.EndpointID) {
	manager.endpointProxies.Remove(fmt.Sprint(endpointID))
	manager.k8sClientFactory.RemoveKubeClient(endpointID)
}

// CreateLegacyExtensionProxy creates a new HTTP reverse proxy for a legacy extension and adds it to the registered proxies
func (manager *Manager) CreateLegacyExtensionProxy(key, extensionAPIURL string) (http.Handler, error) {
	proxy, err := manager.proxyFactory.NewLegacyExtensionProxy(extensionAPIURL)
	if err != nil {
		return nil, err
	}

	manager.legacyExtensionProxies.Set(key, proxy)
	return proxy, nil
}

// GetLegacyExtensionProxy returns a legacy extension proxy associated to a key
func (manager *Manager) GetLegacyExtensionProxy(key string) http.Handler {
	proxy, ok := manager.legacyExtensionProxies.Get(key)
	if !ok {
		return nil
	}
	return proxy.(http.Handler)
}

// CreateGitlabProxy creates a new HTTP reverse proxy that can be used to send requests to the Gitlab API
func (manager *Manager) CreateGitlabProxy(url string) (http.Handler, error) {
	return manager.proxyFactory.NewGitlabProxy(url)
}

package factory

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
)

const azureAPIBaseURL = "https://management.azure.com"

type (
	// ProxyFactory is a factory to create reverse proxies
	ProxyFactory struct {
		dataStore                   dataservices.DataStore
		signatureService            portainer.DigitalSignatureService
		reverseTunnelService        portaineree.ReverseTunnelService
		dockerClientFactory         *client.ClientFactory
		kubernetesClientFactory     *cli.ClientFactory
		kubernetesTokenCacheManager *kubernetes.TokenCacheManager
		authService                 *authorization.Service
		userActivityService         portaineree.UserActivityService
		gitService                  portainer.GitService
	}
)

// NewProxyFactory returns a pointer to a new instance of a ProxyFactory
func NewProxyFactory(
	dataStore dataservices.DataStore,
	signatureService portainer.DigitalSignatureService,
	tunnelService portaineree.ReverseTunnelService,
	clientFactory *client.ClientFactory,
	kubernetesClientFactory *cli.ClientFactory,
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager,
	authService *authorization.Service,
	userActivityService portaineree.UserActivityService,
	gitService portainer.GitService,
) *ProxyFactory {
	return &ProxyFactory{
		dataStore:                   dataStore,
		signatureService:            signatureService,
		reverseTunnelService:        tunnelService,
		dockerClientFactory:         clientFactory,
		kubernetesClientFactory:     kubernetesClientFactory,
		kubernetesTokenCacheManager: kubernetesTokenCacheManager,
		authService:                 authService,
		userActivityService:         userActivityService,
		gitService:                  gitService,
	}
}

// NewEndpointProxy returns a new reverse proxy (filesystem based or HTTP) to an environment(endpoint) API server
func (factory *ProxyFactory) NewEndpointProxy(endpoint *portaineree.Endpoint) (http.Handler, error) {
	switch endpoint.Type {
	case portaineree.AzureEnvironment:
		return newAzureProxy(factory.userActivityService, endpoint, factory.dataStore)
	case portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.AgentOnKubernetesEnvironment, portaineree.KubernetesLocalEnvironment:
		return factory.newKubernetesProxy(endpoint)
	}

	return factory.newDockerProxy(endpoint)
}

// NewGitlabProxy returns a new HTTP proxy to a Gitlab API server
func (factory *ProxyFactory) NewGitlabProxy(gitlabAPIUri string) (http.Handler, error) {
	return newGitlabProxy(gitlabAPIUri, factory.userActivityService)
}

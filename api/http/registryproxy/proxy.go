package registryproxy

import (
	"net/http"
	"strings"

	cmap "github.com/orcaman/concurrent-map"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/crypto"
)

// Service represents a service used to manage registry proxies.
type Service struct {
	proxies             cmap.ConcurrentMap
	userActivityService portaineree.UserActivityService
}

// NewService returns a pointer to a Service.
func NewService(userActivityService portaineree.UserActivityService) *Service {
	return &Service{
		proxies:             cmap.New(),
		userActivityService: userActivityService,
	}
}

// GetProxy returns the registry proxy associated to a key if it exists.
// Otherwise, it will create it and return it.
func (service *Service) GetProxy(key, uri string, config *portaineree.RegistryManagementConfiguration, forceCreate bool) (http.Handler, error) {
	proxy, ok := service.proxies.Get(key)
	if ok && !forceCreate {
		return proxy.(http.Handler), nil
	}

	return service.createProxy(key, uri, config)
}

// DeleteProxy deletes the registry proxy associated to a key.
func (service *Service) DeleteProxy(key string) {
	service.proxies.Remove(key)
}

func (service *Service) createProxy(key, uri string, config *portaineree.RegistryManagementConfiguration) (http.Handler, error) {
	var proxy http.Handler
	var err error
	transport := &http.Transport{}

	switch config.Type {
	case portaineree.AzureRegistry, portaineree.EcrRegistry:
		proxy, err = newTokenSecuredRegistryProxy(uri, config, NewLoggingTransport(service.userActivityService, transport))
	case portaineree.GithubRegistry:
		proxy, err = newGithubRegistryProxy(uri, config, NewLoggingTransport(service.userActivityService, transport))
	case portaineree.GitlabRegistry:
		if strings.Contains(key, "gitlab") {
			proxy, err = newGitlabRegistryProxy(uri, config, NewLoggingTransport(service.userActivityService, transport))
		} else {
			proxy, err = newTokenSecuredRegistryProxy(uri, config, NewLoggingTransport(service.userActivityService, transport))
		}
	default:
		if config.TLSConfig.TLS {
			tlsConfig, err := crypto.CreateTLSConfigurationFromDisk(config.TLSConfig.TLSCACertPath, config.TLSConfig.TLSCertPath, config.TLSConfig.TLSKeyPath, config.TLSConfig.TLSSkipVerify)
			if err != nil {
				return nil, err
			}

			transport.TLSClientConfig = tlsConfig
		}

		proxy, err = newCustomRegistryProxy(uri, config, NewLoggingTransport(service.userActivityService, transport))
	}

	if err != nil {
		return nil, err
	}

	service.proxies.Set(key, proxy)
	return proxy, nil
}

package registryproxy

import (
	"net/http"
	"strings"

	cmap "github.com/orcaman/concurrent-map"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
)

// Service represents a service used to manage registry proxies.
type Service struct {
	proxies             cmap.ConcurrentMap
	userActivityService portainer.UserActivityService
}

// NewService returns a pointer to a Service.
func NewService(userActivityService portainer.UserActivityService) *Service {
	return &Service{
		proxies:             cmap.New(),
		userActivityService: userActivityService,
	}
}

// GetProxy returns the registry proxy associated to a key if it exists.
// Otherwise, it will create it and return it.
func (service *Service) GetProxy(key, uri string, config *portainer.RegistryManagementConfiguration, forceCreate bool) (http.Handler, error) {
	proxy, ok := service.proxies.Get(key)
	if ok && !forceCreate {
		return proxy.(http.Handler), nil
	}

	return service.createProxy(key, uri, config)
}

func (service *Service) createProxy(key, uri string, config *portainer.RegistryManagementConfiguration) (http.Handler, error) {
	var proxy http.Handler
	var err error
	transport := &http.Transport{}

	switch config.Type {
	case portainer.AzureRegistry, portainer.EcrRegistry:
		proxy, err = newTokenSecuredRegistryProxy(uri, config, NewLoggingTransport(service.userActivityService, transport))
	case portainer.GitlabRegistry:
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

package images

import (
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"time"
)

var (
	_registriesCache = cache.New(5*time.Minute, 5*time.Minute)
)

type RegistryClient struct {
	registryService dataservices.RegistryService
}

func NewRegistryClient(registryService dataservices.RegistryService) *RegistryClient {
	return &RegistryClient{registryService: registryService}
}

func (c *RegistryClient) getRegistry(domain string) (*portaineree.Registry, error) {
	registry, err := cachedRegistry(domain)
	if err == nil {
		return registry, nil
	}

	registries, err := c.registryService.Registries()
	if err != nil {
		return nil, err
	}

	for _, r := range registries {
		_registriesCache.Set(r.URL, r, 0)
	}

	return cachedRegistry(domain)
}

func cachedRegistry(domain string) (*portaineree.Registry, error) {
	cachedRegistry, ok := _registriesCache.Get(domain)
	if ok {
		registry, ok := cachedRegistry.(portaineree.Registry)
		if ok {
			return &registry, nil
		}
	}

	return nil, errors.Errorf("no registry found in cache: %s", domain)
}

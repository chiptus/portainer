package images

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
)

var (
	_registriesCache = cache.New(5*time.Minute, 5*time.Minute)
)

type (
	RegistryClient struct {
		dataStore dataservices.DataStore
	}
)

func NewRegistryClient(dataStore dataservices.DataStore) *RegistryClient {
	return &RegistryClient{dataStore: dataStore}
}

func (c *RegistryClient) RegistryAuth(image Image) (string, string, error) {
	registries, err := c.dataStore.Registry().ReadAll()
	if err != nil {
		return "", "", err
	}

	registry, err := findBestMatchRegistry(image.opts.Name, registries)
	if err != nil {
		return "", "", err
	}

	if !registry.Authentication {
		return "", "", errors.New("authentication is disabled")
	}

	return c.CertainRegistryAuth(registry)
}

func (c *RegistryClient) CertainRegistryAuth(registry *portaineree.Registry) (string, string, error) {
	err := registryutils.EnsureRegTokenValid(c.dataStore, registry)
	if err != nil {
		return "", "", err
	}

	if !registry.Authentication {
		return "", "", errors.New("authentication is disabled")
	}

	return registryutils.GetRegEffectiveCredential(registry)
}

func (c *RegistryClient) EncodedRegistryAuth(image Image) (string, error) {
	registries, err := c.dataStore.Registry().ReadAll()
	if err != nil {
		return "", err
	}

	registry, err := findBestMatchRegistry(image.opts.Name, registries)
	if err != nil {
		return "", err
	}

	if !registry.Authentication {
		return "", errors.New("authentication is disabled")
	}

	return c.EncodedCertainRegistryAuth(registry)
}

func (c *RegistryClient) EncodedCertainRegistryAuth(registry *portaineree.Registry) (string, error) {
	err := registryutils.EnsureRegTokenValid(c.dataStore, registry)
	if err != nil {
		return "", err
	}

	return registryutils.GetRegistryAuthHeader(registry)
}

// findBestMatchRegistry finds out the best match registry for repository @Meng
// matching precedence:
// 1. both domain name and username matched (for dockerhub only)
// 2. only URL matched
// 3. pick up the first dockerhub registry
func findBestMatchRegistry(repository string, registries []portaineree.Registry) (*portaineree.Registry, error) {
	cachedRegistry, err := cachedRegistry(repository)
	if err == nil {
		return cachedRegistry, nil
	}

	var match1, match2, match3 *portaineree.Registry
	for i := 0; i < len(registries); i++ {
		registry := registries[i]
		if registry.Type == portaineree.DockerHubRegistry {

			// try to match repository examples:
			//   <USERNAME>/nginx:latest
			//   docker.io/<USERNAME>/nginx:latest
			if strings.HasPrefix(repository, registry.Username+"/") || strings.HasPrefix(repository, registry.URL+"/"+registry.Username+"/") {
				match1 = &registry
			}

			// try to match repository examples:
			//   portainer/portainer-ee:latest
			//   <NON-USERNAME>/portainer-ee:latest
			if match3 == nil {
				match3 = &registry
			}
		}

		if strings.Contains(repository, registry.URL) {
			match2 = &registry
		}
	}

	match := match1
	if match == nil {
		match = match2
	}
	if match == nil {
		match = match3
	}

	if match == nil {
		return nil, errors.New("no registries matched")
	}
	_registriesCache.Set(repository, match, 0)
	return match, nil
}

func cachedRegistry(cacheKey string) (*portaineree.Registry, error) {
	r, ok := _registriesCache.Get(cacheKey)
	if ok {
		registry, ok := r.(portaineree.Registry)
		if ok {
			return &registry, nil
		}
	}

	return nil, errors.Errorf("no registry found in cache: %s", cacheKey)
}

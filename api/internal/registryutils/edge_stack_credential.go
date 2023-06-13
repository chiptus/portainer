package registryutils

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer/api/edge"
	log "github.com/rs/zerolog/log"
)

func GetRegistryCredentialsForEdgeStack(dataStore dataservices.DataStoreTx, stack *portaineree.EdgeStack, endpoint *portaineree.Endpoint) []edge.RegistryCredentials {
	registries := []edge.RegistryCredentials{}
	for _, id := range stack.Registries {
		registry, _ := dataStore.Registry().Registry(id)

		registryCredential := GetRegistryCredential(registry)
		if registryCredential != nil {
			registries = append(registries, *registryCredential)
		}
	}

	// Only provide registry credentials if we are sure that the agent connection is https
	// We will still allow the stack deployment to be attempted without credentials so that
	// failure can be seen rather than having the stack sit in deploying state forever
	if len(registries) > 0 && !secureEndpoint(endpoint) {
		environmentType := "environment"

		log.Warn().
			Str("type", environmentType).
			Str("environment", endpoint.Name).
			Msg("the environment was deployed using HTTP and is insecure (registry credentials withheld), to use private registries, please update it to use HTTPS")

		registries = []edge.RegistryCredentials{}
	}

	return registries
}

func GetRegistryCredential(registry *portaineree.Registry) *edge.RegistryCredentials {
	registryCredential := &edge.RegistryCredentials{}
	if registry != nil {
		var username string
		var password string

		if registry.Type == portaineree.EcrRegistry {
			config := portaineree.RegistryManagementConfiguration{
				Type: portaineree.EcrRegistry,
				Ecr:  registry.Ecr,
			}
			config.Username, config.Password, _ = GetManagementCredential(registry)
			EnsureManageTokenValid(&config)

			// Now split the token into username and password (separator is :)
			s := strings.Split(config.AccessToken, ":")
			username = s[0]
			password = s[1]
		} else {
			username = registry.Username
			password = registry.Password
		}

		registryCredential.ServerURL = registry.URL
		registryCredential.Username = username
		registryCredential.Secret = password
		return registryCredential
	}
	return nil
}

// secureEndpoint returns true if the endpoint is secure, false otherwise
// security is determined by the scheme being https.  We use the edge key because
// it's gauranteed not to have been altered
func secureEndpoint(endpoint *portaineree.Endpoint) bool {
	portainerUrl, error := getPortainerServerUrlFromEdgeKey(endpoint.EdgeKey)
	if error != nil {
		return false
	}

	u, err := url.Parse(portainerUrl)
	if err != nil {
		return false
	}

	return u.Scheme == "https"
}

// getPortainerServerUrlFromEdgeKey decodes a base64 encoded key and extract the portainer server URL
// edge key format: <portainer_instance_url>|<tunnel_server_addr>|<tunnel_server_fingerprint>|<endpoint_id>
func getPortainerServerUrlFromEdgeKey(key string) (string, error) {
	decodedKey, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	keyInfo := strings.Split(string(decodedKey), "|")

	if len(keyInfo) != 4 {
		return "", errors.New("invalid key format")
	}

	return keyInfo[0], nil
}

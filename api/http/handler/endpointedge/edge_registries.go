package endpointedge

import (
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
)

func (handler *Handler) getRegistryCredentialsForEdgeStack(stack *portaineree.EdgeStack) []Credentials {
	registries := []Credentials{}
	for _, id := range stack.Registries {
		registry, _ := handler.DataStore.Registry().Registry(id)
		if registry != nil {
			var username string
			var password string

			if registry.Type == portaineree.EcrRegistry {
				config := portaineree.RegistryManagementConfiguration{
					Type: portaineree.EcrRegistry,
					Ecr:  registry.Ecr,
				}
				config.Username, config.Password, _ = registryutils.GetManagementCredential(registry)
				registryutils.EnsureManageTokenValid(&config)

				// Now split the token into username and password (separator is :)
				s := strings.Split(config.AccessToken, ":")
				username = s[0]
				password = s[1]
			} else {
				username = registry.Username
				password = registry.Password
			}

			registries = append(registries, Credentials{
				ServerURL: registry.URL,
				Username:  username,
				Secret:    password,
			})
		}
	}

	return registries
}

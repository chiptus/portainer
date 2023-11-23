package docker

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	portainer "github.com/portainer/portainer/api"
)

type (
	// isAdmin = true for admin and edge admin
	registryAccessContext struct {
		isAdmin         bool
		user            *portaineree.User
		endpointID      portainer.EndpointID
		teamMemberships []portainer.TeamMembership
		registries      []portaineree.Registry
	}

	registryAuthenticationHeader struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		Serveraddress string `json:"serveraddress"`
	}

	portainerRegistryAuthenticationHeader struct {
		RegistryId *portainer.RegistryID `json:"registryId"`
	}
)

func createRegistryAuthenticationHeader(
	dataStore dataservices.DataStore,
	registryId portainer.RegistryID,
	accessContext *registryAccessContext,
) (authenticationHeader registryAuthenticationHeader, err error) {
	if registryId == 0 { // dockerhub (anonymous)
		authenticationHeader.Serveraddress = "docker.io"
	} else { // any "custom" registry
		var matchingRegistry *portaineree.Registry
		for _, registry := range accessContext.registries {
			registry := registry
			if registry.ID == registryId &&
				(accessContext.isAdmin ||
					security.AuthorizedRegistryAccess(&registry, accessContext.user, accessContext.teamMemberships, accessContext.endpointID)) {
				matchingRegistry = &registry
				break
			}
		}

		if matchingRegistry != nil {
			err = registryutils.EnsureRegTokenValid(dataStore, matchingRegistry)
			if err != nil {
				return
			}
			authenticationHeader.Serveraddress = matchingRegistry.URL
			authenticationHeader.Username, authenticationHeader.Password, err = registryutils.GetRegEffectiveCredential(matchingRegistry)
		}
	}

	return
}

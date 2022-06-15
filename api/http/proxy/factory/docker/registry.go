package docker

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
)

type (
	registryAccessContext struct {
		isAdmin         bool
		user            *portaineree.User
		endpointID      portaineree.EndpointID
		teamMemberships []portaineree.TeamMembership
		registries      []portaineree.Registry
	}

	registryAuthenticationHeader struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		Serveraddress string `json:"serveraddress"`
	}

	portainerRegistryAuthenticationHeader struct {
		RegistryId *portaineree.RegistryID `json:"registryId"`
	}
)

func createRegistryAuthenticationHeader(
	dataStore dataservices.DataStore,
	registryId portaineree.RegistryID,
	accessContext *registryAccessContext,
) (authenticationHeader registryAuthenticationHeader, err error) {
	if registryId == 0 { // dockerhub (anonymous)
		authenticationHeader.Serveraddress = "docker.io"
	} else { // any "custom" registry
		var matchingRegistry *portaineree.Registry
		for _, registry := range accessContext.registries {
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

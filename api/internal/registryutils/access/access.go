package access

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
)

func hasPermission(
	dataStore dataservices.DataStore,
	userID portainer.UserID,
	endpointID portainer.EndpointID,
	registry *portaineree.Registry,
) (hasPermission bool, err error) {
	user, err := dataStore.User().Read(userID)
	if err != nil {
		return
	}

	if user.Role == portaineree.AdministratorRole {
		return true, err
	}

	teamMemberships, err := dataStore.TeamMembership().TeamMembershipsByUserID(userID)
	if err != nil {
		return
	}

	hasPermission = security.AuthorizedRegistryAccess(registry, user, teamMemberships, endpointID)

	return
}

// GetAccessibleRegistry get the registry if the user has permission
func GetAccessibleRegistry(
	dataStore dataservices.DataStore,
	userID portainer.UserID,
	endpointID portainer.EndpointID,
	registryID portainer.RegistryID,
) (registry *portaineree.Registry, err error) {

	registry, err = dataStore.Registry().Read(registryID)
	if err != nil {
		return
	}

	hasPermission, err := hasPermission(dataStore, userID, endpointID, registry)
	if err != nil {
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not has permission to get the registry")
		return nil, err
	}

	return
}

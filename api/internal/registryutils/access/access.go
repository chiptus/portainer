package access

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
)

func hasPermission(
	dataStore dataservices.DataStore,
	userID portaineree.UserID,
	endpointID portaineree.EndpointID,
	registry *portaineree.Registry,
) (hasPermission bool, err error) {
	user, err := dataStore.User().User(userID)
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
	userID portaineree.UserID,
	endpointID portaineree.EndpointID,
	registryID portaineree.RegistryID,
) (registry *portaineree.Registry, err error) {

	registry, err = dataStore.Registry().Registry(registryID)
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

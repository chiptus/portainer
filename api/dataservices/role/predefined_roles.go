package role

import (
	"sort"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer/api/dataservices/errors"
)

var PredefinedRoles = map[portaineree.RoleID]*portaineree.Role{
	portaineree.RoleIDEndpointAdmin: {
		Name:           "Environment administrator",
		Description:    "Full control of all resources in an environment",
		ID:             portaineree.RoleIDEndpointAdmin,
		Priority:       1,
		Authorizations: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
	},
	portaineree.RoleIDHelpdesk: {
		Name:           "Helpdesk",
		Description:    "Read-only access of all resources in an environment",
		ID:             portaineree.RoleIDHelpdesk,
		Priority:       3,
		Authorizations: authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
	},
	portaineree.RoleIDStandardUser: {
		Name:           "Standard user",
		Description:    "Full control of assigned resources in an environment",
		ID:             portaineree.RoleIDStandardUser,
		Priority:       4,
		Authorizations: authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
	},
	portaineree.RoleIDReadonly: {
		Name:           "Read-only user",
		Description:    "Read-only access of assigned resources in an environment",
		ID:             portaineree.RoleIDReadonly,
		Priority:       5,
		Authorizations: authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(),
	},
	portaineree.RoleIDOperator: {
		Name:           "Operator",
		Description:    "Operational control of all existing resources in an environment",
		ID:             portaineree.RoleIDOperator,
		Priority:       2,
		Authorizations: authorization.DefaultEndpointAuthorizationsForOperatorRole(),
	},
}

// CreateOrUpdatePredefinedRoles update the predefined roles. Create one if it does not exist yet.
func (service *Service) CreateOrUpdatePredefinedRoles() error {

	// The order of iteration over map is undefined and may vary between program to program so
	// to insert in the right order, creating a roles []int and sorting it.
	roles := []int{}
	for roleID := range PredefinedRoles {
		roles = append(roles, int(roleID))
	}
	sort.Ints(roles)

	for _, roleid := range roles {
		roleID := portaineree.RoleID(roleid)
		predefinedRole := PredefinedRoles[roleID]

		_, err := service.Role(roleID)
		if err == errors.ErrObjectNotFound {
			err := service.Create(predefinedRole)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			err = service.UpdateRole(predefinedRole.ID, predefinedRole)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

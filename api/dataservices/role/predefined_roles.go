package role

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer/api/dataservices/errors"
)

// CreateOrUpdatePredefinedRoles update the predefined roles. Create one if it does not exist yet.
func (service *Service) CreateOrUpdatePredefinedRoles() error {
	predefinedRoles := map[portaineree.RoleID]*portaineree.Role{
		portaineree.RoleIDEndpointAdmin: {
			Name:           "Environment administrator",
			Description:    "Full control of all resources in an environment",
			ID:             portaineree.RoleIDEndpointAdmin,
			Priority:       1,
			Authorizations: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
		},
		portaineree.RoleIDOperator: {
			Name:           "Operator",
			Description:    "Operational control of all existing resources in an environment",
			ID:             portaineree.RoleIDOperator,
			Priority:       2,
			Authorizations: authorization.DefaultEndpointAuthorizationsForOperatorRole(),
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
	}

	for roleID, predefinedRole := range predefinedRoles {
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

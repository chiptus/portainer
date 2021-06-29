package migrator

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/authorization"
)

// updateRBACRoles updates roles to current defaults
// running it after changing one of `authorization.DefaultEndpointAuthorizations`
// will update the role
func (m *Migrator) refreshRBACRoles() error {
	defaultAuthorizationsOfRoles := map[portainer.RoleID]portainer.Authorizations{
		portainer.RoleIDEndpointAdmin: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
		portainer.RoleIDHelpdesk:      authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
		portainer.RoleIDOperator:      authorization.DefaultEndpointAuthorizationsForOperatorRole(),
		portainer.RoleIDStandardUser:  authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
		portainer.RoleIDReadonly:      authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(),
	}

	for roleID, defaultAuthorizations := range defaultAuthorizationsOfRoles {
		role, err := m.roleService.Role(roleID)
		if err != nil {
			return err
		}
		role.Authorizations = defaultAuthorizations

		err = m.roleService.UpdateRole(role.ID, role)
		if err != nil {
			return err
		}
	}

	return m.authorizationService.UpdateUsersAuthorizations()
}

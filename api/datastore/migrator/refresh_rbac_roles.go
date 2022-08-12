package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// refreshRBACRoles updates roles to current defaults
// running it after changing one of `authorization.DefaultEndpointAuthorizations`
// will update the role
func (m *Migrator) refreshRBACRoles() error {
	migrateLog.Info("- refreshing RBAC roles")
	defaultAuthorizationsOfRoles := map[portaineree.RoleID]portaineree.Authorizations{
		portaineree.RoleIDEndpointAdmin: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
		portaineree.RoleIDHelpdesk:      authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
		portaineree.RoleIDOperator:      authorization.DefaultEndpointAuthorizationsForOperatorRole(),
		portaineree.RoleIDStandardUser:  authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
		portaineree.RoleIDReadonly:      authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(),
	}

	for roleID, defaultAuthorizations := range defaultAuthorizationsOfRoles {
		role, err := m.roleService.Role(roleID)
		if err != nil {
			if err == bolterrors.ErrObjectNotFound {
				continue
			} else {
				return err
			}
		}
		role.Authorizations = defaultAuthorizations

		err = m.roleService.UpdateRole(role.ID, role)
		if err != nil {
			return err
		}
	}

	return m.authorizationService.UpdateUsersAuthorizations()
}

func (m *Migrator) refreshUserAuthorizations() error {
	migrateLog.Info("- refreshing user authorizations")
	users, err := m.userService.Users()
	if err != nil {
		return err
	}

	for _, user := range users {
		user.PortainerAuthorizations = authorization.DefaultPortainerAuthorizations()

		err = m.userService.UpdateUser(user.ID, &user)
		if err != nil {
			return err
		}
	}

	return nil
}

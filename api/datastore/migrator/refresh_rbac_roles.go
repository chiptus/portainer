package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"

	"github.com/rs/zerolog/log"
)

// refreshRBACRoles updates roles to current defaults
// running it after changing one of `authorization.DefaultEndpointAuthorizations`
// will update the role
func (m *Migrator) refreshRBACRoles() error {
	log.Info().Msg("refreshing RBAC roles")

	defaultAuthorizationsOfRoles := map[portaineree.RoleID]portaineree.Authorizations{
		portaineree.RoleIDEndpointAdmin: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
		portaineree.RoleIDHelpdesk:      authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
		portaineree.RoleIDOperator:      authorization.DefaultEndpointAuthorizationsForOperatorRole(),
		portaineree.RoleIDStandardUser:  authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
		portaineree.RoleIDReadonly:      authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(),
	}

	for roleID, defaultAuthorizations := range defaultAuthorizationsOfRoles {
		role, err := m.roleService.Read(roleID)
		if err != nil {
			if dataservices.IsErrObjectNotFound(err) {
				continue
			} else {
				return err
			}
		}
		role.Authorizations = defaultAuthorizations

		err = m.roleService.Update(role.ID, role)
		if err != nil {
			return err
		}
	}

	return m.authorizationService.UpdateUsersAuthorizations()
}

func (m *Migrator) refreshUserAuthorizations() error {
	log.Info().Msg("refreshing user authorizations")
	users, err := m.userService.ReadAll()
	if err != nil {
		return err
	}

	for _, user := range users {
		user.PortainerAuthorizations = authorization.DefaultPortainerAuthorizations()

		err = m.userService.Update(user.ID, &user)
		if err != nil {
			return err
		}
	}

	return nil
}

package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

func (m *Migrator) updateResourceControlsToDBVersion22() error {
	legacyResourceControls, err := m.resourceControlService.ResourceControls()
	if err != nil {
		return err
	}

	for _, resourceControl := range legacyResourceControls {
		resourceControl.AdministratorsOnly = false

		err := m.resourceControlService.UpdateResourceControl(resourceControl.ID, &resourceControl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateUsersAndRolesToDBVersion22() error {
	legacyUsers, err := m.userService.Users()
	if err != nil {
		return err
	}

	for _, user := range legacyUsers {
		user.PortainerAuthorizations = authorization.DefaultPortainerAuthorizations()
		err = m.userService.UpdateUser(user.ID, &user)
		if err != nil {
			return err
		}
	}

	endpointAdministratorRole, err := m.roleService.Role(portaineree.RoleID(1))
	if err != nil {
		return err
	}
	endpointAdministratorRole.Priority = 1
	endpointAdministratorRole.Authorizations = authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole()

	err = m.roleService.UpdateRole(endpointAdministratorRole.ID, endpointAdministratorRole)

	helpDeskRole, err := m.roleService.Role(portaineree.RoleID(2))
	if err != nil {
		return err
	}
	helpDeskRole.Priority = 2
	helpDeskRole.Authorizations = authorization.DefaultEndpointAuthorizationsForHelpDeskRole()

	err = m.roleService.UpdateRole(helpDeskRole.ID, helpDeskRole)

	standardUserRole, err := m.roleService.Role(portaineree.RoleID(3))
	if err != nil {
		return err
	}
	standardUserRole.Priority = 3
	standardUserRole.Authorizations = authorization.DefaultEndpointAuthorizationsForStandardUserRole()

	err = m.roleService.UpdateRole(standardUserRole.ID, standardUserRole)

	readOnlyUserRole, err := m.roleService.Role(portaineree.RoleID(4))
	if err != nil {
		return err
	}
	readOnlyUserRole.Priority = 4
	readOnlyUserRole.Authorizations = authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole()

	err = m.roleService.UpdateRole(readOnlyUserRole.ID, readOnlyUserRole)
	if err != nil {
		return err
	}

	return m.authorizationService.UpdateUsersAuthorizations()
}
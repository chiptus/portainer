package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB36() error {
	return m.migrateUsersToDB36()
}

func (m *Migrator) migrateUsersToDB36() error {
	log.Info().Msg("updating user authorizations")

	users, err := m.userService.Users()
	if err != nil {
		return err
	}

	for _, user := range users {
		currentAuthorizations := authorization.DefaultPortainerAuthorizations()
		currentAuthorizations[portaineree.OperationPortainerUserListToken] = true
		currentAuthorizations[portaineree.OperationPortainerUserCreateToken] = true
		currentAuthorizations[portaineree.OperationPortainerUserRevokeToken] = true
		user.PortainerAuthorizations = currentAuthorizations
		err = m.userService.UpdateUser(user.ID, &user)
		if err != nil {
			return err
		}
	}

	return nil
}

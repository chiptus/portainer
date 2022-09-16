package migrator

import (
	"github.com/portainer/portainer-ee/api/datastore/validate"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (m *Migrator) updateUsersAndRolesToDBVersion28() error {
	log.Info().Msg("updating users and roles")

	err := m.roleService.CreateOrUpdatePredefinedRoles()
	if err != nil {
		return err
	}

	err = m.authorizationService.UpdateUsersAuthorizations()
	if err != nil {
		return err
	}

	roles, err := m.roleService.Roles()
	if err != nil {
		return errors.Wrap(err, "while getting roles from db")
	}

	err = validate.ValidatePredefinedRoles(roles)
	if err != nil {
		return errors.Wrap(err, "while validating roles")
	}

	return nil
}

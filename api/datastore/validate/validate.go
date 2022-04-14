package validate

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices/role"
)

var validate *validator.Validate

func ValidateLDAPSettings(ldp *portaineree.LDAPSettings) error {
	validate = validator.New()
	registerValidationMethods(validate)

	return validate.Struct(ldp)
}

func ValidatePredefinedRoles(roles []portaineree.Role) error {
	if len(roles) < len(role.PredefinedRoles) {
		return fmt.Errorf("predefined roles missing. Want=%d, Got=%d", len(role.PredefinedRoles), len(roles))
	}

	for _, rol := range role.PredefinedRoles {
		if !isRoleExists(roles, *rol) {
			return fmt.Errorf("role %s missing in the DB", rol.Name)
		}
	}

	return nil
}

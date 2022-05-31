package customtemplates

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
)

func validateVariablesDefinitions(variables []portaineree.CustomTemplateVariableDefinition) error {
	for _, variable := range variables {
		if variable.Name == "" {
			return errors.New("variable name is required")
		}
		if variable.Label == "" {
			return errors.New("variable label is required")
		}
	}
	return nil
}

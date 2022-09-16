package migrator

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/dataservices/errors"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) updateStackResourceControlToDB27() error {
	log.Info().Msg("updating stack resource controls")

	resourceControls, err := m.resourceControlService.ResourceControls()
	if err != nil {
		return err
	}

	for _, resource := range resourceControls {
		if resource.Type != portaineree.StackResourceControl {
			continue
		}

		stackName := resource.ResourceID
		if err != nil {
			return err
		}

		stack, err := m.stackService.StackByName(stackName)
		if err != nil {
			if err == errors.ErrObjectNotFound {
				continue
			}

			return err
		}

		resource.ResourceID = fmt.Sprintf("%d_%s", stack.EndpointID, stack.Name)

		err = m.resourceControlService.UpdateResourceControl(resource.ID, &resource)
		if err != nil {
			return err
		}
	}

	return nil
}

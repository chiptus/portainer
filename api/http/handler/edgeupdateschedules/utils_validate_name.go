package edgeupdateschedules

import (
	"github.com/pkg/errors"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
)

func (handler *Handler) validateUniqueName(name string, id edgetypes.UpdateScheduleID) error {
	list, err := handler.updateService.Schedules()
	if err != nil {
		return errors.WithMessage(err, "Unable to list edge update schedules")
	}

	for _, schedule := range list {
		if id != schedule.ID && schedule.Name == name {
			return errors.New("Edge update schedule name already in use")
		}
	}

	return nil
}

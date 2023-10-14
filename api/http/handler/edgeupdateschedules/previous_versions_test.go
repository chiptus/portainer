package edgeupdateschedules

import (
	"testing"

	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
)

func TestPreviousVersions(t *testing.T) {

	activeSchedulesMap := map[portainer.EndpointID]*edgetypes.EndpointUpdateScheduleRelation{}

	schedules := []edgetypes.UpdateSchedule{
		{
			ID:   1,
			Type: edgetypes.UpdateScheduleUpdate,
			EnvironmentsPreviousVersions: map[portainer.EndpointID]string{
				1: "2.11.0",

				2: "2.12.0",
			},
			Created: 1500000000,
		},
		{
			ID:   2,
			Type: edgetypes.UpdateScheduleRollback,
			EnvironmentsPreviousVersions: map[portainer.EndpointID]string{
				1: "2.14.0",
			},
			Created: 1500000001,
		},
		{
			ID:   3,
			Type: edgetypes.UpdateScheduleUpdate,
			EnvironmentsPreviousVersions: map[portainer.EndpointID]string{
				2: "2.13.0",
			},
			Created: 1500000002,
		},
	}

	actual := previousVersions(
		schedules,
		func(environmentID portainer.EndpointID) *edgetypes.EndpointUpdateScheduleRelation {
			return activeSchedulesMap[environmentID]
		},
		0,
	)

	assert.Equal(t, map[portainer.EndpointID]string{
		2: "2.13.0",
	}, actual)

}

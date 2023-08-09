package edgeupdateschedules

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"

	"github.com/stretchr/testify/assert"
)

func TestPreviousVersions(t *testing.T) {

	activeSchedulesMap := map[portaineree.EndpointID]*edgetypes.EndpointUpdateScheduleRelation{}

	schedules := []edgetypes.UpdateSchedule{
		{
			ID:   1,
			Type: edgetypes.UpdateScheduleUpdate,
			EnvironmentsPreviousVersions: map[portaineree.EndpointID]string{
				1: "2.11.0",

				2: "2.12.0",
			},
			Created: 1500000000,
		},
		{
			ID:   2,
			Type: edgetypes.UpdateScheduleRollback,
			EnvironmentsPreviousVersions: map[portaineree.EndpointID]string{
				1: "2.14.0",
			},
			Created: 1500000001,
		},
		{
			ID:   3,
			Type: edgetypes.UpdateScheduleUpdate,
			EnvironmentsPreviousVersions: map[portaineree.EndpointID]string{
				2: "2.13.0",
			},
			Created: 1500000002,
		},
	}

	actual := previousVersions(
		schedules,
		func(environmentID portaineree.EndpointID) *edgetypes.EndpointUpdateScheduleRelation {
			return activeSchedulesMap[environmentID]
		},
		0,
	)

	assert.Equal(t, map[portaineree.EndpointID]string{
		2: "2.13.0",
	}, actual)

}

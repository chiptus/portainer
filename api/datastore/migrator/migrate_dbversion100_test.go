package migrator

import (
	"testing"

	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
)

func TestPreviousVersions(t *testing.T) {

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

	actual := getEdgeUpdatesPreviousVersions(
		schedules,
	)

	assert.Equal(t, map[portainer.EndpointID]string{
		2: "2.13.0",
	}, actual)

}

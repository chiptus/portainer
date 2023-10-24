package edgeupdateschedules

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	portainer "github.com/portainer/portainer/api"
)

type mockUpdateService struct {
	schedules []edgetypes.UpdateSchedule
}

func (m *mockUpdateService) ActiveSchedule(environmentID portainer.EndpointID) *edgetypes.EndpointUpdateScheduleRelation {
	return nil
}

func (m *mockUpdateService) ActiveSchedules(environmentsIDs []portainer.EndpointID) []edgetypes.EndpointUpdateScheduleRelation {
	return nil
}

func (m *mockUpdateService) RemoveActiveSchedule(environmentID portainer.EndpointID, scheduleID edgetypes.UpdateScheduleID) error {
	return nil
}

func (m *mockUpdateService) EdgeStackDeployed(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID) {
}

func (m *mockUpdateService) Schedules(tx dataservices.DataStoreTx) ([]edgetypes.UpdateSchedule, error) {
	return m.schedules, nil
}

func (m *mockUpdateService) Schedule(scheduleID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error) {
	return nil, nil
}

func (m *mockUpdateService) CreateSchedule(tx dataservices.DataStoreTx, schedule *edgetypes.UpdateSchedule, metadata updateschedules.CreateMetadata) error {

	return nil
}

func (m *mockUpdateService) UpdateSchedule(tx dataservices.DataStoreTx, id edgetypes.UpdateScheduleID, item *edgetypes.UpdateSchedule, metadata updateschedules.CreateMetadata) error {
	return nil
}

func (m *mockUpdateService) DeleteSchedule(id edgetypes.UpdateScheduleID) error {
	return nil
}

func (m *mockUpdateService) HandleStatusChange(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID, status portainer.EdgeStackStatusType, agentVersion string) error {
	return nil
}

func TestFilterEnvironments(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	handler := &Handler{
		dataStore:     store,
		updateService: &mockUpdateService{},
	}

	/*
		1. regular
			3 endpoints with 3 different versions. the selected version is the highest (should return 2)
			3 endpoints with 3 different versions. the selected version is the lowest (should return 0)
			3 endpoints with 3 different versions. the selected version is higher then highest (should return 3)
		2. rollback
			2 endpoints with the same version, but with different previous version. the select version is the lowest (should return 1)
	*/

	// create endpoints
	endpoint1 := &portaineree.Endpoint{
		ID:   1,
		Name: "endpoint1",
		Type: portaineree.EdgeAgentOnDockerEnvironment,
		Agent: portaineree.EnvironmentAgentData{
			Version: "1.0.0",
		},
	}

	snapshot1 := &portaineree.Snapshot{
		EndpointID: 1,
		Docker: &portainer.DockerSnapshot{
			Swarm: false,
		},
	}

	endpoint2 := &portaineree.Endpoint{
		ID:   2,
		Name: "endpoint2",
		Type: portaineree.EdgeAgentOnDockerEnvironment,
		Agent: portaineree.EnvironmentAgentData{
			Version: "1.8.0",
		},
	}

	snapshot2 := &portaineree.Snapshot{
		EndpointID: 2,
		Docker: &portainer.DockerSnapshot{
			Swarm: false,
		},
	}

	endpoint3 := &portaineree.Endpoint{
		ID:   3,
		Name: "endpoint3",
		Type: portaineree.EdgeAgentOnDockerEnvironment,
		Agent: portaineree.EnvironmentAgentData{
			Version: "1.8.3",
		},
	}

	snapshot3 := &portaineree.Snapshot{
		EndpointID: 3,
		Docker: &portainer.DockerSnapshot{
			Swarm: false,
		},
	}

	endpoint4 := &portaineree.Endpoint{
		ID:   4,
		Name: "endpoint4",
		Type: portaineree.EdgeAgentOnDockerEnvironment,
		Agent: portaineree.EnvironmentAgentData{
			Version: "1.8.3",
		},
	}

	snapshot4 := &portaineree.Snapshot{
		EndpointID: 4,
		Docker: &portainer.DockerSnapshot{
			Swarm: false,
		},
	}

	endpoints := []*portaineree.Endpoint{
		endpoint1,
		endpoint2,
		endpoint3,
		endpoint4,
	}
	snapshots := []*portaineree.Snapshot{
		snapshot1,
		snapshot2,
		snapshot3,
		snapshot4,
	}

	for idx, endpoint := range endpoints {
		err := handler.dataStore.Endpoint().Create(endpoint)
		if err != nil {
			t.Fatal(err)
		}

		err = handler.dataStore.Snapshot().Create(snapshots[idx])
		if err != nil {
			t.Fatal(err)
		}
	}

	// create edge groups
	regularGroup := &portaineree.EdgeGroup{
		ID:        1,
		Name:      "edgeGroup1",
		TagIDs:    []portainer.TagID{},
		Endpoints: []portainer.EndpointID{1, 2, 3},
	}

	rollbackGroup := &portaineree.EdgeGroup{
		ID:        2,
		Name:      "edgeGroup2",
		TagIDs:    []portainer.TagID{},
		Endpoints: []portainer.EndpointID{3, 4},
	}

	edgeGroups := []*portaineree.EdgeGroup{
		regularGroup,
		rollbackGroup,
	}

	for _, edgeGroup := range edgeGroups {
		err := handler.dataStore.EdgeGroup().Create(edgeGroup)
		if err != nil {
			t.Fatal(err)
		}
	}

	type expectedResult struct {
		relatedEnvIds   []portainer.EndpointID
		currentVersions map[portainer.EndpointID]string
		envType         portainer.EndpointType
	}

	type testCase struct {
		name         string
		version      string
		environments []portainer.EndpointID
		expected     expectedResult
	}

	testCases := []testCase{
		{
			name:         "update - 3 endpoints with 3 different versions. the selected version is the highest (should return 2)",
			version:      "1.8.3",
			environments: []portainer.EndpointID{1, 2, 3},
			expected: expectedResult{
				relatedEnvIds: []portainer.EndpointID{1, 2},
				currentVersions: map[portainer.EndpointID]string{
					1: endpoint1.Agent.Version,
					2: endpoint2.Agent.Version,
				},
				envType: portaineree.EdgeAgentOnDockerEnvironment,
			},
		},
		{
			name:         "update - 3 endpoints with 3 different versions. the selected version is the lowest (should return 0)",
			version:      "1.0.0",
			environments: []portainer.EndpointID{1, 2, 3},
			expected: expectedResult{
				relatedEnvIds:   []portainer.EndpointID{},
				currentVersions: map[portainer.EndpointID]string{},
				envType:         portaineree.EdgeAgentOnDockerEnvironment,
			},
		},
		{
			name:         "update - 3 endpoints with 3 different versions. the selected version is higher then highest (should return 3)",
			version:      "1.8.4",
			environments: []portainer.EndpointID{1, 2, 3},
			expected: expectedResult{
				relatedEnvIds: []portainer.EndpointID{1, 2, 3},
				currentVersions: map[portainer.EndpointID]string{
					1: endpoint1.Agent.Version,
					2: endpoint2.Agent.Version,
					3: endpoint3.Agent.Version,
				},
				envType: portaineree.EdgeAgentOnDockerEnvironment,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			edgeGroup := &portaineree.EdgeGroup{
				Name:      tc.name,
				Endpoints: tc.environments,
			}

			err := handler.dataStore.EdgeGroup().Create(edgeGroup)
			if err != nil {
				t.Fatal(err)
			}

			// filter environments
			relatedEnvIds, envType, err := handler.filterEnvironments(handler.dataStore, []portainer.EdgeGroupID{edgeGroup.ID}, tc.version, false)
			if err != nil {
				if len(tc.expected.relatedEnvIds) == 0 && err.Error() == "no related environments that require update" {
					return
				}

				t.Fatal(err)
			}

			if len(relatedEnvIds) != len(tc.expected.relatedEnvIds) {
				t.Fatalf("expected %d related environments, got %d", len(tc.expected.relatedEnvIds), len(relatedEnvIds))
			}

			if envType != tc.expected.envType {
				t.Fatalf("expected env type to be %d, got %d", tc.expected.envType, envType)
			}

		})
	}

	t.Run("rollback - 2 endpoints with the same version, but with different previous version. the select version is the lowest (should return 1)", func(t *testing.T) {
		requestedVersion := "1.8.0"
		endpoint3.Agent.PreviousVersion = requestedVersion
		err := handler.dataStore.Endpoint().UpdateEndpoint(3, endpoint3)
		if err != nil {
			t.Fatal(err)
		}

		edgeGroup := &portaineree.EdgeGroup{
			Name:      "rollback - 2 endpoints with the same version, but with different previous version. the select version is the lowest (should return 1)",
			Endpoints: []portainer.EndpointID{3, 4},
		}

		err = handler.dataStore.EdgeGroup().Create(edgeGroup)
		if err != nil {
			t.Fatal(err)
		}

		handler.updateService = &mockUpdateService{
			schedules: []edgetypes.UpdateSchedule{
				{
					ID:          1,
					EdgeStackID: 1,
					Version:     endpoint3.Agent.Version,
				},
			},
		}

		relatedEnvIds, envType, err := handler.filterEnvironments(handler.dataStore, []portainer.EdgeGroupID{edgeGroup.ID}, requestedVersion, true)
		if err != nil {
			t.Fatal(err)
		}

		if len(relatedEnvIds) != 1 {
			t.Fatalf("expected 1 related environment, got %d", len(relatedEnvIds))
		}

		if envType != portaineree.EdgeAgentOnDockerEnvironment {
			t.Fatalf("expected env type to be %d, got %d", portaineree.EdgeAgentOnDockerEnvironment, envType)
		}

	})

}

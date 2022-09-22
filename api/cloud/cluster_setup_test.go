package cloud

import (
	"context"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
)

type MockedSnapshotService struct {
}

func (m *MockedSnapshotService) Start() {
	panic("not implemented")
}

func (m *MockedSnapshotService) SetSnapshotInterval(snapshotInterval string) error {
	panic("not implemented")
}

func (m *MockedSnapshotService) SnapshotEndpoint(endpoint *portaineree.Endpoint) error {
	panic("not implemented")
}

func (m *MockedSnapshotService) FillSnapshotData(endpoint *portaineree.Endpoint) error {
	panic("not implemented")
}

func TestChangeState(t *testing.T) {
	authorizationService := new(authorization.Service)
	snapshotService := new(MockedSnapshotService)
	clientFactory := new(kubecli.ClientFactory)

	requests := make(chan *portaineree.CloudProvisioningRequest, 10)
	result := make(chan *cloudPrevisioningResult, 10)

	tests := []struct {
		endpoint *portaineree.Endpoint
		task     *portaineree.CloudProvisioningTask
		state    ProvisioningState
		message  string
	}{
		{
			endpoint: &portaineree.Endpoint{},
			task: &portaineree.CloudProvisioningTask{
				EndpointID: 0,
				ClusterID:  "ID",
				State:      int(ProvisioningStatePending),
			},
			state:   ProvisioningStateWaitingForCluster,
			message: "Creating KaaS Cluster",
		},
		{
			endpoint: &portaineree.Endpoint{},
			task: &portaineree.CloudProvisioningTask{
				EndpointID: 0,
				ClusterID:  "",
				State:      int(ProvisioningStatePending),
			},
			state:   ProvisioningStateAgentSetup,
			message: "Deploying portainer agent",
		},
		{
			endpoint: &portaineree.Endpoint{},
			task: &portaineree.CloudProvisioningTask{
				EndpointID: 0,
				ClusterID:  "ID",
				State:      int(ProvisioningStatePending),
			},
			state:   ProvisioningStateWaitingForAgent,
			message: "Waiting for agent response",
		},
		{
			endpoint: &portaineree.Endpoint{},
			task: &portaineree.CloudProvisioningTask{
				EndpointID: 0,
				ClusterID:  "ID",
				State:      int(ProvisioningStatePending),
			},
			state:   ProvisioningStateUpdatingEndpoint,
			message: "Updating environment",
		},
		{
			endpoint: &portaineree.Endpoint{},
			task: &portaineree.CloudProvisioningTask{
				EndpointID: 0,
				ClusterID:  "civoID",
				State:      int(ProvisioningStatePending),
			},
			state:   ProvisioningStateDone,
			message: "Connecting",
		},
	}

	for _, test := range tests {
		var endpoints []portaineree.Endpoint
		endpoints = append(endpoints, *test.endpoint)
		dataStore := testhelpers.NewDatastore(testhelpers.WithEndpoints(endpoints))

		service := &CloudClusterSetupService{
			dataStore:            dataStore,
			shutdownCtx:          context.TODO(),
			requests:             requests,
			result:               result,
			snapshotService:      snapshotService,
			authorizationService: authorizationService,
			clientFactory:        clientFactory,
		}

		service.changeState(test.task, test.state, test.message)
		if test.task.State != int(test.state) {
			t.Error("failed setting task state in changeState")
		}
		endpoint, _ := service.dataStore.Endpoint().Endpoint(test.task.EndpointID)
		if endpoint.StatusMessage.Summary != test.message {
			t.Error("failed setting task message from changeState")
		}
	}
}

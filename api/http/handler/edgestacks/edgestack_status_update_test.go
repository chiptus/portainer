package edgestacks

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
)

// Update Status
func TestUpdateStatusAndInspect(t *testing.T) {
	handler, rawAPIKey := setupHandler(t)

	endpoint := createEndpoint(t, handler.DataStore)
	edgeStack := createEdgeStack(t, handler.DataStore, endpoint.ID)

	// Update edge stack status
	newStatus := portainer.EdgeStackStatusError
	payload := updateStatusPayload{
		Error:      "test-error",
		Status:     &newStatus,
		EndpointID: endpoint.ID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatal("request error:", err)
	}

	r := bytes.NewBuffer(jsonPayload)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/edge_stacks/%d/status", edgeStack.ID), r)
	if err != nil {
		t.Fatal("request error:", err)
	}

	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, endpoint.EdgeID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected a %d response, found: %d", http.StatusOK, rec.Code)
	}

	// Get updated edge stack
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/edge_stacks/%d", edgeStack.ID), nil)
	if err != nil {
		t.Fatal("request error:", err)
	}

	req.Header.Add("x-api-key", rawAPIKey)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected a %d response, found: %d", http.StatusOK, rec.Code)
	}

	updatedStack := portaineree.EdgeStack{}
	err = json.NewDecoder(rec.Body).Decode(&updatedStack)
	if err != nil {
		t.Fatal("error decoding response:", err)
	}

	endpointStatus, ok := updatedStack.Status[payload.EndpointID]
	if !ok {
		t.Fatal("Missing status")
	}

	lastStatus := endpointStatus.Status[len(endpointStatus.Status)-1]

	if len(endpointStatus.Status) == len(edgeStack.Status[payload.EndpointID].Status) {
		t.Fatal("expected status array to be updated")
	}

	if lastStatus.Type != *payload.Status {
		t.Fatalf("expected EdgeStackStatusType %d, found %d", *payload.Status, lastStatus.Type)
	}

	if endpointStatus.EndpointID != portainer.EndpointID(payload.EndpointID) {
		t.Fatalf("expected EndpointID %d, found %d", payload.EndpointID, endpointStatus.EndpointID)
	}

}
func TestUpdateStatusWithInvalidPayload(t *testing.T) {
	handler, _ := setupHandler(t)

	endpoint := createEndpoint(t, handler.DataStore)
	edgeStack := createEdgeStack(t, handler.DataStore, endpoint.ID)

	// Update edge stack status
	statusError := portainer.EdgeStackStatusError
	statusOk := portainer.EdgeStackStatusDeploymentReceived
	cases := []struct {
		Name                 string
		Payload              updateStatusPayload
		ExpectedErrorMessage string
		ExpectedStatusCode   int
	}{
		{
			"Update with no Status",
			updateStatusPayload{
				Error:      "test-error",
				Status:     nil,
				EndpointID: endpoint.ID,
			},
			"Invalid status",
			400,
		},
		{
			"Update with error status and empty error message",
			updateStatusPayload{
				Error:      "",
				Status:     &statusError,
				EndpointID: endpoint.ID,
			},
			"Error message is mandatory when status is error",
			400,
		},
		{
			"Update with missing EndpointID",
			updateStatusPayload{
				Error:      "",
				Status:     &statusOk,
				EndpointID: 0,
			},
			"Invalid EnvironmentID",
			400,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tc.Payload)
			if err != nil {
				t.Fatal("request error:", err)
			}

			r := bytes.NewBuffer(jsonPayload)
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/edge_stacks/%d/status", edgeStack.ID), r)
			if err != nil {
				t.Fatal("request error:", err)
			}

			req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, endpoint.EdgeID)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.ExpectedStatusCode {
				t.Fatalf("expected a %d response, found: %d", tc.ExpectedStatusCode, rec.Code)
			}
		})
	}
}

func TestInserRollbackEndpointStatus(t *testing.T) {
	t.Run("RolledBack status should be inserted before Running status", func(t *testing.T) {
		testStatus := []portainer.EdgeStackDeploymentStatus{
			{
				Type: portainer.EdgeStackStatusAcknowledged,
				Time: 1697163144,
			},
			{
				Type: portainer.EdgeStackStatusDeploying,
				Time: 1697163147,
			},
			{
				Type: portainer.EdgeStackStatusDeploymentReceived,
				Time: 1697163148,
			},
			{
				Type: portainer.EdgeStackStatusRunning,
				Time: 1697163150,
			},
		}

		expectedStatus := []portainer.EdgeStackDeploymentStatus{
			{
				Type: portainer.EdgeStackStatusAcknowledged,
				Time: 1697163144,
			},
			{
				Type: portainer.EdgeStackStatusDeploying,
				Time: 1697163147,
			},
			{
				Type: portainer.EdgeStackStatusDeploymentReceived,
				Time: 1697163148,
			},
			{
				Type: portainer.EdgeStackStatusRolledBack,
				Time: 1697163150,
			},
			{
				Type: portainer.EdgeStackStatusRunning,
				Time: 1697163150,
			},
		}

		actualStatus := insertRollbackEndpointStatus(testStatus)
		assert.Equal(t, expectedStatus, actualStatus, "status queue should be updated")
	})

	t.Run("RolledBack status should be inserted if there is only Running status", func(t *testing.T) {
		testStatus := []portainer.EdgeStackDeploymentStatus{
			{
				Type: portainer.EdgeStackStatusRunning,
				Time: 1697163144,
			},
		}

		expectedStatus := []portainer.EdgeStackDeploymentStatus{
			{
				Type: portainer.EdgeStackStatusRolledBack,
				Time: 1697163144,
			},
			{
				Type: portainer.EdgeStackStatusRunning,
				Time: 1697163144,
			},
		}

		actualStatus := insertRollbackEndpointStatus(testStatus)
		assert.Equal(t, expectedStatus, actualStatus, "status queue should be updated")
	})

	t.Run("No RolledBack status should be inserted if there is no Running status", func(t *testing.T) {
		testStatus := []portainer.EdgeStackDeploymentStatus{
			{
				Type: portainer.EdgeStackStatusPending,
				Time: 1697163144,
			},
		}

		expectedStatus := []portainer.EdgeStackDeploymentStatus{
			{
				Type: portainer.EdgeStackStatusPending,
				Time: 1697163144,
			},
		}

		actualStatus := insertRollbackEndpointStatus(testStatus)
		assert.Equal(t, expectedStatus, actualStatus, "status queue should not be updated")
	})
}

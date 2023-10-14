package endpointedge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	"github.com/portainer/portainer-ee/api/jwt"
	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
)

type endpointTestCase struct {
	endpoint           portaineree.Endpoint
	endpointRelation   portainer.EndpointRelation
	expectedStatusCode int
}

var endpointTestCases = []endpointTestCase{
	{
		portaineree.Endpoint{},
		portainer.EndpointRelation{},
		http.StatusForbidden,
	},
	{
		portaineree.Endpoint{
			ID:     -1,
			Name:   "endpoint-id--1",
			Type:   portaineree.EdgeAgentOnDockerEnvironment,
			URL:    "https://portainer.io:9443",
			EdgeID: "edge-id",
		},
		portainer.EndpointRelation{
			EndpointID: -1,
		},
		http.StatusForbidden,
	},
	{
		portaineree.Endpoint{
			ID:     2,
			Name:   "endpoint-id-2",
			Type:   portaineree.EdgeAgentOnDockerEnvironment,
			URL:    "https://portainer.io:9443",
			EdgeID: "",
		},
		portainer.EndpointRelation{
			EndpointID: 2,
		},
		http.StatusForbidden,
	},
	{
		portaineree.Endpoint{
			ID:     4,
			Name:   "endpoint-id-4",
			Type:   portaineree.EdgeAgentOnDockerEnvironment,
			URL:    "https://portainer.io:9443",
			EdgeID: "edge-id",
		},
		portainer.EndpointRelation{
			EndpointID: 4,
		},
		http.StatusOK,
	},
}

func mustSetupHandler(t *testing.T) *Handler {
	tmpDir := t.TempDir()
	fs, err := filesystem.NewService(tmpDir, "")
	if err != nil {
		t.Fatalf("could not start a new filesystem service: %s", err)
	}

	_, store := datastore.MustNewTestStore(t, true, true)

	ctx := context.Background()
	shutdownCtx, cancelFn := context.WithCancel(ctx)
	t.Cleanup(cancelFn)

	jwtService, err := jwt.NewService("1h", store)
	if err != nil {
		t.Fatalf("could not start a new JWT service: %s", err)
	}

	apiKeyService := apikey.NewAPIKeyService(nil, nil)

	settings, err := store.Settings().Settings()
	if err != nil {
		t.Fatalf("could not create new settings: %s", err)
	}
	settings.TrustOnFirstConnect = true

	err = store.Settings().UpdateSettings(settings)
	if err != nil {
		t.Fatalf("could not update settings: %s", err)
	}

	edgeService := edgeasync.NewService(store, fs)

	updateService, err := updateschedules.NewService(store)
	if err != nil {
		t.Fatalf("could not create update service: %s", err)
	}

	handler := NewHandler(
		security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil),
		store,
		fs,
		chisel.NewService(store, shutdownCtx, nil),
		edgeService,
		nil,
		updateService,
		nil,
	)

	handler.ReverseTunnelService = chisel.NewService(store, shutdownCtx, nil)

	return handler
}

func createEndpoint(handler *Handler, endpoint portaineree.Endpoint, endpointRelation portainer.EndpointRelation) (err error) {
	// Avoid setting ID below 0 to generate invalid test cases
	if endpoint.ID <= 0 {
		return nil
	}

	err = handler.DataStore.Endpoint().Create(&endpoint)
	if err != nil {
		return err
	}

	return handler.DataStore.EndpointRelation().Create(&endpointRelation)
}

func TestMissingEdgeIdentifier(t *testing.T) {
	handler := mustSetupHandler(t)
	endpointID := portainer.EndpointID(45)

	err := createEndpoint(handler, portaineree.Endpoint{
		ID:     endpointID,
		Name:   "endpoint-id-45",
		Type:   portaineree.EdgeAgentOnDockerEnvironment,
		URL:    "https://portainer.io:9443",
		EdgeID: "edge-id",
	}, portainer.EndpointRelation{EndpointID: endpointID})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/endpoints/%d/edge/status", endpointID), nil)
	if err != nil {
		t.Fatal("request error:", err)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(fmt.Sprintf("expected a %d response, found: %d without Edge identifier", http.StatusForbidden, rec.Code))
	}
}

func TestWithEndpoints(t *testing.T) {
	handler := mustSetupHandler(t)

	for _, test := range endpointTestCases {
		err := createEndpoint(handler, test.endpoint, test.endpointRelation)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/endpoints/%d/edge/status", test.endpoint.ID), nil)
		if err != nil {
			t.Fatal("request error:", err)
		}
		req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, test.endpoint.EdgeID)
		req.Header.Set(portaineree.HTTPResponseAgentPlatform, "1")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != test.expectedStatusCode {
			t.Fatalf(fmt.Sprintf("expected a %d response, found: %d for endpoint ID: %d", test.expectedStatusCode, rec.Code, test.endpoint.ID))
		}
	}
}

func TestLastCheckInDateIncreases(t *testing.T) {
	handler := mustSetupHandler(t)

	endpointID := portainer.EndpointID(56)
	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-56",
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	endpointRelation := portainer.EndpointRelation{
		EndpointID: endpoint.ID,
	}

	err := createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/endpoints/%d/edge/status", endpoint.ID), nil)
	if err != nil {
		t.Fatal("request error:", err)
	}
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, "edge-id")
	req.Header.Set(portaineree.HTTPResponseAgentPlatform, "1")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmt.Sprintf("expected a %d response, found: %d", http.StatusOK, rec.Code))
	}

	updatedEndpoint, err := handler.DataStore.Endpoint().Endpoint(endpoint.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, updatedEndpoint.LastCheckInDate, endpoint.LastCheckInDate)
}

func TestEmptyEdgeIdWithAgentPlatformHeader(t *testing.T) {
	handler := mustSetupHandler(t)

	endpointID := portainer.EndpointID(44)
	edgeId := "edge-id"
	endpoint := portaineree.Endpoint{
		ID:     endpointID,
		Name:   "test-endpoint-44",
		Type:   portaineree.EdgeAgentOnDockerEnvironment,
		URL:    "https://portainer.io:9443",
		EdgeID: "",
	}
	endpointRelation := portainer.EndpointRelation{
		EndpointID: endpoint.ID,
	}

	err := createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/endpoints/%d/edge/status", endpoint.ID), nil)
	if err != nil {
		t.Fatal("request error:", err)
	}
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeId)
	req.Header.Set(portaineree.HTTPResponseAgentPlatform, "1")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmt.Sprintf("expected a %d response, found: %d with empty edge ID", http.StatusOK, rec.Code))
	}

	updatedEndpoint, err := handler.DataStore.Endpoint().Endpoint(endpoint.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, updatedEndpoint.EdgeID, edgeId)
}

func TestEdgeStackStatus(t *testing.T) {
	handler := mustSetupHandler(t)

	endpointID := portainer.EndpointID(7)
	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-7",
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	edgeStackID := portainer.EdgeStackID(17)
	edgeStack := portaineree.EdgeStack{
		ID:   edgeStackID,
		Name: "test-edge-stack-17",
		Status: map[portainer.EndpointID]portainer.EdgeStackStatus{
			endpointID: {
				Status:         []portainer.EdgeStackDeploymentStatus{},
				EndpointID:     portainer.EndpointID(endpointID),
				DeploymentInfo: portainer.StackDeploymentInfo{},
			},
		},
		CreationDate:     time.Now().Unix(),
		EdgeGroups:       []portainer.EdgeGroupID{1, 2},
		ProjectPath:      "/project/path",
		EntryPoint:       "entrypoint",
		Version:          237,
		StackFileVersion: 238,
		ManifestPath:     "/manifest/path",
		DeploymentType:   1,
	}

	endpointRelation := portainer.EndpointRelation{
		EndpointID: endpoint.ID,
		EdgeStacks: map[portainer.EdgeStackID]bool{
			edgeStack.ID: true,
		},
	}
	handler.DataStore.EdgeStack().Create(edgeStack.ID, &edgeStack)

	err := createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/endpoints/%d/edge/status", endpoint.ID), nil)
	if err != nil {
		t.Fatal("request error:", err)
	}
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, "edge-id")
	req.Header.Set(portaineree.HTTPResponseAgentPlatform, "1")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmt.Sprintf("expected a %d response, found: %d", http.StatusOK, rec.Code))
	}

	var data endpointEdgeStatusInspectResponse
	err = json.NewDecoder(rec.Body).Decode(&data)
	if err != nil {
		t.Fatal("error decoding response:", err)
	}

	assert.Len(t, data.Stacks, 1)
	assert.Equal(t, edgeStack.ID, data.Stacks[0].ID)
	assert.Equal(t, edgeStack.Version, data.Stacks[0].Version)
}

func TestEdgeJobsResponse(t *testing.T) {
	handler := mustSetupHandler(t)

	endpointID := portainer.EndpointID(77)
	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-77",
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	endpointRelation := portainer.EndpointRelation{
		EndpointID: endpoint.ID,
	}

	err := createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	path, err := handler.FileService.StoreEdgeJobFileFromBytes("test-script", []byte("pwd"))
	if err != nil {
		t.Fatal(err)
	}

	edgeJobID := portainer.EdgeJobID(35)
	edgeJob := portainer.EdgeJob{
		ID:             edgeJobID,
		Created:        time.Now().Unix(),
		CronExpression: "* * * * *",
		Name:           "test-edge-job",
		ScriptPath:     path,
		Recurring:      true,
		Version:        57,
	}

	handler.ReverseTunnelService.AddEdgeJob(&endpoint, &edgeJob)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/endpoints/%d/edge/status", endpoint.ID), nil)
	if err != nil {
		t.Fatal("request error:", err)
	}
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, "edge-id")
	req.Header.Set(portaineree.HTTPResponseAgentPlatform, "1")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmt.Sprintf("expected a %d response, found: %d", http.StatusOK, rec.Code))
	}

	var data endpointEdgeStatusInspectResponse
	err = json.NewDecoder(rec.Body).Decode(&data)
	if err != nil {
		t.Fatal("error decoding response:", err)
	}

	assert.Len(t, data.Schedules, 1)
	assert.Equal(t, edgeJob.ID, data.Schedules[0].ID)
	assert.Equal(t, edgeJob.CronExpression, data.Schedules[0].CronExpression)
	assert.Equal(t, edgeJob.Version, data.Schedules[0].Version)
}

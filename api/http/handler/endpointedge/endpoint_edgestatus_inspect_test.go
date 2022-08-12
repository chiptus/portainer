package endpointedge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/stretchr/testify/assert"
)

type endpointTestCase struct {
	endpoint           portaineree.Endpoint
	endpointRelation   portaineree.EndpointRelation
	expectedStatusCode int
}

var endpointTestCases = []endpointTestCase{
	{
		portaineree.Endpoint{},
		portaineree.EndpointRelation{},
		http.StatusNotFound,
	},
	{
		portaineree.Endpoint{
			ID:     -1,
			Name:   "endpoint-id--1",
			Type:   portaineree.EdgeAgentOnDockerEnvironment,
			URL:    "https://portainer.io:9443",
			EdgeID: "edge-id",
		},
		portaineree.EndpointRelation{
			EndpointID: -1,
		},
		http.StatusNotFound,
	},
	{
		portaineree.Endpoint{
			ID:     2,
			Name:   "endpoint-id-2",
			Type:   portaineree.EdgeAgentOnDockerEnvironment,
			URL:    "https://portainer.io:9443",
			EdgeID: "",
		},
		portaineree.EndpointRelation{
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
		portaineree.EndpointRelation{
			EndpointID: 4,
		},
		http.StatusOK,
	},
}

func setupHandler() (*Handler, func(), error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "portainer-test")
	if err != nil {
		return nil, nil, fmt.Errorf("could not create a tmp dir: %w", err)
	}

	fs, err := filesystem.NewService(tmpDir, "")
	if err != nil {
		return nil, nil, fmt.Errorf("could not start a new filesystem service: %w", err)
	}

	_, store, storeTeardown := datastore.MustNewTestStore(true, true)

	ctx := context.Background()
	shutdownCtx, cancelFn := context.WithCancel(ctx)

	teardown := func() {
		cancelFn()
		storeTeardown()
	}

	jwtService, err := jwt.NewService("1h", store)
	if err != nil {
		teardown()
		return nil, nil, fmt.Errorf("could not start a new jwt service: %w", err)
	}

	apiKeyService := apikey.NewAPIKeyService(nil, nil)

	settings, err := store.Settings().Settings()
	if err != nil {
		teardown()
		return nil, nil, fmt.Errorf("could not create new settings: %w", err)
	}
	settings.TrustOnFirstConnect = true

	err = store.Settings().UpdateSettings(settings)
	if err != nil {
		teardown()
		return nil, nil, fmt.Errorf("could not update settings: %w", err)
	}

	edgeService := edge.NewService(store, fs)

	handler := NewHandler(
		security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil),
		store,
		fs,
		chisel.NewService(store, shutdownCtx),
		*edgeService,
		nil,
	)

	handler.ReverseTunnelService = chisel.NewService(store, shutdownCtx)

	return handler, teardown, nil
}

func createEndpoint(handler *Handler, endpoint portaineree.Endpoint, endpointRelation portaineree.EndpointRelation) (err error) {
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
	handler, teardown, err := setupHandler()
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	endpointID := portaineree.EndpointID(45)
	err = createEndpoint(handler, portaineree.Endpoint{
		ID:     endpointID,
		Name:   "endpoint-id-45",
		Type:   portaineree.EdgeAgentOnDockerEnvironment,
		URL:    "https://portainer.io:9443",
		EdgeID: "edge-id",
	}, portaineree.EndpointRelation{EndpointID: endpointID})

	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d/edge/status", endpointID), nil)
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
	handler, teardown, err := setupHandler()
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	for _, test := range endpointTestCases {
		err = createEndpoint(handler, test.endpoint, test.endpointRelation)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d/edge/status", test.endpoint.ID), nil)
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
	handler, teardown, err := setupHandler()
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	endpointID := portaineree.EndpointID(56)
	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-56",
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	endpointRelation := portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
	}

	err = createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d/edge/status", endpoint.ID), nil)
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
	handler, teardown, err := setupHandler()
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	endpointID := portaineree.EndpointID(44)
	edgeId := "edge-id"
	endpoint := portaineree.Endpoint{
		ID:     endpointID,
		Name:   "test-endpoint-44",
		Type:   portaineree.EdgeAgentOnDockerEnvironment,
		URL:    "https://portainer.io:9443",
		EdgeID: "",
	}
	endpointRelation := portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
	}

	err = createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d/edge/status", endpoint.ID), nil)
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
	handler, teardown, err := setupHandler()
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	endpointID := portaineree.EndpointID(7)
	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-7",
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	edgeStackID := portaineree.EdgeStackID(17)
	edgeStack := portaineree.EdgeStack{
		ID:   edgeStackID,
		Name: "test-edge-stack-17",
		Status: map[portaineree.EndpointID]portaineree.EdgeStackStatus{
			endpointID: {Type: portaineree.StatusOk, Error: "", EndpointID: endpoint.ID},
		},
		CreationDate:   time.Now().Unix(),
		EdgeGroups:     []portaineree.EdgeGroupID{1, 2},
		ProjectPath:    "/project/path",
		EntryPoint:     "entrypoint",
		Version:        237,
		ManifestPath:   "/manifest/path",
		DeploymentType: 1,
	}

	endpointRelation := portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
		EdgeStacks: map[portaineree.EdgeStackID]bool{
			edgeStack.ID: true,
		},
	}
	handler.DataStore.EdgeStack().Create(edgeStack.ID, &edgeStack)

	err = createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d/edge/status", endpoint.ID), nil)
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
	handler, teardown, err := setupHandler()
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	endpointID := portaineree.EndpointID(77)
	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-77",
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	endpointRelation := portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
	}

	err = createEndpoint(handler, endpoint, endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	path, err := handler.FileService.StoreEdgeJobFileFromBytes("test-script", []byte("pwd"))
	if err != nil {
		t.Fatal(err)
	}

	edgeJobID := portaineree.EdgeJobID(35)
	edgeJob := portaineree.EdgeJob{
		ID:             edgeJobID,
		Created:        time.Now().Unix(),
		CronExpression: "* * * * *",
		Name:           "test-edge-job",
		ScriptPath:     path,
		Recurring:      true,
		Version:        57,
	}

	handler.ReverseTunnelService.AddEdgeJob(endpoint.ID, &edgeJob)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d/edge/status", endpoint.ID), nil)
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

package edgeconfigs

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/http/handler/edgegroups"
	"github.com/portainer/portainer-ee/api/http/handler/endpointgroups"
	"github.com/portainer/portainer-ee/api/http/handler/endpoints"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer/api/dataservices/errors"

	"github.com/stretchr/testify/require"
)

func generateEdgeConfigFile() ([]byte, error) {
	buf := &bytes.Buffer{}
	writer := zip.NewWriter(buf)

	cfg, err := writer.Create("config-file")
	if err != nil {
		return nil, err
	}

	_, err = cfg.Write([]byte("test-config-value"))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func createEdgeGroup(t *testing.T, store *datastore.Store, id int, endpoints []portaineree.EndpointID) {
	t.Helper()

	err := store.EdgeGroup().Create(&portaineree.EdgeGroup{
		ID:        portaineree.EdgeGroupID(id),
		Name:      fmt.Sprintf("edge-group-%d", id),
		Endpoints: endpoints,
	})
	require.NoError(t, err)
}

func requireState(t *testing.T, expected, actual portaineree.EdgeConfigStateType) {
	t.Helper()

	require.Equal(t, expected.String(), actual.String())
}

func requireProgress(t *testing.T, config *portaineree.EdgeConfig, success, total int) {
	t.Helper()

	require.Equal(t, portaineree.EdgeConfigProgress{
		Success: success,
		Total:   total,
	}, config.Progress)
}

func TestStdFlow(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	edgeAsyncService := edgeasync.NewService(store, fileService)

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)

	usr := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(usr)
	require.NoError(t, err)

	token, err := jwtService.GenerateToken(&portaineree.TokenData{ID: usr.ID, Username: usr.Username, Role: portaineree.AdministratorRole})
	require.NoError(t, err)

	settings, err := store.Settings().Settings()
	require.NoError(t, err)

	settings.EnableEdgeComputeFeatures = true
	err = store.Settings().UpdateSettings(settings)
	require.NoError(t, err)

	configID := portaineree.EdgeConfigID(1)

	endpointID := portaineree.EndpointID(1)
	edgeID := "edge-id-1"
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID:          endpointID,
		Name:        "endpoint-1",
		EdgeID:      edgeID,
		GroupID:     1,
		Type:        portaineree.EdgeAgentOnDockerEnvironment,
		UserTrusted: true,
	})
	require.NoError(t, err)

	endpointIDtoRemove := portaineree.EndpointID(2)
	edgeIDtoRemove := "edge-id-2"
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID:          endpointIDtoRemove,
		Name:        "endpoint-2",
		EdgeID:      edgeIDtoRemove,
		GroupID:     1,
		Type:        portaineree.EdgeAgentOnDockerEnvironment,
		UserTrusted: true,
	})
	require.NoError(t, err)

	endpointIDtoAddRm := portaineree.EndpointID(3)
	edgeIDtoAddRm := "edge-id-3"
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID:          endpointIDtoAddRm,
		Name:        "endpoint-3",
		EdgeID:      edgeIDtoAddRm,
		GroupID:     1,
		Type:        portaineree.EdgeAgentOnDockerEnvironment,
		UserTrusted: true,
	})
	require.NoError(t, err)

	err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
		EndpointID: endpointID,
		EdgeStacks: make(map[portaineree.EdgeStackID]bool),
	})
	require.NoError(t, err)

	err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
		EndpointID: endpointIDtoRemove,
		EdgeStacks: make(map[portaineree.EdgeStackID]bool),
	})
	require.NoError(t, err)

	err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
		EndpointID: endpointIDtoAddRm,
		EdgeStacks: make(map[portaineree.EdgeStackID]bool),
	})
	require.NoError(t, err)

	err = store.EndpointGroup().Create(&portaineree.EndpointGroup{
		ID:   1,
		Name: "endpoint-group-1",
	})
	require.NoError(t, err)

	createEdgeGroup(t, store, 1, []portaineree.EndpointID{endpointID})
	createEdgeGroup(t, store, 2, []portaineree.EndpointID{})
	createEdgeGroup(t, store, 3, []portaineree.EndpointID{endpointIDtoAddRm})

	bouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, nil, nil)

	edgeGroupsHandler := edgegroups.NewHandler(bouncer, testhelpers.NewUserActivityService(), edgeAsyncService)
	edgeGroupsHandler.DataStore = store

	type edgeGroupUpdatePayload struct {
		Name         string
		Dynamic      bool
		TagIDs       []portaineree.TagID
		Endpoints    []portaineree.EndpointID
		PartialMatch *bool
	}

	edgeGroupPayload := &bytes.Buffer{}

	cache.Set(endpointID, []byte("fake-cache"))
	cache.Set(endpointIDtoRemove, []byte("fake-cache"))

	_, ok := cache.Get(endpointID)
	require.True(t, ok)

	_, ok = cache.Get(endpointIDtoRemove)
	require.True(t, ok)

	h := NewHandler(store, bouncer, testhelpers.NewUserActivityService(), edgeAsyncService, fileService)

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	configPart, err := writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Name:         "edge-config-1",
		BaseDir:      "/tmp",
		Type:         "general",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1, 2},
	})
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err := generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/edge_configurations", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err := store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigSavingState, config.State)
	requireProgress(t, config, 0, 1)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/edge_configurations/%d", configID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)
	err = json.Unmarshal(rr.Body.Bytes(), config)
	require.NoError(t, err)
	require.Equal(t, configID, config.ID)
	requireState(t, portaineree.EdgeConfigSavingState, config.State)
	requireProgress(t, config, 0, 1)

	configState, err := store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigSavingState, configState.States[config.ID])

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/edge_configurations/%d/files", configID), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	dirEntries, err := fileService.GetEdgeConfigDirEntries(config, "", portaineree.EdgeConfigCurrent)
	require.NoError(t, err)
	require.Equal(t, 1, len(dirEntries))
	require.Equal(t, "/config-file", dirEntries[0].Name)

	configFiles := &edgeConfigFilesPayload{}
	err = json.Unmarshal(rr.Body.Bytes(), configFiles)
	require.NoError(t, err)
	require.Equal(t, configID, configFiles.ID)
	require.Equal(t, "edge-config-1", configFiles.Name)
	require.Equal(t, "/tmp", configFiles.BaseDir)
	require.Equal(t, 1, len(configFiles.DirEntries))
	require.Equal(t, "/config-file", configFiles.DirEntries[0].Name)

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	// Update the edge group to include the second endpoint

	err = json.NewEncoder(edgeGroupPayload).Encode(edgeGroupUpdatePayload{
		Name:      "edge-group-2",
		TagIDs:    []portaineree.TagID{},
		Endpoints: []portaineree.EndpointID{endpointIDtoRemove},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/edge_groups/2", edgeGroupPayload)
	req.Header.Set("Authorization", "Bearer "+token)

	edgeGroupsHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 2)

	configState, err = store.EdgeConfigState().Read(endpointIDtoRemove)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoRemove, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigSavingState, configState.States[config.ID])

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeIDtoRemove)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 2, 2)

	configState, err = store.EdgeConfigState().Read(endpointIDtoRemove)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoRemove, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	_, ok = cache.Get(endpointIDtoRemove)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))
	cache.Set(endpointIDtoRemove, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	_, ok = cache.Get(endpointIDtoRemove)
	require.True(t, ok)

	// Update the edge group to remove the second endpoint

	err = json.NewEncoder(edgeGroupPayload).Encode(edgeGroupUpdatePayload{
		Name:      "edge-group-2",
		TagIDs:    []portaineree.TagID{},
		Endpoints: []portaineree.EndpointID{},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/edge_groups/2", edgeGroupPayload)
	req.Header.Set("Authorization", "Bearer "+token)

	edgeGroupsHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 2, 2)

	configState, err = store.EdgeConfigState().Read(endpointIDtoRemove)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoRemove, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigDeletingState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	_, ok = cache.Get(endpointIDtoRemove)
	require.False(t, ok)

	cache.Set(endpointIDtoRemove, []byte("fake-cache"))

	_, ok = cache.Get(endpointIDtoRemove)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeIDtoRemove)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointIDtoRemove)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoRemove, configState.EndpointID)

	_, ok = configState.States[configID]
	require.False(t, ok)

	// Update the edge config to add the third endpoint

	writer = multipart.NewWriter(body)

	configPart, err = writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Type:         "foldername",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1, 2, 3},
	})
	require.NoError(t, err)

	filePart, err = writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err = generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/edge_configurations/1", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigUpdatingState, config.State)
	requireProgress(t, config, 0, 2)

	// Endpoint 1
	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigUpdatingState, configState.States[config.ID])

	// Endpoint 3
	configState, err = store.EdgeConfigState().Read(endpointIDtoAddRm)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoAddRm, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigSavingState, configState.States[config.ID])

	dirEntries, err = fileService.GetEdgeConfigDirEntries(config, "", portaineree.EdgeConfigCurrent)
	require.NoError(t, err)
	require.Equal(t, 1, len(dirEntries))
	require.Equal(t, "config-file", dirEntries[0].Name)

	// Simulate failure for endpoint 1
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigFailureState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigFailureState, config.State)
	requireProgress(t, config, 0, 2)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigFailureState, configState.States[config.ID])

	// Simulate progress for endpoint 3
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeIDtoAddRm)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigFailureState, config.State)
	requireProgress(t, config, 1, 2)

	configState, err = store.EdgeConfigState().Read(endpointIDtoAddRm)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoAddRm, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	_, ok = cache.Get(endpointIDtoRemove)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Update the edge config to remove the third endpoint

	writer = multipart.NewWriter(body)

	configPart, err = writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Type:         "foldername",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1, 2},
	})
	require.NoError(t, err)

	filePart, err = writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err = generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/edge_configurations/1", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigUpdatingState, config.State)
	requireProgress(t, config, 0, 2)

	// Endpoint 1
	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigUpdatingState, configState.States[config.ID])

	// Endpoint 3
	configState, err = store.EdgeConfigState().Read(endpointIDtoAddRm)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoAddRm, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigDeletingState, configState.States[config.ID])

	// Simulate progress for endpoint 1
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigUpdatingState, config.State)
	requireProgress(t, config, 1, 2)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	// Simulate progress for endpoint 3
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeIDtoAddRm)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointIDtoAddRm)
	require.NoError(t, err)
	require.Equal(t, endpointIDtoAddRm, configState.EndpointID)

	_, ok = configState.States[config.ID]
	require.False(t, ok)

	// Delete the edge config

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/edge_configurations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigDeletingState, config.State)
	requireProgress(t, config, 0, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, configState.EndpointID, endpointID)
	require.Equal(t, configState.States[config.ID], portaineree.EdgeConfigDeletingState)

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	_, err = store.EdgeConfig().Read(configID)
	require.ErrorIs(t, err, errors.ErrObjectNotFound)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)

	_, ok = configState.States[configID]
	require.False(t, ok)
}

func TestEnvTagsAddRm(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	edgeAsyncService := edgeasync.NewService(store, fileService)

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)

	usr := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(usr)
	require.NoError(t, err)

	token, err := jwtService.GenerateToken(&portaineree.TokenData{ID: usr.ID, Username: usr.Username, Role: portaineree.AdministratorRole})
	require.NoError(t, err)

	settings, err := store.Settings().Settings()
	require.NoError(t, err)

	settings.EnableEdgeComputeFeatures = true
	err = store.Settings().UpdateSettings(settings)
	require.NoError(t, err)

	configID := portaineree.EdgeConfigID(1)

	endpointID := portaineree.EndpointID(1)
	edgeID := "edge-id-1"
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID:          endpointID,
		Name:        "endpoint-1",
		EdgeID:      edgeID,
		GroupID:     1,
		Type:        portaineree.EdgeAgentOnDockerEnvironment,
		UserTrusted: true,
	})
	require.NoError(t, err)

	err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
		EndpointID: endpointID,
		EdgeStacks: make(map[portaineree.EdgeStackID]bool),
	})
	require.NoError(t, err)

	err = store.EndpointGroup().Create(&portaineree.EndpointGroup{
		ID:   1,
		Name: "endpoint-group-1",
	})
	require.NoError(t, err)

	err = store.Tag().Create(&portaineree.Tag{
		ID:        1,
		Name:      "tag-1",
		Endpoints: make(map[portaineree.EndpointID]bool),
	})
	require.NoError(t, err)

	err = store.Tag().Create(&portaineree.Tag{
		ID:        2,
		Name:      "unrelated-tag",
		Endpoints: make(map[portaineree.EndpointID]bool),
	})
	require.NoError(t, err)

	err = store.EdgeGroup().Create(&portaineree.EdgeGroup{
		ID:      1,
		Name:    "edge-group-1",
		Dynamic: true,
		TagIDs:  []portaineree.TagID{1},
	})
	require.NoError(t, err)

	bouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, nil, nil)

	endpointHandler := endpoints.NewHandler(bouncer, testhelpers.NewUserActivityService(), store, edgeAsyncService, nil, nil, testhelpers.Licenseservice{})
	endpointHandler.FileService = fileService
	endpointHandler.SnapshotService, err = snapshot.NewService("1h", store, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	h := NewHandler(store, bouncer, testhelpers.NewUserActivityService(), edgeAsyncService, fileService)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok := cache.Get(endpointID)
	require.True(t, ok)

	// Create Edge Config
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	configPart, err := writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Name:         "test",
		BaseDir:      "/tmp",
		Type:         "foldername",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1},
	})
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err := generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/edge_configurations", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err := store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 0)

	// Add the tag to the environment

	type endpointUpdatePayload struct {
		TagIDs []portaineree.TagID
	}

	body.Reset()
	err = json.NewEncoder(body).Encode(endpointUpdatePayload{
		TagIDs: []portaineree.TagID{1},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/endpoints/1", body)
	req.Header.Set("Authorization", "Bearer "+token)

	endpointHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	endpoint, err := store.Endpoint().Endpoint(endpointID)
	require.NoError(t, err)
	require.ElementsMatch(t, []portaineree.TagID{1}, endpoint.TagIDs)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 1)

	configState, err := store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigSavingState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	// Add the unrelated tag to the environment

	body.Reset()
	err = json.NewEncoder(body).Encode(endpointUpdatePayload{
		TagIDs: []portaineree.TagID{1, 2},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/endpoints/1", body)
	req.Header.Set("Authorization", "Bearer "+token)

	endpointHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	endpoint, err = store.Endpoint().Endpoint(endpointID)
	require.NoError(t, err)
	require.ElementsMatch(t, []portaineree.TagID{1, 2}, endpoint.TagIDs)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Remove the tag from the environment

	body.Reset()
	err = json.NewEncoder(body).Encode(endpointUpdatePayload{
		TagIDs: []portaineree.TagID{},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/endpoints/1", body)
	req.Header.Set("Authorization", "Bearer "+token)

	endpointHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	endpoint, err = store.Endpoint().Endpoint(endpointID)
	require.NoError(t, err)
	require.ElementsMatch(t, []portaineree.TagID{}, endpoint.TagIDs)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigDeletingState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 0)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)

	_, ok = configState.States[config.ID]
	require.False(t, ok)
}

func TestEnvGroups(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	edgeAsyncService := edgeasync.NewService(store, fileService)

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)

	usr := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(usr)
	require.NoError(t, err)

	token, err := jwtService.GenerateToken(&portaineree.TokenData{ID: usr.ID, Username: usr.Username, Role: portaineree.AdministratorRole})
	require.NoError(t, err)

	settings, err := store.Settings().Settings()
	require.NoError(t, err)

	settings.EnableEdgeComputeFeatures = true
	err = store.Settings().UpdateSettings(settings)
	require.NoError(t, err)

	configID := portaineree.EdgeConfigID(1)

	endpointID := portaineree.EndpointID(1)
	edgeID := "edge-id-1"
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID:          endpointID,
		Name:        "endpoint-1",
		EdgeID:      edgeID,
		GroupID:     1,
		Type:        portaineree.EdgeAgentOnDockerEnvironment,
		UserTrusted: true,
	})
	require.NoError(t, err)

	err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
		EndpointID: endpointID,
		EdgeStacks: make(map[portaineree.EdgeStackID]bool),
	})
	require.NoError(t, err)

	err = store.EndpointGroup().Create(&portaineree.EndpointGroup{
		ID:   1,
		Name: "endpoint-group-1",
	})
	require.NoError(t, err)

	err = store.Tag().Create(&portaineree.Tag{
		ID:             1,
		Name:           "tag-1",
		Endpoints:      make(map[portaineree.EndpointID]bool),
		EndpointGroups: map[portaineree.EndpointGroupID]bool{},
	})
	require.NoError(t, err)

	err = store.EdgeGroup().Create(&portaineree.EdgeGroup{
		ID:      1,
		Name:    "edge-group-1",
		Dynamic: true,
		TagIDs:  []portaineree.TagID{1},
	})
	require.NoError(t, err)

	bouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, nil, nil)

	groupHandler := endpointgroups.NewHandler(bouncer, testhelpers.NewUserActivityService(), edgeAsyncService)
	groupHandler.DataStore = store
	require.NoError(t, err)

	h := NewHandler(store, bouncer, testhelpers.NewUserActivityService(), edgeAsyncService, fileService)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok := cache.Get(endpointID)
	require.True(t, ok)

	// Create Edge Config
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	configPart, err := writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Name:         "test",
		BaseDir:      "/tmp",
		Type:         "foldername",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1},
	})
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err := generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/edge_configurations", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err := store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 0)

	// Add the tag to the group

	type endpointGroupUpdatePayload struct {
		TagIDs []portaineree.TagID
	}

	body.Reset()
	err = json.NewEncoder(body).Encode(endpointGroupUpdatePayload{
		TagIDs: []portaineree.TagID{1},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/endpoint_groups/1", body)
	req.Header.Set("Authorization", "Bearer "+token)

	groupHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	group, err := store.EndpointGroup().Read(1)
	require.NoError(t, err)
	require.ElementsMatch(t, []portaineree.TagID{1}, group.TagIDs)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 1)

	configState, err := store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigSavingState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigIdleState, configState.States[config.ID])

	// Remove the tag from the group

	body.Reset()
	err = json.NewEncoder(body).Encode(endpointGroupUpdatePayload{
		TagIDs: []portaineree.TagID{},
	})
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/endpoint_groups/1", body)
	req.Header.Set("Authorization", "Bearer "+token)

	groupHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	endpoint, err := store.Endpoint().Endpoint(endpointID)
	require.NoError(t, err)
	require.ElementsMatch(t, []portaineree.TagID{}, endpoint.TagIDs)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 1, 1)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)
	requireState(t, portaineree.EdgeConfigDeletingState, configState.States[config.ID])

	_, ok = cache.Get(endpointID)
	require.False(t, ok)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok = cache.Get(endpointID)
	require.True(t, ok)

	// Simulate progress

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/edge_configurations/1/%d", portaineree.EdgeConfigIdleState), nil)
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 0)

	configState, err = store.EdgeConfigState().Read(endpointID)
	require.NoError(t, err)
	require.Equal(t, endpointID, configState.EndpointID)

	_, ok = configState.States[config.ID]
	require.False(t, ok)
}

func TestDeleteEmptyConfig(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	edgeAsyncService := edgeasync.NewService(store, fileService)

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)

	usr := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(usr)
	require.NoError(t, err)

	token, err := jwtService.GenerateToken(&portaineree.TokenData{ID: usr.ID, Username: usr.Username, Role: portaineree.AdministratorRole})
	require.NoError(t, err)

	settings, err := store.Settings().Settings()
	require.NoError(t, err)

	settings.EnableEdgeComputeFeatures = true
	err = store.Settings().UpdateSettings(settings)
	require.NoError(t, err)

	configID := portaineree.EdgeConfigID(1)

	err = store.EndpointGroup().Create(&portaineree.EndpointGroup{
		ID:   1,
		Name: "endpoint-group-1",
	})
	require.NoError(t, err)

	createEdgeGroup(t, store, 1, []portaineree.EndpointID{})

	bouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, nil, nil)

	h := NewHandler(store, bouncer, testhelpers.NewUserActivityService(), edgeAsyncService, fileService)

	// Create Edge Config
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	configPart, err := writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Name:         "test",
		BaseDir:      "/tmp",
		Type:         "foldername",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1},
	})
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err := generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/edge_configurations", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err := store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 0)

	// Delete the edge config

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/edge_configurations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	_, err = store.EdgeConfig().Read(configID)
	require.ErrorIs(t, err, errors.ErrObjectNotFound)
}

func TestEndpointDelete(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	edgeAsyncService := edgeasync.NewService(store, fileService)

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)

	usr := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(usr)
	require.NoError(t, err)

	token, err := jwtService.GenerateToken(&portaineree.TokenData{ID: usr.ID, Username: usr.Username, Role: portaineree.AdministratorRole})
	require.NoError(t, err)

	settings, err := store.Settings().Settings()
	require.NoError(t, err)

	settings.EnableEdgeComputeFeatures = true
	err = store.Settings().UpdateSettings(settings)
	require.NoError(t, err)

	configID := portaineree.EdgeConfigID(1)

	endpointID := portaineree.EndpointID(1)
	edgeID := "edge-id-1"
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID:          endpointID,
		Name:        "endpoint-1",
		EdgeID:      edgeID,
		GroupID:     1,
		Type:        portaineree.EdgeAgentOnDockerEnvironment,
		UserTrusted: true,
	})
	require.NoError(t, err)

	err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
		EndpointID: endpointID,
		EdgeStacks: make(map[portaineree.EdgeStackID]bool),
	})
	require.NoError(t, err)

	err = store.EndpointGroup().Create(&portaineree.EndpointGroup{
		ID:   1,
		Name: "endpoint-group-1",
	})
	require.NoError(t, err)

	createEdgeGroup(t, store, 1, []portaineree.EndpointID{endpointID})

	bouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, nil, nil)

	endpointHandler := endpoints.NewHandler(bouncer, testhelpers.NewUserActivityService(), store, edgeAsyncService, demo.NewService(), nil, testhelpers.Licenseservice{})
	endpointHandler.FileService = fileService
	endpointHandler.ProxyManager = proxy.NewManager(store, nil, nil, nil, nil, nil, nil, nil, nil)
	endpointHandler.AuthorizationService = authorization.NewService(store)
	endpointHandler.SnapshotService, err = snapshot.NewService("1h", store, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	h := NewHandler(store, bouncer, testhelpers.NewUserActivityService(), edgeAsyncService, fileService)

	cache.Set(endpointID, []byte("fake-cache"))

	_, ok := cache.Get(endpointID)
	require.True(t, ok)

	// Create Edge Config
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	configPart, err := writer.CreateFormField("edgeConfiguration")
	require.NoError(t, err)

	err = json.NewEncoder(configPart).Encode(edgeConfigCreatePayload{
		Name:         "test",
		BaseDir:      "/tmp",
		Type:         "foldername",
		Category:     "configuration",
		EdgeGroupIDs: []portaineree.EdgeGroupID{1},
	})
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "test.zip")
	require.NoError(t, err)

	content, err := generateEdgeConfigFile()
	require.NoError(t, err)

	_, err = filePart.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/edge_configurations", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	config, err := store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigSavingState, config.State)
	requireProgress(t, config, 0, 1)

	// Remove the environment

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/endpoints/1", body)
	req.Header.Set("Authorization", "Bearer "+token)

	endpointHandler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	_, err = store.Endpoint().Endpoint(endpointID)
	require.ErrorIs(t, err, errors.ErrObjectNotFound)

	config, err = store.EdgeConfig().Read(configID)
	require.NoError(t, err)
	requireState(t, portaineree.EdgeConfigIdleState, config.State)
	requireProgress(t, config, 0, 0)

	_, err = store.EdgeConfigState().Read(endpointID)
	require.ErrorIs(t, err, errors.ErrObjectNotFound)

	_, ok = cache.Get(endpointID)
	require.False(t, ok)
}

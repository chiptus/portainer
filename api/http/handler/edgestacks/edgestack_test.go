package edgestacks

import (
	"os"
	"strconv"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/pkg/errors"
)

// Helpers
func setupHandler(t *testing.T) (*Handler, string, func()) {
	t.Helper()

	_, store, storeTeardown := datastore.MustNewTestStore(t, true, true)

	jwtService, err := jwt.NewService("1h", store)
	if err != nil {
		storeTeardown()
		t.Fatal(err)
	}

	user := &portaineree.User{ID: 2, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(user)
	if err != nil {
		storeTeardown()
		t.Fatal(err)
	}

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	rawAPIKey, _, err := apiKeyService.GenerateApiKey(*user, "test")
	if err != nil {
		storeTeardown()
		t.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp(t.TempDir(), "portainer-test")
	if err != nil {
		storeTeardown()
		t.Fatal(err)
	}

	fs, err := filesystem.NewService(tmpDir, "")
	if err != nil {
		storeTeardown()
		t.Fatal(err)
	}

	edgeAsyncService := edgeasync.NewService(store, fs)

	edgeUpdateService, err := updateschedules.NewService(store)
	if err != nil {
		storeTeardown()
		t.Fatal(err)
	}

	handler := NewHandler(
		security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil),
		helper.NewUserActivityService(),
		store,
		edgeAsyncService,
		edgestacks.NewService(store, edgeAsyncService),
		edgeUpdateService,
	)

	handler.FileService = fs

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		t.Fatal(err)
	}
	settings.EnableEdgeComputeFeatures = true

	err = handler.DataStore.Settings().UpdateSettings(settings)
	if err != nil {
		t.Fatal(err)
	}

	handler.GitService = testhelpers.NewGitService(errors.New("Clone error"), "git-service-id")

	return handler, rawAPIKey, storeTeardown
}

func createEndpointWithId(t *testing.T, store dataservices.DataStore, endpointID portaineree.EndpointID) portaineree.Endpoint {
	t.Helper()

	endpoint := portaineree.Endpoint{
		ID:              endpointID,
		Name:            "test-endpoint-" + strconv.Itoa(int(endpointID)),
		Type:            portaineree.EdgeAgentOnDockerEnvironment,
		URL:             "https://portainer.io:9443",
		EdgeID:          "edge-id",
		LastCheckInDate: time.Now().Unix(),
	}

	err := store.Endpoint().Create(&endpoint)
	if err != nil {
		t.Fatal(err)
	}

	return endpoint
}

func createEndpoint(t *testing.T, store dataservices.DataStore) portaineree.Endpoint {
	return createEndpointWithId(t, store, 5)
}

func createEdgeStack(t *testing.T, store dataservices.DataStore, endpointID portaineree.EndpointID) portaineree.EdgeStack {
	t.Helper()

	edgeGroup := portaineree.EdgeGroup{
		ID:           1,
		Name:         "EdgeGroup 1",
		Dynamic:      false,
		TagIDs:       nil,
		Endpoints:    []portaineree.EndpointID{endpointID},
		PartialMatch: false,
	}

	err := store.EdgeGroup().Create(&edgeGroup)
	if err != nil {
		t.Fatal(err)
	}

	edgeStackID := portaineree.EdgeStackID(14)
	edgeStack := portaineree.EdgeStack{
		ID:   edgeStackID,
		Name: "test-edge-stack-" + strconv.Itoa(int(edgeStackID)),
		Status: map[portaineree.EndpointID]portainer.EdgeStackStatus{
			endpointID: {Details: portainer.EdgeStackStatusDetails{Ok: true}, Error: "", EndpointID: portainer.EndpointID(endpointID)},
		},
		CreationDate:   time.Now().Unix(),
		EdgeGroups:     []portaineree.EdgeGroupID{edgeGroup.ID},
		ProjectPath:    "tmpDir",
		EntryPoint:     "entrypoint",
		Version:        237,
		ManifestPath:   "tmpDir",
		DeploymentType: portaineree.EdgeStackDeploymentKubernetes,
	}

	endpointRelation := portaineree.EndpointRelation{
		EndpointID: endpointID,
		EdgeStacks: map[portaineree.EdgeStackID]bool{
			edgeStack.ID: true,
		},
	}

	err = store.EdgeStack().Create(edgeStack.ID, &edgeStack)
	if err != nil {
		t.Fatal(err)
	}

	err = store.EndpointRelation().Create(&endpointRelation)
	if err != nil {
		t.Fatal(err)
	}

	return edgeStack
}

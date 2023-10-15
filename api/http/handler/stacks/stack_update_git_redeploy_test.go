package stacks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

type mockGitService struct {
}

func newMockGitService() portainer.GitService {
	return &mockGitService{}
}

func (g *mockGitService) CloneRepository(destination, repositoryURL, referenceName, username, password string, tlsSkipVerify bool) error {
	return os.MkdirAll(destination, 0644)
}

func (g *mockGitService) LatestCommitID(repositoryURL, referenceName, username, password string, tlsSkipVerify bool) (string, error) {
	if referenceName == "update" {
		return fmt.Sprintf("update-%d", time.Now().UnixNano()), nil
	}

	return "123", nil
}

func (g *mockGitService) ListRefs(repositoryURL, username, password string, hardRefresh bool, tlsSkipVerify bool) ([]string, error) {
	return nil, nil
}

func (g *mockGitService) ListFiles(repositoryURL, referenceName, username, password string, dirOnly, hardRefresh bool, includedExts []string, tlsSkipVerify bool) ([]string, error) {
	return nil, nil
}

func mockCreateGitRedeployStackRequest(stackID portainer.StackID, endpointId portainer.EndpointID, payload []byte) *http.Request {
	target := fmt.Sprintf("/stacks/%d/git/redeploy?endpointId=%d", stackID, endpointId)
	return mockCreateStackRequestWithSecurityContext(http.MethodPut, target, bytes.NewBuffer(payload))
}

func TestHandler_StackGitRedeploy_KeepOneVersionFolder(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	_, err := mockCreateUser(store)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := mockCreateEndpoint(store)
	if err != nil {
		t.Fatal(err)
	}

	gitService := newMockGitService()

	tempDir := t.TempDir()
	fileService, err := filesystem.NewService(tempDir, "")
	if err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(testhelpers.NewTestRequestBouncer(), store, testhelpers.NewUserActivityService())
	handler.FileService = fileService
	handler.GitService = gitService
	handler.StackDeployer = testhelpers.NewTestStackDeployer()

	stack := &portaineree.Stack{
		ID:          1,
		EndpointID:  endpoint.ID,
		ProjectPath: tempDir,
		Type:        portaineree.DockerComposeStack,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "123",
		},
		PreviousDeploymentInfo: &portainer.StackDeploymentInfo{
			ConfigHash: "122",
		},
	}
	err = store.Stack().Create(stack)
	if err != nil {
		t.Fatal(err)
	}

	versionFolder := fileService.FormProjectPathByVersion(stack.ProjectPath, 0, stack.GitConfig.ConfigHash)
	err = os.Mkdir(versionFolder, 0644)
	if err != nil {
		t.Fatal(err)
	}

	previousVersionFolder := fileService.FormProjectPathByVersion(stack.ProjectPath, 0, stack.PreviousDeploymentInfo.ConfigHash)
	err = os.Mkdir(previousVersionFolder, 0644)
	if err != nil {
		t.Fatal(err)
	}

	resourceContrl := &portainer.ResourceControl{
		ID:         1,
		ResourceID: stackutils.ResourceControlID(stack.EndpointID, stack.Name),
		Type:       portaineree.StackResourceControl,
	}
	err = store.ResourceControl().Create(resourceContrl)
	if err != nil {
		t.Fatal(err)
	}

	// Case1: No version folders will be deleted when the commit hash is
	// not changed
	t.Run("no existing version folders are removed when redeploying with the same commit", func(t *testing.T) {
		stack, err := store.Stack().Read(stack.ID)
		if err != nil {
			t.Fatal(err)
		}

		data := stackGitRedployPayload{
			RepositoryReferenceName: "ref",
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req := mockCreateGitRedeployStackRequest(stack.ID, stack.EndpointID, payload)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code, "expected status code 200")

		is.DirExists(versionFolder, "expected current version folder to be kept")
		is.DirExists(previousVersionFolder, "expected previous version folder to be kept")

		result, err := store.Stack().Read(stack.ID)
		if err != nil {
			t.Fatal(err)
		}

		newVersionFolder := fileService.FormProjectPathByVersion(stack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.Equal(versionFolder, newVersionFolder, "expected current version folder is not changed")
	})

	// Case2: When there are only two version folders, one is
	// GitConfig.ConfigHash and the other is PreviousDeploymentInfo.ConfigHash,
	// redeploying stack will keep the new version folder only, the
	// existing two version folders are deleted
	t.Run("remove current and previous version folders when a new commit is redeployed", func(t *testing.T) {
		data := stackGitRedployPayload{
			RepositoryReferenceName: "update",
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req := mockCreateGitRedeployStackRequest(stack.ID, stack.EndpointID, payload)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code, "expected status code 200")

		is.NoDirExists(versionFolder, "expected current version folder to be deleted")
		is.NoDirExists(previousVersionFolder, "expected previous version folder to be deleted")

		result, err := store.Stack().Read(stack.ID)
		if err != nil {
			t.Fatal(err)
		}

		newVersionFolder := fileService.FormProjectPathByVersion(stack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(newVersionFolder, "expected new version folder to be kept")
	})

	// Case3: When there is only one version folder, redeploying stack
	// will keepthe new version folder only, the existing version folder
	// is deleted
	t.Run("remove current version folder when a new commit redeployed", func(t *testing.T) {
		stack, err := store.Stack().Read(stack.ID)
		if err != nil {
			t.Fatal(err)
		}

		data := stackGitRedployPayload{
			RepositoryReferenceName: "update",
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req := mockCreateGitRedeployStackRequest(stack.ID, stack.EndpointID, payload)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code, "expected status code 200")

		versionFolder := fileService.FormProjectPathByVersion(stack.ProjectPath, 0, stack.GitConfig.ConfigHash)
		is.NoDirExists(versionFolder, "expected current version folder to be deleted")

		result, err := store.Stack().Read(stack.ID)
		if err != nil {
			t.Fatal(err)
		}

		newVersionFolder := fileService.FormProjectPathByVersion(stack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(newVersionFolder, "expected new version folder to be kept")
	})
}

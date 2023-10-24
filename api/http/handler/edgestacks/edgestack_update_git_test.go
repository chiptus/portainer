package edgestacks

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
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

type mockGitService struct {
}

func newMockGitService() portainer.GitService {
	return &mockGitService{}
}

func (g *mockGitService) CloneRepository(destination, repositoryURL, referenceName, username, password string, tlsSkipVerify bool) error {
	return os.Mkdir(destination, 0644)
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

func TestHandler_EdgeStackUpdateFromGit_RemovePreviousVersionFolder(t *testing.T) {
	is := assert.New(t)

	handler, rawAPIKey := setupHandler(t)
	handler.GitService = newMockGitService()

	edgeStack := &portaineree.EdgeStack{
		ID:          1,
		ProjectPath: handler.FileService.GetEdgeStackProjectPath("1"),
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "123",
		},
	}

	err := handler.DataStore.EdgeStack().Create(edgeStack.ID, edgeStack)
	if err != nil {
		t.Fatal(err)
	}

	versionFolder1 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, edgeStack.GitConfig.ConfigHash)

	err = os.MkdirAll(versionFolder1, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Case1: When there is only one version folder, redeploying stack
	// will not remove any version folders. Keep two latest version folders
	t.Run("No version folder will be removed if the previous version is null during redeployment", func(t *testing.T) {
		edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		data := stackGitUpdatePayload{
			RefName:       "update",
			UpdateVersion: true,
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPut, "/edge_stacks/1/git", bytes.NewBuffer(payload))
		req.Header.Add("x-api-key", rawAPIKey)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusNoContent, rr.Code, "expected status code 200")

		result, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		previousVersionFolder := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.PreviousDeploymentInfo.ConfigHash)
		is.DirExists(previousVersionFolder, "expected previous version folder to be kept")
		is.Equal(versionFolder1, previousVersionFolder, "expected previous current version folder changed to previous version folder")

		versionFolder2 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(versionFolder2, "expected current version folder to be kept")
	})

	// Case2: When there are two version folders, one is GitConfig.ConfigHash
	// and the other is PreviousDeploymentInfo.ConfigHash,
	// redeploying stack will keep the latest two version folders
	t.Run("keep the latest two version folders during redeployment", func(t *testing.T) {
		edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		folderToBeRemoved := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, edgeStack.PreviousDeploymentInfo.ConfigHash)

		data := stackGitUpdatePayload{
			RefName:       "update",
			UpdateVersion: true,
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPut, "/edge_stacks/1/git", bytes.NewBuffer(payload))
		req.Header.Add("x-api-key", rawAPIKey)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusNoContent, rr.Code, "expected status code 200")

		is.Equal(versionFolder1, folderToBeRemoved, "expected previous version folder is marked as to be removed")
		is.NoDirExists(folderToBeRemoved, "expected previous version folder to be removed")

		result, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		versionFolder2 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.PreviousDeploymentInfo.ConfigHash)
		is.DirExists(versionFolder2, "expected previous version folder to be kept")

		versionFolder3 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(versionFolder3, "expected current version folder to be kept")
	})

	// Case3: when there is no new commit, redeploying stack will
	// not remove any version folders
	t.Run("no version folder will be removed if there is no new commit", func(t *testing.T) {
		edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		data := stackGitUpdatePayload{
			UpdateVersion: false,
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPut, "/edge_stacks/1/git", bytes.NewBuffer(payload))
		req.Header.Add("x-api-key", rawAPIKey)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusNoContent, rr.Code, "expected status code 200")

		result, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		versionFolder2 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.PreviousDeploymentInfo.ConfigHash)
		is.DirExists(versionFolder2, "expected previous version folder to be kept")

		versionFolder3 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(versionFolder3, "expected current version folder to be kept")
	})

	// Case4: parallel update will only keep the latest two version folders
	t.Run("keep the latest two version folders during parallel update", func(t *testing.T) {
		edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		folderToBeRemoved := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, edgeStack.PreviousDeploymentInfo.ConfigHash)

		handler.staggerService = testhelpers.NewTestStaggerService()
		data := stackGitUpdatePayload{
			RefName:       "update",
			UpdateVersion: true,
			StaggerConfig: &portaineree.EdgeStaggerConfig{
				StaggerOption:         portaineree.EdgeStaggerOptionParallel,
				StaggerParallelOption: portaineree.EdgeStaggerParallelOptionFixed,
				DeviceNumber:          1,
				UpdateFailureAction:   portaineree.EdgeUpdateFailureActionContinue,
			},
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPut, "/edge_stacks/1/git", bytes.NewBuffer(payload))
		req.Header.Add("x-api-key", rawAPIKey)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(http.StatusNoContent, rr.Code, "expected status code 200")
		is.NoDirExists(folderToBeRemoved, "expected previous version folder to be removed")

		result, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		versionFolder3 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.PreviousDeploymentInfo.ConfigHash)
		is.DirExists(versionFolder3, "expected previous version folder to be kept")

		versionFolder4 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(versionFolder4, "expected current version folder to be kept")
	})
}

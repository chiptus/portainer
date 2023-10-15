package edgestacks

import (
	"os"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

func TestHandler_AutoUpdate_RemovePreviousVersionFolder(t *testing.T) {
	is := assert.New(t)

	handler, _ := setupHandler(t)
	handler.GitService = newMockGitService()

	edgeStack := &portaineree.EdgeStack{
		ID:          1,
		ProjectPath: handler.FileService.GetEdgeStackProjectPath("1"),
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "123",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate: true,
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
		// enable to generate a new commit ID
		edgeStack.GitConfig.ReferenceName = "update"
		err = handler.DataStore.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack, true)
		if err != nil {
			t.Fatal(err)
		}

		err = handler.autoUpdate(edgeStack.ID, nil)
		is.NoError(err, "expected no error during auto update")

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

		// enable to generate a new commit ID
		edgeStack.GitConfig.ReferenceName = "update"
		err = handler.DataStore.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack, true)
		if err != nil {
			t.Fatal(err)
		}

		err = handler.autoUpdate(edgeStack.ID, nil)
		is.NoError(err, "expected no error during auto update")

		is.Equal(versionFolder1, folderToBeRemoved, "expected previous version folder is marked as to be removed")
		is.NoDirExists(folderToBeRemoved, "expected previous version folder to be removed")

		result, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		versionFolder2 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.PreviousDeploymentInfo.ConfigHash)
		is.DirExists(versionFolder2, "expected previous version folder to be kept")
		is.Contains(versionFolder2, result.PreviousDeploymentInfo.ConfigHash, "expected version folder 2 moved to previous version folder")

		versionFolder3 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(versionFolder3, "expected current version folder to be kept")
		is.Contains(versionFolder3, result.GitConfig.ConfigHash, "expected version folder 3 becomes current version folder")
	})

	// Case3: when there is no new commit, redeploying stack will
	// not remove any version folders
	t.Run("no version folder will be removed if there is no new commit", func(t *testing.T) {
		edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		err = handler.autoUpdate(edgeStack.ID, nil)
		is.NoError(err, "expected no error during auto update")

		result, err := handler.DataStore.EdgeStack().EdgeStack(edgeStack.ID)
		if err != nil {
			t.Fatal(err)
		}

		versionFolder2 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.PreviousDeploymentInfo.ConfigHash)
		is.DirExists(versionFolder2, "expected previous version folder to be kept")
		is.Contains(versionFolder2, result.PreviousDeploymentInfo.ConfigHash, "expected version folder 2 is stll the previous version folder")

		versionFolder3 := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, result.GitConfig.ConfigHash)
		is.DirExists(versionFolder3, "expected current version folder to be kept")
		is.Contains(versionFolder3, result.GitConfig.ConfigHash, "expected version folder 3 is stll the current version folder")
	})
}

package migrator

import (
	"os"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

func TestRemoveGitStackVersionFolders(t *testing.T) {
	is := assert.New(t)

	tempDir := t.TempDir()
	projectPath := filesystem.JoinPaths(tempDir, "1")
	err := os.MkdirAll(projectPath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	gitConfig := &gittypes.RepoConfig{
		ConfigHash: "kept-path-1",
	}

	prevInfo := &portainer.StackDeploymentInfo{
		ConfigHash: "kept-path-2",
	}

	simulatedPaths := []string{
		filesystem.JoinPaths(projectPath, gitConfig.ConfigHash),
		filesystem.JoinPaths(projectPath, prevInfo.ConfigHash),
		filesystem.JoinPaths(projectPath, "removed-path-1"),
		filesystem.JoinPaths(projectPath, "removed-path-2"),
	}

	for _, path := range simulatedPaths {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			t.Fatal(err)
		}
	}

	formPath := func(projectPath string, version int, configHash string) string {
		return filesystem.JoinPaths(projectPath, configHash)
	}

	t.Run("should not remove any folder if git config is nil", func(t *testing.T) {
		err := RemoveGitStackVersionFolders(projectPath, nil, nil, formPath)
		is.NoError(err, "expected no error occurs")
		for _, path := range simulatedPaths {
			is.DirExists(path, "expected path exists")
		}
	})

	t.Run("keep the latest version folder and the previous version folder", func(t *testing.T) {
		err := RemoveGitStackVersionFolders(projectPath, gitConfig, prevInfo, formPath)
		is.NoError(err, "expected no error occurs")
		is.DirExists(filesystem.JoinPaths(projectPath, gitConfig.ConfigHash), "expected kept path exists")
		is.DirExists(filesystem.JoinPaths(projectPath, prevInfo.ConfigHash), "expected kept path exists")
		is.NoDirExists(filesystem.JoinPaths(projectPath, "removed-path-1"), "expected removed path does not exist")
		is.NoDirExists(filesystem.JoinPaths(projectPath, "removed-path-2"), "expected removed path does not exist")
	})

	t.Run("only keep the latest version folder", func(t *testing.T) {
		err := RemoveGitStackVersionFolders(projectPath, gitConfig, nil, formPath)
		is.NoError(err, "expected no error occurs")
		is.DirExists(filesystem.JoinPaths(projectPath, gitConfig.ConfigHash), "expected kept path exists")
		is.NoDirExists(filesystem.JoinPaths(projectPath, prevInfo.ConfigHash), "expected removed path does not exist")
	})
}

func TestRemoveUnreferencedGitStackVersionFolders(t *testing.T) {
	is := assert.New(t)

	tempDir := t.TempDir()
	root := filesystem.JoinPaths(tempDir, "1")
	err := os.MkdirAll(root, 0700)
	if err != nil {
		t.Fatal(err)
	}

	keptPaths := []string{
		filesystem.JoinPaths(root, "kept-path-1"),
		filesystem.JoinPaths(root, "kept-path-2"),
	}
	for _, path := range keptPaths {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			t.Fatal(err)
		}
	}

	foldersToBeRemoved := []string{
		filesystem.JoinPaths(root, "removed-path-1"),
		filesystem.JoinPaths(root, "removed-path-2"),
	}
	for _, path := range foldersToBeRemoved {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = removeUnreferencedGitStackVersionFolders(root, keptPaths)
	is.NoError(err, "expected no error occurs")

	for _, path := range keptPaths {
		is.DirExists(path, "expected kept path exists")
	}

	for _, path := range foldersToBeRemoved {
		is.NoDirExists(path, "expected removed path does not exist")
	}
}

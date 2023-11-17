package stacks

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/git"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
)

func ensureIntegrationTest(t *testing.T) {
	if _, ok := os.LookupEnv("INTEGRATION_TEST"); !ok {
		t.Skip("skip an integration test")
	}
}

func testGetYAMLContentDeployedByAppTemplate(env string) string {
	return `version: '2'

services:
   db:
     image: mysql:5.7
     volumes:
       - db_data:/var/lib/mysql
     restart: always
     environment:
       MYSQL_ROOT_PASSWORD: ${MYSQL_DATABASE_PASSWORD}
       MYSQL_DATABASE: wordpress
       MYSQL_USER: wordpress
       MYSQL_PASSWORD: wordpress

   wordpress:
     image: wordpress:latest
     ports:
       - 80
     restart: always
     environment:
       WORDPRESS_DB_HOST: db:3306
       WORDPRESS_DB_USER: wordpress
       WORDPRESS_DB_PASSWORD: wordpress
       ` + env + `

volumes:
    db_data:
`
}

func testGetYAMLContentDeployedByGitRepo(env string) string {
	return `version: '3'

services:
  httpd-container:
    image: httpd:latest
    container_name: httpd-test-container
    labels: 
      version: 1
      ` + env + `
    ports:
      - "${PORT:-5567}:80"
`
}

func testGetYAMLContentDeployedByFileContent(env string) string {
	return `version: '3'

services:
  httpd-container:
    image: httpd:latest
    container_name: httpd-test-container
    labels: 
      version: 1
      ` + env + `
    ports:
      - "${PORT:-5567}:80"
`
}

func testCreateStackFromAppTemplate(store *datastore.Store, gitService portainer.GitService, fileService portainer.FileService) (*portaineree.Stack, error) {
	stack := &portaineree.Stack{
		ID:         1,
		Name:       "testStack1",
		EndpointID: 1,
		EntryPoint: "stacks/wordpress/docker-compose.yml",
		Env: []portainer.Pair{
			{Name: "MYSQL_DATABASE_PASSWORD", Value: "wordpress"},
		},
		FromAppTemplate: true,
		GitConfig: &gittypes.RepoConfig{
			ConfigFilePath: "stacks/wordpress/docker-compose.yml",
			URL:            "https://github.com/portainer/templates",
		},
		ProjectPath:      fmt.Sprintf("%s/compose/1", fileService.GetDatastorePath()),
		StackFileVersion: 1,
		Status:           portaineree.StackStatusActive,
		Type:             portaineree.DockerComposeStack}

	err := store.Stack().Create(stack)
	if err != nil {
		return nil, err
	}

	err = testCloneGitRepo(store, gitService, fileService, stack)

	return stack, err
}

func testCreateStackByGitRepo(store *datastore.Store, gitService portainer.GitService, fileService portainer.FileService) (*portaineree.Stack, error) {
	stack := &portaineree.Stack{
		ID:         2,
		Name:       "testStack2",
		EndpointID: 1,
		EntryPoint: "docker-compose.yml",
		GitConfig: &gittypes.RepoConfig{
			ConfigFilePath: "docker-compose.yml",
			ReferenceName:  "refs/heads/archive",
			URL:            "https://github.com/oscarzhou-portainer/repo-httpd",
		},
		ProjectPath:      fmt.Sprintf("%s/compose/2", fileService.GetDatastorePath()),
		StackFileVersion: 1,
		Status:           portaineree.StackStatusActive,
		Type:             portaineree.DockerComposeStack}

	err := store.Stack().Create(stack)
	if err != nil {
		return nil, err
	}

	err = testCloneGitRepo(store, gitService, fileService, stack)

	return stack, err
}

func testCreateStackByFileContent(store *datastore.Store, fileService portainer.FileService) (*portaineree.Stack, error) {
	stack := &portaineree.Stack{
		ID:               3,
		Name:             "testStack3",
		EndpointID:       1,
		EntryPoint:       "docker-compose.yml",
		ProjectPath:      fmt.Sprintf("%s/compose/3", fileService.GetDatastorePath()),
		StackFileVersion: 1,
		Status:           portaineree.StackStatusActive,
		Type:             portaineree.DockerComposeStack}

	err := store.Stack().Create(stack)
	if err != nil {
		return nil, err
	}

	fileContent := []byte(testGetYAMLContentDeployedByFileContent(""))
	_, err = fileService.StoreStackFileFromBytesByVersion(strconv.Itoa(int(stack.ID)), stack.EntryPoint, stack.StackFileVersion, fileContent)

	return stack, err
}

func testDetachGitRepoFromStack(store *datastore.Store, stack *portaineree.Stack) error {
	stack.GitConfig = nil
	stack.PreviousDeploymentInfo = nil
	stack.IsDetachedFromGit = true
	return store.Stack().Update(stack.ID, stack)
}

func testIncrementFileVersion(store *datastore.Store, stack *portaineree.Stack) error {
	stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
		FileVersion: stack.StackFileVersion,
	}
	stack.StackFileVersion++
	return store.Stack().Update(stack.ID, stack)
}

func testCloneGitRepo(store *datastore.Store, gitService portainer.GitService, fileService portainer.FileService, stack *portaineree.Stack) error {
	stackFolder := strconv.Itoa(int(stack.ID))
	getProjectPath := func(enableVersionFolder bool, commitHash string) string {
		if enableVersionFolder {
			return fileService.GetStackProjectPathByVersion(stackFolder, 0, commitHash)
		}
		return fileService.GetStackProjectPath(stackFolder)
	}

	commitID, err := stackutils.DownloadGitRepository(*stack.GitConfig, gitService, true, getProjectPath)
	stack.GitConfig.ConfigHash = commitID
	return err
}

func mockCreateUpdateStackRequest(stackID portainer.StackID, payload []byte) *http.Request {
	target := fmt.Sprintf("/stacks/%d", stackID)
	return mockCreateStackRequestWithSecurityContext(http.MethodPut, target, bytes.NewBuffer(payload))
}

func TestHandler_updateSwarmOrComposeStack(t *testing.T) {
	// start: prepare test data
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	_, err := mockCreateUser(store)
	is.NoError(err, "error creating user")

	endpoint, err := mockCreateEndpoint(store)
	is.NoError(err, "error creating endpoint")

	gitService := git.NewService(context.TODO())

	testDataPath := filepath.Join(os.TempDir(), fmt.Sprintf("portainer-teststacks-%d", time.Now().Unix()))

	fileService, err := filesystem.NewService(testDataPath, "")
	is.NoError(err, "error init file service")

	stackDeployedFromAppTemplate, err := testCreateStackFromAppTemplate(store, gitService, fileService)
	is.NoError(err, "error creating stack deployed from app template")

	stackDeployedByGitRepo, err := testCreateStackByGitRepo(store, gitService, fileService)
	is.NoError(err, "error creating stack deployed from git repo")

	stackDeployedByFileContent, err := testCreateStackByFileContent(store, fileService)
	is.NoError(err, "error creating stack deployed by file content")

	h := NewHandler(testhelpers.NewTestRequestBouncer(), store, nil)

	h.FileService = fileService
	h.StackDeployer = testhelpers.NewTestStackDeployer()
	// end: prepare test data

	// case 1:
	// folder structure before update: "/data/compose/1/f94735d43bc9a046bc0f6a794f588140db860742"
	// folder structure after update: "/data/compose/1/v2"
	t.Run("update stack deployed from app template when yaml content is changed", func(t *testing.T) {
		yamlContentWithChange := testGetYAMLContentDeployedByGitRepo("FOO: ${FOO}")
		data := &updateStackPayload{
			StackFileContent: yamlContentWithChange,
			Env: []portainer.Pair{
				{
					Name: "MYSQL_DATABASE_PASSWORD", Value: "wordpress"},
				{
					Name: "FOO", Value: "BAR",
				}},
		}

		payload, err := json.Marshal(data)
		is.NoError(err, "failed to marshal payload")

		req := mockCreateUpdateStackRequest(stackDeployedFromAppTemplate.ID, payload)

		httpErr := h.updateSwarmOrComposeStack(req, stackDeployedFromAppTemplate, endpoint)
		is.Nil(httpErr, "error should be nil")

		expectedDir := fileService.GetStackProjectPathByVersion(
			strconv.Itoa(int(stackDeployedFromAppTemplate.ID)),
			stackDeployedFromAppTemplate.StackFileVersion,
			"")
		is.DirExists(expectedDir, "/compose/1/v1 folder should exist")
	})

	// case 2.1: update stack deployed from git repo when yaml content is NOT changed
	// folder structure before update: "/data/compose/2/dc04597399c9da00895cbd791ea2ca6d1702c35f"
	// folder structure after update: "/data/compose/2/v1"
	t.Run("detach stack from git", func(t *testing.T) {
		yamlContentWithoutChange := testGetYAMLContentDeployedByGitRepo("")
		data := &updateStackPayload{
			StackFileContent: yamlContentWithoutChange,
		}
		payload, err := json.Marshal(data)
		is.NoError(err, "failed to marshal payload")

		req := mockCreateUpdateStackRequest(stackDeployedByGitRepo.ID, payload)

		httpErr := h.updateSwarmOrComposeStack(req, stackDeployedByGitRepo, endpoint)
		is.Nil(httpErr, "error should be nil")

		expectedDir := fileService.GetStackProjectPathByVersion(
			strconv.Itoa(int(stackDeployedByGitRepo.ID)),
			stackDeployedByGitRepo.StackFileVersion,
			"")
		is.DirExists(expectedDir, "/compose/2/v1 folder should exist")
	})

	// case 2.2: update stack deployed from git repo when yaml content is changed
	// folder structure before update: "/data/compose/2/v1"
	// folder structure after update: "/data/compose/2/v2"
	t.Run("update stack that has detached from git", func(t *testing.T) {
		yamlContentWithChange := testGetYAMLContentDeployedByGitRepo("FOO: ${FOO}")
		data := &updateStackPayload{
			StackFileContent: yamlContentWithChange,
		}
		payload, err := json.Marshal(data)
		is.NoError(err, "failed to marshal payload")

		req := mockCreateUpdateStackRequest(stackDeployedByGitRepo.ID, payload)

		// detach stack from git
		err = testDetachGitRepoFromStack(store, stackDeployedByGitRepo)
		is.NoError(err, "error detaching git repo from stack")

		httpErr := h.updateSwarmOrComposeStack(req, stackDeployedByGitRepo, endpoint)
		is.Nil(httpErr, "error should be nil")

		expectedDir := fileService.GetStackProjectPathByVersion(
			strconv.Itoa(int(stackDeployedByGitRepo.ID)),
			stackDeployedByGitRepo.StackFileVersion,
			"")
		is.DirExists(expectedDir, "/compose/2/v2 folder should exist")
	})

	// case 3.1: update stack deployed by file content when yaml content is changed
	// folder structure before update: "/data/compose/3/v1"
	// folder structure after update: "/data/compose/3/v2" and "/data/compose/3/v1"
	t.Run("update stack deployed by file content with changed content", func(t *testing.T) {
		yamlContentWithChange := testGetYAMLContentDeployedByFileContent("FOO: ${FOO}")
		data := &updateStackPayload{
			StackFileContent: yamlContentWithChange,
		}
		payload, err := json.Marshal(data)
		is.NoError(err, "failed to marshal payload")

		req := mockCreateUpdateStackRequest(stackDeployedByFileContent.ID, payload)

		httpErr := h.updateSwarmOrComposeStack(req, stackDeployedByFileContent, endpoint)
		is.Nil(httpErr, "error should be nil")

		expectedDir := fileService.GetStackProjectPathByVersion(
			strconv.Itoa(int(stackDeployedByFileContent.ID)),
			stackDeployedByFileContent.StackFileVersion,
			"")
		is.DirExists(expectedDir, "/compose/3/v2 folder should exist")
	})

	// case 3.2: update stack deployed by file content with rollback
	// folder structure before update: "/data/compose/3/v1" and "/data/compose/3/v2"
	// folder structure after update: "/data/compose/3/v1" and "/data/compose/3/v2"
	t.Run("rollback to the pevious version", func(t *testing.T) {
		yamlContentWithChange := testGetYAMLContentDeployedByFileContent("")
		rollbackTo := new(int)
		*rollbackTo = stackDeployedByFileContent.PreviousDeploymentInfo.FileVersion
		fmt.Println("previous version=", *rollbackTo)
		data := &updateStackPayload{
			StackFileContent: yamlContentWithChange,
			RollbackTo:       rollbackTo,
		}
		payload, err := json.Marshal(data)
		is.NoError(err, "failed to marshal payload")

		req := mockCreateUpdateStackRequest(stackDeployedByFileContent.ID, payload)

		httpErr := h.updateSwarmOrComposeStack(req, stackDeployedByFileContent, endpoint)
		is.Nil(httpErr, "error should be nil")

		expectedDir := fileService.GetStackProjectPathByVersion(
			strconv.Itoa(int(stackDeployedByFileContent.ID)),
			*rollbackTo,
			"")
		is.DirExists(expectedDir, "/compose/3/v1 folder should exist")

		is.Nil(stackDeployedByFileContent.PreviousDeploymentInfo, "previous deployment info should be nil")
	})
}

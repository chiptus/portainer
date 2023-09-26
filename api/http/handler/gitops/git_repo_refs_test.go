package gitops

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

const (
	testRepo     string = "https://github.com/portainer/git-test.git"
	testUsername string = "test-username"
	testPassword string = "test-password"
	targetURL    string = "/gitops/repo/refs"
)

type TestGitService struct {
	portainer.GitService
}

func (g *TestGitService) ListRefs(repositoryURL, username, password string, hardRefresh bool, tlsSkipVerify bool) ([]string, error) {
	if repositoryURL == testRepo && testUsername == username && testPassword == password {
		return []string{"refs/head/main", "refs/head/test"}, nil
	}
	return nil, gittypes.ErrAuthenticationFailure
}

func listWithNoError(rr *httptest.ResponseRecorder, is *assert.Assertions) {
	is.Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	is.NoError(err, "ReadAll should not return error")

	var resp []string
	err = json.Unmarshal(body, &resp)
	is.NoError(err, "response should be list json")

	is.GreaterOrEqual(len(resp), 1)
}

func Test_gitOperationRepoRefs(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	// create user(s)
	user := &portaineree.User{ID: 1, Username: "standard", Role: portaineree.StandardUserRole, PortainerAuthorizations: authorization.DefaultPortainerAuthorizations()}
	err := store.User().Create(user)
	is.NoError(err, "error creating user")

	// create git credential
	gitCredential := &portaineree.GitCredential{ID: 1, UserID: user.ID, Name: "test-name", Username: testUsername, Password: testPassword}
	err = store.GitCredentialService.Create(gitCredential)
	is.NoError(err, "error creating git credential")

	// create stack with git username/password
	stackWithGitPassword := &portaineree.Stack{ID: 1, GitConfig: &gittypes.RepoConfig{Authentication: &gittypes.GitAuthentication{
		Username: testUsername,
		Password: testPassword,
	}}}
	err = store.StackService.Create(stackWithGitPassword)
	is.NoError(err, "error creating stack with git username/password")

	// create stack with git credential
	stackWithGitCredential := &portaineree.Stack{ID: 2, GitConfig: &gittypes.RepoConfig{Authentication: &gittypes.GitAuthentication{
		GitCredentialID: 1,
	}}}
	err = store.StackService.Create(stackWithGitCredential)
	is.NoError(err, "error creating stack with git credential")

	// setup services
	gitService := &TestGitService{}

	h := NewHandler(testhelpers.NewTestRequestBouncer(), store, gitService, nil)

	type response struct {
		Message string `json:"message"`
		Details string `json:"details"`
	}

	t.Run("query parameter (repo) must be provided", func(t *testing.T) {
		data := repositoryReferenceListPayload{}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusBadRequest, rr.Code)
	})

	t.Run("fail to authenticate git credential", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository: testRepo,
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusInternalServerError, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp response
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")
		is.Equal(gittypes.ErrAuthenticationFailure.Error(), resp.Details)
	})

	t.Run("list refs with git username/password", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository: testRepo,
			Username:   testUsername,
			Password:   testPassword,
		}

		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		listWithNoError(rr, is)
	})

	t.Run("list refs with git credential ID", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository:      testRepo,
			GitCredentialID: int(gitCredential.ID),
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		listWithNoError(rr, is)
	})

	t.Run("list refs with stack ID when stack is configured with git password", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository: testRepo,
			StackID:    stackWithGitPassword.ID,
		}

		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		listWithNoError(rr, is)
	})

	t.Run("list refs with stack ID if the stack is configured with git credential", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository: testRepo,
			StackID:    stackWithGitCredential.ID,
		}

		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		listWithNoError(rr, is)
	})

	t.Run("list refs with both stack ID and git credentialID even if the stack is configured with git credential", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository:      testRepo,
			StackID:         stackWithGitCredential.ID,
			GitCredentialID: int(gitCredential.ID),
		}

		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		listWithNoError(rr, is)
	})
}

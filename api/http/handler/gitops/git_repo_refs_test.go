package gitops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

const (
	testRepo     string = "https://github.com/portainer/git-test.git"
	testUsername string = "test-username"
	testPassword string = "test-password"
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

	// create stack
	stack := &portaineree.Stack{ID: 1, GitConfig: &gittypes.RepoConfig{Authentication: &gittypes.GitAuthentication{
		Username: testUsername,
		Password: testPassword,
	}}}
	err = store.StackService.Create(stack)
	is.NoError(err, "error creating stack")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	requestBouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, nil, nil)

	gitService := &TestGitService{}

	h := NewHandler(requestBouncer, store, gitService, nil)

	// generate standard and admin user tokens
	jwt, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})

	type response struct {
		Message string `json:"message"`
		Details string `json:"details"`
	}
	t.Run("query parameter (repo) must be provided", func(t *testing.T) {
		data := repositoryReferenceListPayload{}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/refs", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

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
		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/refs", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

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

	t.Run("authenticated user can list refs with username/password", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository: testRepo,
			Username:   testUsername,
			Password:   testPassword,
		}
		payload, err := json.Marshal(data)
		is.NoError(err)
		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/refs", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []string
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.GreaterOrEqual(len(resp), 1)
	})

	t.Run("authenticated user can list refs with git credential ID", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository:      testRepo,
			GitCredentialID: 1,
		}
		payload, err := json.Marshal(data)
		is.NoError(err)
		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/refs", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []string
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.GreaterOrEqual(len(resp), 1)
	})

	t.Run("authenticated user can list refs with stack ID", func(t *testing.T) {
		data := repositoryReferenceListPayload{
			Repository: testRepo,
			StackID:    1,
		}
		payload, err := json.Marshal(data)
		is.NoError(err)
		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/refs", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []string
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.GreaterOrEqual(len(resp), 1)
	})
}

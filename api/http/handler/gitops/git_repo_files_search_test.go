package gitops

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer/api/git"
	gittypes "github.com/portainer/portainer/api/git/types"

	"github.com/stretchr/testify/assert"
)

func Test_gitOperationRepoFilesSearch(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	// create  user(s)
	user := &portaineree.User{ID: 1, Username: "standard", Role: portaineree.StandardUserRole, PortainerAuthorizations: authorization.DefaultPortainerAuthorizations()}
	err := store.User().Create(user)
	is.NoError(err, "error creating user")

	// setup services
	gitService := git.NewService(context.TODO())

	h := NewHandler(testhelpers.NewTestRequestBouncer(), store, gitService, nil)

	type response struct {
		Message string `json:"message"`
		Details string `json:"details"`
	}

	t.Run("query parameter repo must be provided", func(t *testing.T) {
		data := repositoryFileSearchPayload{}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/files/search", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusBadRequest, rr.Code)
	})

	t.Run("fail to authenticate git credential", func(t *testing.T) {
		data := repositoryFileSearchPayload{
			Repository: "https://github.com/portainer/portainer-ee.git",
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/files/search", bytes.NewBuffer(payload))
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

	t.Run("authenticated user can list files of a git repository", func(t *testing.T) {
		data := repositoryFileSearchPayload{
			Repository: "https://github.com/portainer/portainer.git",
			Reference:  "refs/heads/develop",
			Keyword:    "docker",
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/files/search", bytes.NewBuffer(payload))
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

	t.Run("authenticated user can list files of a git repository without providing keyword", func(t *testing.T) {
		data := repositoryFileSearchPayload{
			Repository: "https://github.com/portainer/portainer.git",
			Reference:  "refs/heads/develop",
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/gitops/repo/files/search", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []string
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.Greater(len(resp), 0)
	})
}

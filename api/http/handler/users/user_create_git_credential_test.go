package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/stretchr/testify/assert"
)

func Test_userCreateGitCredential(t *testing.T) {
	is := assert.New(t)

	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()

	// create admin and standard user(s)
	adminUser := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err := store.User().Create(adminUser)
	is.NoError(err, "error creating admin user")

	user := &portaineree.User{ID: 2, Username: "standard", Role: portaineree.StandardUserRole, PortainerAuthorizations: authorization.DefaultPortainerAuthorizations()}
	err = store.User().Create(user)
	is.NoError(err, "error creating user")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, apiKeyService, nil)
	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	passwordChecker := security.NewPasswordStrengthChecker(store.SettingsService)

	h := NewHandler(requestBouncer, rateLimiter, apiKeyService, testhelpers.NewUserActivityService(), &demo.Service{}, passwordChecker)
	h.DataStore = store

	// generate standard and admin user git credential
	adminJWT, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})
	jwt, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})

	t.Run("standard user successfully creates git credential", func(t *testing.T) {
		data := userGitCredentialCreatePayload{
			Name:     "test-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/gitcredentials", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusCreated, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp gitCredentialResponse
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be json")
		is.EqualValues(data.Name, resp.GitCredential.Name)
		is.EqualValues(data.Username, resp.GitCredential.Username)
		is.Empty(resp.GitCredential.Password)
	})

	t.Run("standard user cannot creates git credential with invalid name", func(t *testing.T) {
		data := userGitCredentialCreatePayload{
			Name:     "TEST**",
			Username: "test-username",
			Password: "test-password",
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/gitcredentials", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusBadRequest, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})

	t.Run("git credential name is unique to the user", func(t *testing.T) {
		data := userGitCredentialCreatePayload{
			Name:     "test-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}

		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/gitcredentials", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusBadRequest, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})

	t.Run("admin cannot create git credential for standard user", func(t *testing.T) {
		data := userGitCredentialCreatePayload{
			Name:     "test-admin-credential",
			Username: "test-username",
			Password: "test-password",
		}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/gitcredentials", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})
}

func Test_userGitCredentialCreatePayload(t *testing.T) {
	is := assert.New(t)

	tests := []struct {
		payload    userGitCredentialCreatePayload
		shouldFail bool
	}{
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test-credential",
				Username: "test username",
				Password: "test password",
			},
			shouldFail: false,
		},
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test credential",
				Username: "test username",
				Password: "test password",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialCreatePayload{
				Username: "test username",
				Password: "test password",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test-credential",
				Username: "",
				Password: "test password",
			},
			shouldFail: false,
		},
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test credential",
				Username: " ",
				Password: "test password",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test credential",
				Username: "test username",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test credential",
				Username: "test username",
				Password: "",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialCreatePayload{
				Name:     "test credential",
				Username: "test username",
				Password: " ",
			},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		err := test.payload.Validate(nil)
		if test.shouldFail {
			is.Error(err)
		} else {
			is.NoError(err)
		}
	}
}

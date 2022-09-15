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

func Test_userUpdateGitCredential(t *testing.T) {
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

	// generate standard and admin user tokens
	adminJWT, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})
	jwt, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})

	t.Run("standard user can successfully update Git Credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		newAddedCred, err := store.GitCredentialService.GetGitCredentialByName(user.ID, "test-standard-credential")
		is.NoError(err)
		is.NotNil(newAddedCred)

		updatedData := userGitCredentialUpdatePayload{
			Name:     "test-standard-credential",
			Username: "test-update-username",
			Password: "test-update-password",
		}
		payload, err := json.Marshal(updatedData)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", "/users/2/gitcredentials", newAddedCred.ID), bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp gitCredentialResponse
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be json")
		is.EqualValues(updatedData.Name, resp.GitCredential.Name)
		is.EqualValues(updatedData.Username, resp.GitCredential.Username)
		is.Empty(resp.GitCredential.Password)
	})

	t.Run("standard user cannot update git credential with invalid name", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		newAddedCred, err := store.GitCredentialService.GetGitCredentialByName(user.ID, "test-standard-credential")
		is.NoError(err)
		is.NotNil(newAddedCred)

		updatedData := userGitCredentialUpdatePayload{
			Name:     "test**",
			Username: "test-update-username",
			Password: "test-update-password",
		}
		payload, err := json.Marshal(updatedData)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", "/users/2/gitcredentials", newAddedCred.ID), bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusBadRequest, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})

	t.Run("admin cannot update a standard user Git Credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		newAddedCred, err := store.GitCredentialService.GetGitCredentialByName(user.ID, "test-standard-credential")
		is.NoError(err)
		is.NotNil(newAddedCred)

		updatedData := userGitCredentialUpdatePayload{
			Name:     "test-update-credential",
			Username: "test-update-username",
			Password: "test-update-password",
		}
		payload, err := json.Marshal(updatedData)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", "/users/2/gitcredentials", newAddedCred.ID), bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)
	})
}

func Test_userGitCredentialUpdatePayload(t *testing.T) {
	is := assert.New(t)

	tests := []struct {
		payload    userGitCredentialUpdatePayload
		shouldFail bool
	}{
		{
			payload: userGitCredentialUpdatePayload{
				Name:     "test-credential",
				Username: "test username",
				Password: "test password",
			},
			shouldFail: false,
		},
		{
			payload: userGitCredentialUpdatePayload{
				Name:     "test credential",
				Username: "test username",
				Password: "test password",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialUpdatePayload{
				Name:     "",
				Username: "test username",
				Password: "test password",
			},
			shouldFail: true,
		},
		{
			payload: userGitCredentialUpdatePayload{
				Name:     "test-credential",
				Username: "test username",
				Password: "",
			},
			shouldFail: false,
		},
		{
			payload: userGitCredentialUpdatePayload{
				Name:     "test-credential",
				Username: "",
				Password: "test password",
			},
			shouldFail: false,
		},
		{
			payload: userGitCredentialUpdatePayload{
				Name:     "test credential",
				Username: " ",
				Password: "test password",
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

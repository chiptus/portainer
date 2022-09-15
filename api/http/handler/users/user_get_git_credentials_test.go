package users

import (
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

func Test_userGetGitCredentials(t *testing.T) {
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

	t.Run("standard user can successfully retrieve Git Credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-get-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/gitcredentials", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []portaineree.GitCredential
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.Len(resp, 1)
		if len(resp) == 1 {
			is.NotZero(resp[0].ID)
			is.Equal(cred.UserID, resp[0].UserID)
			is.Equal(cred.Name, resp[0].Name)
			is.Equal(cred.Username, resp[0].Username)
			is.Empty(resp[0].Password)
		}
	})

	t.Run("admin cannot retrieve standard user git credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-get-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/gitcredentials", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})
}

func Test_userGetGitCredential(t *testing.T) {
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

	t.Run("standard user can successfully retrieve single Git Credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-get-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/gitcredentials/1", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp portaineree.GitCredential
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.NotNil(resp)
		is.NotZero(resp.ID)
		is.Equal(cred.UserID, resp.UserID)
		is.Equal(cred.Name, resp.Name)
		is.Equal(cred.Username, resp.Username)
		is.Empty(resp.Password)
	})

	t.Run("admin cannot retrieve standard user git credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-get-standard-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/gitcredentials/1", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})
}

func Test_hidePasswordFields(t *testing.T) {
	is := assert.New(t)

	expect := &portaineree.GitCredential{
		ID:       1,
		UserID:   2,
		Name:     "test-credential",
		Username: "test-username",
		Password: "",
	}

	cred := &portaineree.GitCredential{
		ID:       1,
		UserID:   2,
		Name:     "test-credential",
		Username: "test-username",
		Password: "test-password",
	}

	hidePasswordFields(cred)

	is.Equal(expect, cred, "credential is not expected")
}

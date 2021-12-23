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
	bolt "github.com/portainer/portainer-ee/api/bolt/bolttest"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/stretchr/testify/assert"
)

func Test_userGetAccessTokens(t *testing.T) {
	is := assert.New(t)

	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	// create admin and standard user(s)
	adminUser := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err := store.User().CreateUser(adminUser)
	is.NoError(err, "error creating admin user")

	user := &portaineree.User{ID: 2, Username: "standard", Role: portaineree.StandardUserRole, PortainerAuthorizations: authorization.DefaultPortainerAuthorizations()}
	err = store.User().CreateUser(user)
	is.NoError(err, "error creating user")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, apiKeyService)
	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)

	h := NewHandler(requestBouncer, rateLimiter, apiKeyService, testhelpers.NewUserActivityService())
	h.DataStore = store

	// generate standard and admin user tokens
	adminJWT, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})
	jwt, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})

	t.Run("standard user can successfully retrieve API key", func(t *testing.T) {
		_, apiKey, err := apiKeyService.GenerateApiKey(*user, "test-get-token")
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/tokens", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []portaineree.APIKey
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.Len(resp, 1)
		if len(resp) == 1 {
			is.Nil(resp[0].Digest)
			is.Equal(apiKey.ID, resp[0].ID)
			is.Equal(apiKey.UserID, resp[0].UserID)
			is.Equal(apiKey.Prefix, resp[0].Prefix)
			is.Equal(apiKey.Description, resp[0].Description)
		}
	})

	t.Run("admin can retrieve standard user API Key", func(t *testing.T) {
		_, _, err := apiKeyService.GenerateApiKey(*user, "test-get-admin-token")
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/tokens", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []portaineree.APIKey
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.True(len(resp) > 0)
	})

	t.Run("user can retrieve API Key using api-key auth", func(t *testing.T) {
		rawAPIKey, _, err := apiKeyService.GenerateApiKey(*user, "test-api-key")
		is.NoError(err)

		req := httptest.NewRequest(http.MethodGet, "/users/2/tokens", nil)
		req.Header.Add("x-api-key", rawAPIKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp []portaineree.APIKey
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.True(len(resp) > 0)
	})
}

func Test_hideAPIKeyFields(t *testing.T) {
	is := assert.New(t)

	apiKey := &portaineree.APIKey{
		ID:          1,
		UserID:      2,
		Prefix:      "abc",
		Description: "test",
		Digest:      nil,
	}

	hideAPIKeyFields(apiKey)

	is.Nil(apiKey.Digest, "digest should be cleared when hiding api key fields")
}

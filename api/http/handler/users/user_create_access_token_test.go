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
	bolt "github.com/portainer/portainer-ee/api/bolt/bolttest"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/stretchr/testify/assert"
)

func Test_userCreateAccessToken(t *testing.T) {
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

	t.Run("standard user successfully generates API key", func(t *testing.T) {
		data := userAccessTokenCreatePayload{Description: "test-token"}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/tokens", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusCreated, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp accessTokenResponse
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be json")
		is.EqualValues(data.Description, resp.APIKey.Description)
		is.NotEmpty(resp.RawAPIKey)
	})

	t.Run("admin cannot generate API key for standard user", func(t *testing.T) {
		data := userAccessTokenCreatePayload{Description: "test-token-admin"}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/tokens", bytes.NewBuffer(payload))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})

	t.Run("endpoint cannot generate api-key using api-key auth", func(t *testing.T) {
		rawAPIKey, _, err := apiKeyService.GenerateApiKey(*user, "test-api-key")
		is.NoError(err)

		data := userAccessTokenCreatePayload{Description: "test-token-fails"}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/tokens", bytes.NewBuffer(payload))
		req.Header.Add("x-api-key", rawAPIKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusUnauthorized, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
		is.Equal("{\"message\":\"Auth not supported\",\"details\":\"JWT Authentication required\"}\n", string(body))
	})
}

func Test_userAccessTokenCreatePayload(t *testing.T) {
	is := assert.New(t)

	tests := []struct {
		payload    userAccessTokenCreatePayload
		shouldFail bool
	}{
		{
			payload:    userAccessTokenCreatePayload{Description: "test-token"},
			shouldFail: false,
		},
		{
			payload:    userAccessTokenCreatePayload{Description: ""},
			shouldFail: true,
		},
		{
			payload:    userAccessTokenCreatePayload{Description: "test token"},
			shouldFail: false,
		},
		{
			payload:    userAccessTokenCreatePayload{Description: "test-token "},
			shouldFail: false,
		},
		{
			payload: userAccessTokenCreatePayload{Description: `
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
`},
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

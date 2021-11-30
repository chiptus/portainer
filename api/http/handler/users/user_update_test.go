package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/apikey"
	bolt "github.com/portainer/portainer/api/bolt/bolttest"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/jwt"
	"github.com/stretchr/testify/assert"
)

func Test_updateUserRemovesAccessTokens(t *testing.T) {
	is := assert.New(t)

	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	// create standard user
	user := &portainer.User{ID: 2, Username: "standard", Role: portainer.StandardUserRole}
	err := store.User().CreateUser(user)
	is.NoError(err, "error creating user")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService)
	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	authorizationService := authorization.NewService(store)

	h := NewHandler(requestBouncer, rateLimiter, apiKeyService)
	h.DataStore = store
	h.AuthorizationService = authorizationService

	t.Run("standard user deletion removes all associated access tokens", func(t *testing.T) {
		_, _, err := apiKeyService.GenerateApiKey(*user, "test-user-token")
		is.NoError(err)

		keys, err := apiKeyService.GetAPIKeys(user.ID)
		is.NoError(err)
		is.Len(keys, 1)

		rr := httptest.NewRecorder()

		h.deleteUser(rr, user)

		is.Equal(http.StatusNoContent, rr.Code)

		keys, err = apiKeyService.GetAPIKeys(user.ID)
		is.NoError(err)
		is.Equal(0, len(keys))
	})
}

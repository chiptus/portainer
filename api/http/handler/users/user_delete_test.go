package users

import (
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

func Test_deleteUserRemovesAccessTokens(t *testing.T) {
	is := assert.New(t)

	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	// create standard user
	user := &portaineree.User{ID: 2, Username: "standard", Role: portaineree.StandardUserRole}
	err := store.User().CreateUser(user)
	is.NoError(err, "error creating user")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService)
	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	authorizationService := authorization.NewService(store)

	h := NewHandler(requestBouncer, rateLimiter, apiKeyService, testhelpers.NewUserActivityService())
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

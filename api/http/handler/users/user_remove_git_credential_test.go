package users

import (
	"fmt"
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

func Test_userRemoveGitCredential(t *testing.T) {
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

	t.Run("standard user can successfully delete Git Credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-delete-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		newAddedCred, err := store.GitCredentialService.GetGitCredentialByName(user.ID, "test-delete-credential")
		is.NoError(err)
		is.NotNil(newAddedCred)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", "/users/2/gitcredentials", newAddedCred.ID), nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusNoContent, rr.Code)

		creds, err := store.GitCredentialService.GetGitCredentialsByUserID(user.ID)
		is.NoError(err)

		is.Equal(0, len(creds))
	})

	t.Run("admin cannot delete a standard user Git Credential", func(t *testing.T) {
		cred := &portaineree.GitCredential{
			UserID:   user.ID,
			Name:     "test-delete-credential",
			Username: "test-username",
			Password: "test-password",
		}
		err = store.GitCredentialService.Create(cred)
		is.NoError(err)

		newAddedCred, err := store.GitCredentialService.GetGitCredentialByName(user.ID, "test-delete-credential")
		is.NoError(err)
		is.NotNil(newAddedCred)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", "/users/2/gitcredentials", newAddedCred.ID), nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adminJWT))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)

		creds, err := store.GitCredentialService.GetGitCredentialsByUserID(user.ID)
		is.NoError(err)

		is.Equal(1, len(creds))
	})

}

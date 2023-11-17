package system

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/database/models"
	"github.com/segmentio/encoding/json"

	"github.com/stretchr/testify/assert"
)

func Test_getSystemVersion(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	// create version data
	version := &models.Version{SchemaVersion: "2.20.0", Edition: 3}
	err := store.Version().UpdateVersion(version)
	is.NoError(err, "error creating version data")

	// create admin and standard user(s)
	adminUser := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err = store.User().Create(adminUser)
	is.NoError(err, "error creating admin user")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, testhelpers.Licenseservice{}, jwtService, apiKeyService, nil)

	h := NewHandler(requestBouncer, &portainer.Status{}, &demo.Service{}, store, nil)

	// generate standard and admin user tokens
	jwt, _ := jwtService.GenerateToken(&portainer.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})

	t.Run("Display Edition", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/system/version", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp versionResponse
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be list json")

		is.Equal("EE", resp.ServerEdition, "Edition is not expected")
	})
}

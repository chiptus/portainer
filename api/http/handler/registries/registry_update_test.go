package registries

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func ps(s string) *string {
	return &s
}

func pb(b bool) *bool {
	return &b
}

func TestHandler_registryUpdate(t *testing.T) {
	payload := registryUpdatePayload{
		Name:           ps("Updated test registry"),
		URL:            ps("http://example.org/feed"),
		BaseURL:        ps("http://example.org"),
		Authentication: pb(true),
		Username:       ps("username"),
		Password:       ps("password"),
	}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)
	registry := portaineree.Registry{Type: portaineree.ProGetRegistry, ID: 5}
	r := httptest.NewRequest(http.MethodPut, "/registries/5", bytes.NewReader(payloadBytes))
	w := httptest.NewRecorder()

	restrictedContext := &security.RestrictedRequestContext{
		IsAdmin: true,
		UserID:  portaineree.UserID(1),
	}

	ctx := security.StoreRestrictedRequestContext(r, restrictedContext)
	r = r.WithContext(ctx)

	updatedRegistry := portaineree.Registry{}
	handler := NewHandler(helper.NewTestRequestBouncer(), helper.NewUserActivityService())
	handler.DataStore = testDataStore{
		registry: &testRegistryService{
			getRegistry: func(_ portaineree.RegistryID) (*portaineree.Registry, error) {
				return &registry, nil
			},
			updateRegistry: func(ID portaineree.RegistryID, r *portaineree.Registry) error {
				assert.Equal(t, ID, r.ID)
				updatedRegistry = *r
				return nil
			},
		},
	}

	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	// Registry type should remain intact
	assert.Equal(t, registry.Type, updatedRegistry.Type)

	assert.Equal(t, *payload.Name, updatedRegistry.Name)
	assert.Equal(t, *payload.URL, updatedRegistry.URL)
	assert.Equal(t, *payload.BaseURL, updatedRegistry.BaseURL)
	assert.Equal(t, *payload.Authentication, updatedRegistry.Authentication)
	assert.Equal(t, *payload.Username, updatedRegistry.Username)
	assert.Equal(t, *payload.Password, updatedRegistry.Password)
}

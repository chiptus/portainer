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

func Test_registryCreatePayload_Validate(t *testing.T) {
	basePayload := registryCreatePayload{Name: "Test registry", URL: "http://example.com"}
	t.Run("Can't create a ProGet registry if BaseURL is empty", func(t *testing.T) {
		payload := basePayload
		payload.Type = portaineree.ProGetRegistry
		err := payload.Validate(nil)
		assert.Error(t, err)
	})
	t.Run("Can create a GitLab registry if BaseURL is empty", func(t *testing.T) {
		payload := basePayload
		payload.Type = portaineree.GitlabRegistry
		err := payload.Validate(nil)
		assert.NoError(t, err)
	})
	t.Run("Can create a ProGet registry if BaseURL is not empty", func(t *testing.T) {
		payload := basePayload
		payload.Type = portaineree.ProGetRegistry
		payload.BaseURL = "http://example.com"
		err := payload.Validate(nil)
		assert.NoError(t, err)
	})
	t.Run("Can't create a AWS ECR registry if authentication required, but access key ID, secret access key or region is empty", func(t *testing.T) {
		payload := basePayload
		payload.Type = portaineree.EcrRegistry
		payload.Authentication = true
		err := payload.Validate(nil)
		assert.Error(t, err)
	})
	t.Run("Do not require access key ID, secret access key, region for public AWS ECR registry", func(t *testing.T) {
		payload := basePayload
		payload.Type = portaineree.EcrRegistry
		payload.Authentication = false
		err := payload.Validate(nil)
		assert.NoError(t, err)
	})
}

type testRegistryService struct {
	portaineree.RegistryService
	createRegistry func(r *portaineree.Registry) error
	updateRegistry func(ID portaineree.RegistryID, r *portaineree.Registry) error
	getRegistry    func(ID portaineree.RegistryID) (*portaineree.Registry, error)
}

type testDataStore struct {
	portaineree.DataStore
	registry *testRegistryService
}

func (t testDataStore) Registry() portaineree.RegistryService {
	return t.registry
}

func (t testRegistryService) CreateRegistry(r *portaineree.Registry) error {
	return t.createRegistry(r)
}

func (t testRegistryService) UpdateRegistry(ID portaineree.RegistryID, r *portaineree.Registry) error {
	return t.updateRegistry(ID, r)
}

func (t testRegistryService) Registry(ID portaineree.RegistryID) (*portaineree.Registry, error) {
	return t.getRegistry(ID)
}

func (t testRegistryService) Registries() ([]portaineree.Registry, error) {
	return nil, nil
}

func TestHandler_registryCreate(t *testing.T) {
	payload := registryCreatePayload{
		Name:           "Test registry",
		Type:           portaineree.ProGetRegistry,
		URL:            "http://example.com",
		BaseURL:        "http://example.com",
		Authentication: false,
		Username:       "username",
		Password:       "password",
		Gitlab:         portaineree.GitlabRegistryData{},
	}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payloadBytes))
	w := httptest.NewRecorder()

	restrictedContext := &security.RestrictedRequestContext{
		IsAdmin: true,
		UserID:  portaineree.UserID(1),
	}

	ctx := security.StoreRestrictedRequestContext(r, restrictedContext)
	r = r.WithContext(ctx)

	registry := portaineree.Registry{}
	handler := NewHandler(helper.NewTestRequestBouncer(), helper.NewUserActivityService())
	handler.DataStore = testDataStore{
		registry: &testRegistryService{
			createRegistry: func(r *portaineree.Registry) error {
				registry = *r
				return nil
			},
		},
	}
	handlerError := handler.registryCreate(w, r)
	assert.Nil(t, handlerError)
	assert.Equal(t, payload.Name, registry.Name)
	assert.Equal(t, payload.Type, registry.Type)
	assert.Equal(t, payload.URL, registry.URL)
	assert.Equal(t, payload.BaseURL, registry.BaseURL)
	assert.Equal(t, payload.Authentication, registry.Authentication)
	assert.Equal(t, payload.Username, registry.Username)
	assert.Equal(t, payload.Password, registry.Password)
}

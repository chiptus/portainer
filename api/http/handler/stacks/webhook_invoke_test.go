package stacks

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/stretchr/testify/assert"
)

func TestHandler_webhookInvoke(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()
	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")
	webhookID := newGuidString(t)
	store.StackService.Create(&portaineree.Stack{
		ID: 1,
		AutoUpdate: &portaineree.StackAutoUpdate{
			Webhook: webhookID,
		},
		CreatedBy: "admin",
		Type:      portaineree.DockerComposeStack,
	})

	h := NewHandler(nil, store, nil)

	t.Run("invalid uuid results in http.StatusBadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest("notuuid")
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("registered webhook ID in http.StatusNoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(webhookID)
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("unregistered webhook ID in http.StatusNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(newGuidString(t))
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func Test_parseQuery(t *testing.T) {
	t.Run("no query params returns no vars", func(t *testing.T) {
		query, _ := url.ParseQuery("")
		pullimage, envs := parseQuery(query)
		assert.Equal(t, nilptr[bool](), pullimage)
		assert.Equal(t, []portaineree.Pair{}, envs)
	})
	t.Run("return query params as pairs", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2")
		pullimage, envs := parseQuery(query)
		assert.Equal(t, nilptr[bool](), pullimage)
		assert.ElementsMatch(t, []portaineree.Pair{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "2"},
		}, envs)
	})
	t.Run("take last value if param is duplicated", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2&second=3")
		pullimage, envs := parseQuery(query)
		assert.Equal(t, nilptr[bool](), pullimage)
		assert.ElementsMatch(t, []portaineree.Pair{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "3"},
		}, envs)
	})
	t.Run("pullimage is a special param", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2&pullimage=true")
		pullimage, envs := parseQuery(query)
		assert.Equal(t, true, *pullimage)
		assert.ElementsMatch(t, []portaineree.Pair{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "2"},
		}, envs)
	})
	t.Run("pullimage is nil if value is incorrect bool", func(t *testing.T) {
		query, _ := url.ParseQuery("pullimage=aye")
		pullimage, envs := parseQuery(query)
		assert.Equal(t, nilptr[bool](), pullimage)
		assert.Equal(t, []portaineree.Pair{}, envs)
	})
}

func nilptr[T any]() *T {
	return nil
}

func newGuidString(t *testing.T) string {
	uuid, err := uuid.NewV4()
	assert.NoError(t, err)

	return uuid.String()
}

func newRequest(webhookID string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/stacks/webhooks/"+webhookID, nil)
}

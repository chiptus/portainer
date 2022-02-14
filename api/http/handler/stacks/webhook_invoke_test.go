package stacks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/bolttest"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestHandler_webhookInvoke(t *testing.T) {
	store, teardown := bolttest.MustNewTestStore(true)
	defer teardown()
	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().CreateUser(admin)
	err = store.Endpoint().CreateEndpoint(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")
	webhookID := newGuidString(t)
	store.StackService.CreateStack(&portaineree.Stack{
		ID: 1,
		AutoUpdate: &portaineree.StackAutoUpdate{
			Webhook: webhookID,
		},
		CreatedBy: "admin",
		Type:      portaineree.DockerComposeStack,
	})

	h := NewHandler(nil, store, helper.NewUserActivityService())

	t.Run("invalid uuid results in http.StatusBadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest("notuuid")
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("registered webhook ID in http.StatusNoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(webhookID)
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("unregistered webhook ID in http.StatusNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(newGuidString(t))
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func newGuidString(t *testing.T) string {
	uuid, err := uuid.NewV4()
	assert.NoError(t, err)

	return uuid.String()
}

func newRequest(webhookID string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/stacks/webhooks/"+webhookID, nil)
}

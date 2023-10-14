package edgestacks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func TestHandler_webhookInvoke(t *testing.T) {
	h, _ := setupHandler(t)

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := h.DataStore.User().Create(admin)
	assert.NoError(t, err, "error creating user")

	err = h.DataStore.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")

	webhookID := newGuidString(t)
	err = h.DataStore.EdgeStack().Create(1, &portaineree.EdgeStack{
		ID: 1,
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook: webhookID,
		},
	})
	assert.NoError(t, err, "error creating edge stack")

	// t.Run("invalid uuid results in http.StatusBadRequest", func(t *testing.T) {
	// 	w := httptest.NewRecorder()
	// 	req := newRequest("notuuid")
	// 	h.Router.ServeHTTP(w, req)
	// 	assert.Equal(t, http.StatusBadRequest, w.Code)
	// })
	t.Run("registered webhook ID in http.StatusNoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(webhookID)
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	// t.Run("unregistered webhook ID in http.StatusNotFound", func(t *testing.T) {
	// 	w := httptest.NewRecorder()
	// 	req := newRequest(newGuidString(t))
	// 	h.Router.ServeHTTP(w, req)
	// 	assert.Equal(t, http.StatusNotFound, w.Code)
	// })
}

func newGuidString(t *testing.T) string {
	uuid, err := uuid.NewV4()
	assert.NoError(t, err)

	return uuid.String()
}

func newRequest(webhookID string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/edge_stacks/webhooks/"+webhookID, nil)
}

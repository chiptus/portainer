package stacks

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandler_webhookInvoke(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating admin user")

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
		AutoUpdate: &portaineree.AutoUpdateSettings{
			Webhook: webhookID,
		},
		CreatedBy: "admin",
		Type:      portaineree.DockerComposeStack,
	})

	h := NewHandler(testhelpers.NewTestRequestBouncer(), store, nil)

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
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, nilptr[bool](), options.PullDockerImage)
		assert.Equal(t, []portaineree.Pair{}, options.AdditionalEnvVars)
	})

	t.Run("return query params as pairs", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2")
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, nilptr[bool](), options.PullDockerImage)
		assert.ElementsMatch(t, []portaineree.Pair{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "2"},
		}, options.AdditionalEnvVars)
	})

	t.Run("take last value if param is duplicated", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2&second=3")
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, nilptr[bool](), options.PullDockerImage)
		assert.ElementsMatch(t, []portaineree.Pair{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "3"},
		}, options.AdditionalEnvVars)
	})

	t.Run("pullimage is a special param", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2&pullimage=true")
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, true, *options.PullDockerImage)
		assert.ElementsMatch(t, []portaineree.Pair{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "2"},
		}, options.AdditionalEnvVars)
	})

	t.Run("pullimage is a not a bool", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2&pullimage=notbool")
		_, err := parseQuery(query)
		assert.Error(t, err)
	})

	t.Run("pullimage has no value", func(t *testing.T) {
		query, _ := url.ParseQuery("first=1&second=2&pullimage")
		_, err := parseQuery(query)
		assert.Error(t, err)
	})

	t.Run("RolloutRestartK8sAll is true if rollout-restart=all", func(t *testing.T) {
		query, _ := url.ParseQuery("rollout-restart=all")
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, true, options.RolloutRestartK8sAll)
		assert.Equal(t, []string(nil), options.RolloutRestartK8sResourceList)
	})

	t.Run("RolloutRestartK8sResourceList is correct for single resource", func(t *testing.T) {
		query, _ := url.ParseQuery("rollout-restart=deployment/nginx")
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, false, options.RolloutRestartK8sAll)
		assert.ElementsMatch(t, []string{"deployment/nginx"}, options.RolloutRestartK8sResourceList)
	})

	t.Run("RolloutRestartK8sResourceList is correct for multiple resources", func(t *testing.T) {
		query, _ := url.ParseQuery("rollout-restart=deployment/nginx,daemonset/helper,statefulset/redis")
		options, err := parseQuery(query)
		assert.NoError(t, err)
		assert.Equal(t, false, options.RolloutRestartK8sAll)
		assert.ElementsMatch(t, []string{
			"deployment/nginx",
			"daemonset/helper",
			"statefulset/redis",
		}, options.RolloutRestartK8sResourceList)
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

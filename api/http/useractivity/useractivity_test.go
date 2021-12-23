package useractivity

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	i "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_CanHandleRequestsWithoutBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(nil))
	uas := userActivityService{}

	var passedThroughBody []byte
	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThroughBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, []byte{}, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "",
		loggedContext:  "Portainer",
		loggedAction:   "POST /",
		loggedPayload:  []byte(nil),
	}, uas)
}
func Test_OnlyLogWriteRequests(t *testing.T) {
	payload := []byte{1, 2, 3}
	r := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer(payload))
	uas := userActivityService{}

	var passedThroughBody []byte
	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThroughBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, payload, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         false,
		loggedUsername: "",
		loggedContext:  "",
		loggedAction:   "",
		loggedPayload:  nil,
	}, uas)
}

func Test_LogSkipUnknownPayloads(t *testing.T) {
	payload := []byte(`{"a":1,"password":1}`)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(payload))
	uas := userActivityService{}

	var passedThroughBody []byte
	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThroughBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, payload, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "",
		loggedContext:  "Portainer",
		loggedAction:   "POST /",
		loggedPayload:  []byte(nil),
	}, uas)
}

func Test_LogSanitisedJsonPayloads(t *testing.T) {
	payload := []byte(`{"a":1,"password":1}`)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(payload))
	r.Header.Add("Content-Type", "application/json")
	uas := userActivityService{}

	var passedThroughBody []byte
	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThroughBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, payload, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "",
		loggedContext:  "Portainer",
		loggedAction:   "POST /",
		loggedPayload:  []byte(`{"a":1,"password":"[REDACTED]"}`),
	}, uas)
}

func Test_ClearTarPayloads(t *testing.T) {
	payload := []byte(`{"a":1,"password":1}`)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(payload))
	r.Header.Add("Content-Type", "application/x-tar")
	uas := userActivityService{}

	var passedThroughBody []byte
	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThroughBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, payload, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "",
		loggedContext:  "Portainer",
		loggedAction:   "POST /",
		loggedPayload:  []byte(nil),
	}, uas)
}

func Test_LogSanitisedAndSkipFiles_InMultipartPayloads(t *testing.T) {
	payload := `--xxx
Content-Disposition: form-data; name="field1"

value1
--xxx
Content-Disposition: form-data; name="field2"

value2
--xxx
Content-Disposition: form-data; name="file"; filename="file"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
`
	r := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(payload)))
	r.Header.Add("Content-Type", `multipart/form-data; boundary=xxx`)
	uas := userActivityService{}

	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "",
		loggedContext:  "Portainer",
		loggedAction:   "POST /",
		loggedPayload:  []byte(`{"field1":"value1","field2":"value2"}`),
	}, uas)
}

func Test_PicksUsernameFromRequestSecurityToken(t *testing.T) {
	payload := []byte(``)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(payload))
	r = r.WithContext(security.StoreTokenData(r, &portaineree.TokenData{Username: "superuser"}))

	uas := userActivityService{}

	var passedThroughBody []byte
	LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThroughBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, payload, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "superuser",
		loggedContext:  "Portainer",
		loggedAction:   "POST /",
		loggedPayload:  []byte(nil),
	}, uas)
}

func Test_PicksEnpointFromRequestContext(t *testing.T) {
	payload := []byte(``)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(payload))
	r = mux.SetURLVars(r, map[string]string{"endpointId": "1"})

	uas := userActivityService{}
	ds := i.NewDatastore(i.WithEndpoints([]portaineree.Endpoint{{ID: portaineree.EndpointID(1), Name: "MyEndpoint"}}))

	var passedThroughBody []byte
	middlewares.WithEndpoint(ds.Endpoint(), "endpointId")(
		LogUserActivity(&uas)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			passedThroughBody, _ = ioutil.ReadAll(r.Body)
			r.Body.Close()
			w.WriteHeader(http.StatusOK)
		}))).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, payload, passedThroughBody)
	assert.Equal(t, userActivityService{
		called:         true,
		loggedUsername: "",
		loggedContext:  "MyEndpoint",
		loggedAction:   "POST /",
		loggedPayload:  []byte(nil),
	}, uas)
}

type userActivityService struct {
	loggedUsername string
	loggedContext  string
	loggedAction   string
	loggedPayload  []byte
	called         bool
}

func (service *userActivityService) LogAuthActivity(username string, origin string, context portaineree.AuthenticationMethod, activityType portaineree.AuthenticationActivityType) error {
	return nil
}

func (service *userActivityService) LogUserActivity(username string, context string, action string, payload []byte) error {
	service.called = true
	service.loggedUsername = username
	service.loggedContext = context
	service.loggedAction = action
	service.loggedPayload = payload
	return nil
}

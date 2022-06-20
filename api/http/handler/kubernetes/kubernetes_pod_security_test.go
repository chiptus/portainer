package kubernetes

import (
	"bytes"
	"encoding/json"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/exec/exectest"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	"github.com/portainer/portainer/api/filesystem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/stretchr/testify/assert"
)

type endpointListEdgeDeviceTest struct {
	title    string
	expected []portaineree.EndpointID
	filter   string
}

func Test_getK8sPodSecurityRule(t *testing.T) {
	is := assert.New(t)

	_, store, teardown := datastore.MustNewTestStore(true, true)
	defer teardown()

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1, Type: portaineree.AgentOnKubernetesEnvironment})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	tk, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: 1, Username: "admin", Role: portaineree.AdministratorRole})

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)
	kubernetesDeployer := exectest.NewKubernetesDeployer()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	authorizationService := authorization.NewService(store)
	handler := NewHandler(requestBouncer, authorizationService, store, jwtService, kubeClusterAccessService,
		nil, testhelpers.NewUserActivityService(),
		kubernetesDeployer, nil, "./")
	is.NotNil(handler, "Handler should not fail")

	req := httptest.NewRequest(http.MethodGet, "/kubernetes/1/opa", nil)
	ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+tk)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 403")

}
func Test_updateK8sPodSecurityRule(t *testing.T) {
	is := assert.New(t)

	_, store, teardown := datastore.MustNewTestStore(true, true)
	defer teardown()

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1, Type: portaineree.AgentOnKubernetesEnvironment})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	tk, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: 1, Username: "admin", Role: portaineree.AdministratorRole})

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)
	kubernetesDeployer := exectest.NewKubernetesDeployer()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	authorizationService := authorization.NewService(store)
	handler := NewHandler(requestBouncer, authorizationService, store, jwtService, kubeClusterAccessService,
		nil, testhelpers.NewUserActivityService(),
		kubernetesDeployer, nil, "./")
	is.NotNil(handler, "Handler should not fail")

	req := httptest.NewRequest(http.MethodGet, "/kubernetes/1/opa", nil)
	ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+tk)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 403")
}

func TestHandler_updateK8sPodSecurityRule(t *testing.T) {
	_mockCheckGatekeeper := checkGetekeeperStatus
	defer func() {
		checkGetekeeperStatus = _mockCheckGatekeeper
	}()
	checkGetekeeperStatus = func(handler *Handler, endpoint *portaineree.Endpoint, r *http.Request) error {
		return nil
	}
	is := assert.New(t)

	_, store, teardown := datastore.MustNewTestStore(true, true)
	defer teardown()

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1, Type: portaineree.AgentOnKubernetesEnvironment})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	tk, _ := jwtService.GenerateToken(&portaineree.TokenData{ID: 1, Username: "admin", Role: portaineree.AdministratorRole})

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)
	kubernetesDeployer := exectest.NewKubernetesDeployer()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	authorizationService := authorization.NewService(store)

	tmpDir, err := os.MkdirTemp(os.TempDir(), "portainer-test-global-key-*")
	if err != nil {
		teardown()
	}
	fs, err := filesystem.NewService(tmpDir, "")
	if err != nil {
		teardown()
	}
	handler := NewHandler(requestBouncer, authorizationService, store, jwtService, kubeClusterAccessService,
		nil, testhelpers.NewUserActivityService(),
		kubernetesDeployer, fs, "./")
	is.NotNil(handler, "Handler should not fail")
	payload := &podsecurity.PodSecurityRule{}
	payload.Enabled = true
	optdata, err := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/kubernetes/1/opa", bytes.NewBuffer(optdata))
	ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+tk)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 200")
}

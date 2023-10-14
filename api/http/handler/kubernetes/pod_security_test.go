package kubernetes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/exec/exectest"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/stretchr/testify/assert"
)

func Test_getK8sPodSecurityRule(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1, Type: portaineree.AgentOnKubernetesEnvironment})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	tk, _ := jwtService.GenerateToken(&portainer.TokenData{ID: 1, Username: "admin", Role: portaineree.AdministratorRole})

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)
	kubernetesDeployer := exectest.NewKubernetesDeployer()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	authorizationService := authorization.NewService(store)
	handler := NewHandler(requestBouncer, authorizationService, store, jwtService, kubeClusterAccessService,
		nil, nil, testhelpers.NewUserActivityService(),
		kubernetesDeployer, nil, "./")
	is.NotNil(handler, "Handler should not fail")

	req := httptest.NewRequest(http.MethodGet, "/kubernetes/1/opa", nil)
	ctx := security.StoreTokenData(req, &portainer.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+tk)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 403")

}
func Test_updateK8sPodSecurityRule(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1, Type: portaineree.AgentOnKubernetesEnvironment})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	tk, _ := jwtService.GenerateToken(&portainer.TokenData{ID: 1, Username: "admin", Role: portaineree.AdministratorRole})

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)
	kubernetesDeployer := exectest.NewKubernetesDeployer()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	authorizationService := authorization.NewService(store)
	handler := NewHandler(requestBouncer, authorizationService, store, jwtService, kubeClusterAccessService,
		nil, nil, testhelpers.NewUserActivityService(),
		kubernetesDeployer, nil, "./")
	is.NotNil(handler, "Handler should not fail")

	req := httptest.NewRequest(http.MethodGet, "/kubernetes/1/opa", nil)
	ctx := security.StoreTokenData(req, &portainer.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+tk)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 403")
}

func TestHandler_updateK8sPodSecurityRule(t *testing.T) {
	_mockCheckGatekeeper := checkGatekeeperStatus
	defer func() {
		checkGatekeeperStatus = _mockCheckGatekeeper
	}()
	checkGatekeeperStatus = func(handler *Handler, endpoint *portaineree.Endpoint, r *http.Request) error {
		return nil
	}
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1, Type: portaineree.AgentOnKubernetesEnvironment})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	tk, _ := jwtService.GenerateToken(&portainer.TokenData{ID: 1, Username: "admin", Role: portaineree.AdministratorRole})

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)
	kubernetesDeployer := exectest.NewKubernetesDeployer()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	authorizationService := authorization.NewService(store)

	fs, err := filesystem.NewService(t.TempDir(), "")
	if err != nil {
		t.Fatalf("unable to create filesystem service: %s", err)
	}

	handler := NewHandler(requestBouncer, authorizationService, store, jwtService, kubeClusterAccessService,
		nil, nil, testhelpers.NewUserActivityService(),
		kubernetesDeployer, fs, "./")
	is.NotNil(handler, "Handler should not fail")
	payload := &podsecurity.PodSecurityRule{}
	payload.Enabled = true

	optdata, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unable to marshal payload: %s", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/kubernetes/1/opa", bytes.NewBuffer(optdata))
	ctx := security.StoreTokenData(req, &portainer.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+tk)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 200")
}

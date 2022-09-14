package helm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/portainer/libhelm/binary/test"
	"github.com/portainer/libhelm/options"
	"github.com/portainer/libhelm/release"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/exec/exectest"
	"github.com/portainer/portainer-ee/api/http/security"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/stretchr/testify/assert"
)

func Test_helmList(t *testing.T) {
	is := assert.New(t)

	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1})
	is.NoError(err, "error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initialising jwt service")

	kubernetesDeployer := exectest.NewKubernetesDeployer()
	helmPackageManager := test.NewMockHelmBinaryPackageManager("")
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	h := NewHandler(helper.NewTestRequestBouncer(), store, jwtService, kubernetesDeployer, helmPackageManager, kubeClusterAccessService, helper.NewUserActivityService())

	// Install a single chart.  We expect to get these values back
	options := options.InstallOptions{Name: "nginx-1", Chart: "nginx", Namespace: "default"}
	h.helmPackageManager.Install(options)

	t.Run("helmList", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/1/kubernetes/helm", nil)
		ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "admin", Role: 1})
		req = req.WithContext(ctx)
		req.Header.Add("Authorization", "Bearer dummytoken")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code, "Status should be 200 OK")

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		data := []release.ReleaseElement{}
		json.Unmarshal(body, &data)
		if is.Equal(1, len(data), "Expected one chart entry") {
			is.EqualValues(options.Name, data[0].Name, "Name doesn't match")
			is.EqualValues(options.Chart, data[0].Chart, "Chart doesn't match")
		}
	})
}

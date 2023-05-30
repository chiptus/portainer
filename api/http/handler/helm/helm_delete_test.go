package helm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/exec/exectest"
	"github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/http/security"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/pkg/libhelm/binary/test"
	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/stretchr/testify/assert"
)

func Test_helmDelete(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1})
	is.NoError(err, "Error creating environment")

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "Error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")

	kubernetesDeployer := exectest.NewKubernetesDeployer()
	helmPackageManager := test.NewMockHelmBinaryPackageManager("")
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	tmp := t.TempDir()
	fileService, _ := filesystem.NewService(tmp, tmp)
	h := NewHandler(helper.NewTestRequestBouncer(), store, jwtService, kubernetesDeployer, helmPackageManager, kubeClusterAccessService, helper.NewUserActivityService(), fileService)

	is.NotNil(h, "Handler should not fail")

	// Install a single chart directly, to be deleted by the handler
	options := options.InstallOptions{Name: "nginx-1", Chart: "nginx", Namespace: "default"}
	h.helmPackageManager.Install(options)

	t.Run("helmDelete fails for unauthorized user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/1/kubernetes/helm/%s", options.Name), nil)
		ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "non-admin", Role: 2})
		req = req.WithContext(ctx)
		req.Header.Add("Authorization", "Bearer dummytoken")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code, "Status should be 403")
	})

	t.Run("helmDelete succeeds with admin user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/1/kubernetes/helm/%s", options.Name), nil)
		ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "admin", Role: 1})
		req = req.WithContext(ctx)
		req.Header.Add("Authorization", "Bearer dummytoken")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusNoContent, rr.Code, "Status should be 204")
	})
}

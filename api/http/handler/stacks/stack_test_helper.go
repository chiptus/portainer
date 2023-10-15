package stacks

import (
	"io"
	"net/http"
	"net/http/httptest"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
)

func mockCreateUser(store *datastore.Store) (*portaineree.User, error) {
	user := &portaineree.User{ID: 1, Username: "testUser", Role: portaineree.AdministratorRole, PortainerAuthorizations: authorization.DefaultPortainerAuthorizations()}
	err := store.User().Create(user)
	return user, err
}

func mockCreateEndpoint(store *datastore.Store) (*portaineree.Endpoint, error) {
	endpoint := &portaineree.Endpoint{
		ID:   1,
		Name: "testEndpoint",
		SecuritySettings: portainer.EndpointSecuritySettings{
			AllowBindMountsForRegularUsers:            true,
			AllowPrivilegedModeForRegularUsers:        true,
			AllowVolumeBrowserForRegularUsers:         true,
			AllowHostNamespaceForRegularUsers:         true,
			AllowDeviceMappingForRegularUsers:         true,
			AllowStackManagementForRegularUsers:       true,
			AllowContainerCapabilitiesForRegularUsers: true,
			AllowSysctlSettingForRegularUsers:         true,
			EnableHostManagementFeatures:              true,
		},
	}

	err := store.Endpoint().Create(endpoint)
	return endpoint, err
}

func mockCreateStackRequestWithSecurityContext(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method,
		target,
		body)

	ctx := security.StoreRestrictedRequestContext(req, &security.RestrictedRequestContext{
		IsAdmin: true,
		UserID:  portainer.UserID(1),
	})

	return req.WithContext(ctx)
}

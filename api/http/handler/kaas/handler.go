package kaas

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/license"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	*mux.Router
	dataStore              dataservices.DataStore
	cloudManagementService *cloud.CloudManagementService
	cloudInfoService       *cloud.CloudInfoService
	licenseService         portaineree.LicenseService
	requestBouncer         security.BouncerService
	userActivityService    portaineree.UserActivityService
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(
	bouncer security.BouncerService,
	dataStore dataservices.DataStore,
	cloudManagementService *cloud.CloudManagementService,
	cloudInfoService *cloud.CloudInfoService,
	userActivityService portaineree.UserActivityService,
	licenseService portaineree.LicenseService,
) *Handler {
	h := &Handler{
		Router:                 mux.NewRouter(),
		dataStore:              dataStore,
		cloudManagementService: cloudManagementService,
		cloudInfoService:       cloudInfoService,
		userActivityService:    userActivityService,
		licenseService:         licenseService,
		requestBouncer:         bouncer,
	}

	endpointRouter := h.NewRoute().Subrouter()

	endpointRouter.Use(bouncer.AuthenticatedAccess, middlewares.WithEndpoint(dataStore.Endpoint(), "endpointid"))

	// requires node write authorization: OperationK8sClusterNodeW
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/nodes/remove", httperror.LoggerHandler(h.removeNodes)).Methods(http.MethodPost)
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/nodes/add", httperror.LoggerHandler(h.addNodes)).Methods(http.MethodPost)
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/nodes/nodestatus", httperror.LoggerHandler(h.microk8sGetNodeStatus)).Methods(http.MethodGet)
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/upgrade", httperror.LoggerHandler(h.upgrade)).Methods(http.MethodPost)
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/testssh", license.NotOverused(licenseService, dataStore, httperror.LoggerHandler(h.sshTestNodeIPs))).Methods(http.MethodPost)

	endpointRouter.Handle("/cloud/endpoints/{endpointid}/version", httperror.LoggerHandler(h.version)).Methods(http.MethodGet)

	// microk8s only
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/addons", httperror.LoggerHandler(h.microk8sGetAddons)).Methods(http.MethodGet)
	endpointRouter.Handle("/cloud/endpoints/{endpointid}/addons", httperror.LoggerHandler(h.microk8sUpdateAddons)).Methods(http.MethodPost)

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)
	authenticatedRouter.Handle("/cloud/{provider}/info", httperror.LoggerHandler(h.providerInfo)).Methods(http.MethodGet)

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess)

	loggedAdminRouter := h.NewRoute().Subrouter()
	loggedAdminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))
	loggedAdminRouter.Handle("/cloud/{provider}/provision", license.NotOverused(licenseService, dataStore, httperror.LoggerHandler(h.provisionCluster))).Methods(http.MethodPost)
	loggedAdminRouter.Handle("/cloud/testssh", license.NotOverused(licenseService, dataStore, httperror.LoggerHandler(h.sshTestNodeIPs))).Methods(http.MethodPost)

	return h
}

// Check if the user is an admin or can write to the cluster node
func canWriteK8sClusterNode(user *portaineree.User, endpointID portaineree.EndpointID) bool {
	isAdmin := user.Role == portaineree.AdministratorRole
	hasAccess := false
	if user.EndpointAuthorizations[portaineree.EndpointID(endpointID)] != nil {
		_, hasAccess = user.EndpointAuthorizations[portaineree.EndpointID(endpointID)][portaineree.OperationK8sClusterNodeW]
	}
	return isAdmin || hasAccess
}

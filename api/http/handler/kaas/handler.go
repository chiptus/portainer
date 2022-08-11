package kaas

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/license"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	*mux.Router
	dataStore                dataservices.DataStore
	cloudClusterSetupService *cloud.CloudClusterSetupService
	cloudClusterInfoService  *cloud.CloudClusterInfoService
	userActivityService      portaineree.UserActivityService
	licenseService           portaineree.LicenseService
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(
	bouncer *security.RequestBouncer,
	dataStore dataservices.DataStore,
	cloudClusterSetupService *cloud.CloudClusterSetupService,
	cloudClusterInfoService *cloud.CloudClusterInfoService,
	userActivityService portaineree.UserActivityService,
	licenseService portaineree.LicenseService,
) *Handler {
	h := &Handler{
		Router:                   mux.NewRouter(),
		dataStore:                dataStore,
		cloudClusterSetupService: cloudClusterSetupService,
		cloudClusterInfoService:  cloudClusterInfoService,
		userActivityService:      userActivityService,
		licenseService:           licenseService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/cloud/{provider}/info", httperror.LoggerHandler(h.kaasProviderInfo)).Methods(http.MethodGet)
	adminRouter.Handle("/cloud/{provider}/cluster", license.NotOverused(licenseService, dataStore, httperror.LoggerHandler(h.provisionKaaSCluster))).Methods(http.MethodPost)

	return h
}

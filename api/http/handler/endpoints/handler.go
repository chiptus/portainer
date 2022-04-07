package endpoints

import (
	"net/http"
	"time"

	werrors "github.com/pkg/errors"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
)

func hideFields(endpoint *portaineree.Endpoint) {
	endpoint.AzureCredentials = portaineree.AzureCredentials{}
	if len(endpoint.Snapshots) > 0 {
		endpoint.Snapshots[0].SnapshotRaw = portainer.DockerSnapshotRaw{}
	}
}

// This requestBouncer exists because security.RequestBounder is a type and not an interface.
// Therefore we can not swit	 it out with a dummy bouncer for go tests.  This interface works around it
type requestBouncer interface {
	AuthenticatedAccess(h http.Handler) http.Handler
	AdminAccess(h http.Handler) http.Handler
	PublicAccess(h http.Handler) http.Handler
	AuthorizedEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint, authorizationCheck bool) error
	AuthorizedEdgeEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint) error
}

// Handler is the HTTP handler used to handle environment(endpoint) operations.
type Handler struct {
	*mux.Router
	requestBouncer       requestBouncer
	AuthorizationService *authorization.Service
	dataStore            dataservices.DataStore
	FileService          portainer.FileService
	ProxyManager         *proxy.Manager
	ReverseTunnelService portaineree.ReverseTunnelService
	SnapshotService      portaineree.SnapshotService
	K8sClientFactory     *cli.ClientFactory
	ComposeStackManager  portaineree.ComposeStackManager
	DockerClientFactory  *docker.ClientFactory
	BindAddress          string
	BindAddressHTTPS     string
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage environment(endpoint) operations.
func NewHandler(bouncer requestBouncer, userActivityService portaineree.UserActivityService, dataStore dataservices.DataStore) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		requestBouncer:      bouncer,
		userActivityService: userActivityService,
		dataStore:           dataStore,
	}

	logEndpointActivity := useractivity.LogUserActivityWithContext(h.userActivityService, middlewares.FindInPath(dataStore.Endpoint(), "id"))

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, logEndpointActivity)

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, logEndpointActivity)

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)

	adminRouter.Handle("/endpoints", httperror.LoggerHandler(h.endpointCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/endpoints/snapshot", httperror.LoggerHandler(h.endpointSnapshots)).Methods(http.MethodPost)
	adminRouter.Handle("/endpoints", httperror.LoggerHandler(h.endpointList)).Methods(http.MethodGet)
	adminRouter.Handle("/endpoints/{id}", httperror.LoggerHandler(h.endpointInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/endpoints/{id}", httperror.LoggerHandler(h.endpointDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/endpoints/{id}/association", httperror.LoggerHandler(h.endpointAssociationDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/endpoints/{id}/snapshot", httperror.LoggerHandler(h.endpointSnapshot)).Methods(http.MethodPost)

	authenticatedRouter.Handle("/endpoints/{id}", httperror.LoggerHandler(h.endpointUpdate)).Methods(http.MethodPut)
	authenticatedRouter.Handle("/endpoints/{id}/settings", httperror.LoggerHandler(h.endpointSettingsUpdate)).Methods(http.MethodPut)
	authenticatedRouter.Handle("/endpoints/{id}/dockerhub/{registryId}", httperror.LoggerHandler(h.endpointDockerhubStatus)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/endpoints/{id}/pools/{rpn}/access", httperror.LoggerHandler(h.endpointPoolsAccessUpdate)).Methods(http.MethodPut)
	authenticatedRouter.Handle("/endpoints/{id}/forceupdateservice", httperror.LoggerHandler(h.endpointForceUpdateService)).Methods(http.MethodPut)

	authenticatedRouter.Handle("/endpoints/{id}/registries", httperror.LoggerHandler(h.endpointRegistriesList)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/endpoints/{id}/registries/{registryId}", httperror.LoggerHandler(h.endpointRegistryAccess)).Methods(http.MethodPut)

	// DEPRECATED
	publicRouter.Handle("/endpoints/{id}/status", httperror.LoggerHandler(h.endpointStatusInspect)).Methods(http.MethodGet)

	return h
}

func validateAutoUpdateSettings(autoUpdateWindow portaineree.EndpointChangeWindow) error {
	if !autoUpdateWindow.Enabled {
		return nil
	}

	if !validTime24(autoUpdateWindow.StartTime) {
		return werrors.New("AutoUpdateWindow.StartTime: invalid time format, expected HH:MM")
	}

	if !validTime24(autoUpdateWindow.EndTime) {
		return werrors.New("AutoUpdateWindow.EndTime: invalid time format, expected HH:MM")
	}

	return nil
}

// Return true if the time string specified is a valid 24hr time. e.g. 22:30
func validTime24(ts string) bool {
	_, err := time.Parse(portaineree.TimeFormat24, ts)
	return err == nil
}

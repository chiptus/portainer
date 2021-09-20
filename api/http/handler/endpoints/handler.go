package endpoints

import (
	"net/http"

	"github.com/portainer/portainer/api/docker"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

func hideFields(endpoint *portainer.Endpoint) {
	endpoint.AzureCredentials = portainer.AzureCredentials{}
	if len(endpoint.Snapshots) > 0 {
		endpoint.Snapshots[0].SnapshotRaw = portainer.DockerSnapshotRaw{}
	}
}

// Handler is the HTTP handler used to handle environment(endpoint) operations.
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	AuthorizationService *authorization.Service
	DataStore            portainer.DataStore
	FileService          portainer.FileService
	ProxyManager         *proxy.Manager
	ReverseTunnelService portainer.ReverseTunnelService
	SnapshotService      portainer.SnapshotService
	K8sClientFactory     *cli.ClientFactory
	ComposeStackManager  portainer.ComposeStackManager
	DockerClientFactory  *docker.ClientFactory
	UserActivityStore    portainer.UserActivityStore
	BindAddress          string
	BindAddressHTTPS     string
}

// NewHandler creates a handler to manage environment(endpoint) operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}

	h.Handle("/endpoints",
		bouncer.AdminAccess(httperror.LoggerHandler(h.endpointCreate))).Methods(http.MethodPost)
	h.Handle("/endpoints/{id}/settings",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointSettingsUpdate))).Methods(http.MethodPut)
	h.Handle("/endpoints/{id}/association",
		bouncer.AdminAccess(httperror.LoggerHandler(h.endpointAssociationDelete))).Methods(http.MethodDelete)
	h.Handle("/endpoints/snapshot",
		bouncer.AdminAccess(httperror.LoggerHandler(h.endpointSnapshots))).Methods(http.MethodPost)
	h.Handle("/endpoints",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.endpointList))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.endpointInspect))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointUpdate))).Methods(http.MethodPut)
	h.Handle("/endpoints/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.endpointDelete))).Methods(http.MethodDelete)
	h.Handle("/endpoints/{id}/dockerhub/{registryId}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.endpointDockerhubStatus))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}/extensions",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.endpointExtensionAdd))).Methods(http.MethodPost)
	h.Handle("/endpoints/{id}/extensions/{extensionType}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.endpointExtensionRemove))).Methods(http.MethodDelete)
	h.Handle("/endpoints/{id}/snapshot",
		bouncer.AdminAccess(httperror.LoggerHandler(h.endpointSnapshot))).Methods(http.MethodPost)
	h.Handle("/endpoints/{id}/status",
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointStatusInspect))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}/pools/{rpn}/access",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointPoolsAccessUpdate))).Methods(http.MethodPut)
	h.Handle("/endpoints/{id}/forceupdateservice",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointForceUpdateService))).Methods(http.MethodPut)
	h.Handle("/endpoints/{id}/registries",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointRegistriesList))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}/registries/{registryId}",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointRegistryInspect))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}/registries/{registryId}",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.endpointRegistryAccess))).Methods(http.MethodPut)

	return h
}

package endpointproxy

import (
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/security"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to proxy requests to external APIs.
type Handler struct {
	*mux.Router
	DataStore            dataservices.DataStore
	requestBouncer       security.BouncerService
	ProxyManager         *proxy.Manager
	ReverseTunnelService portaineree.ReverseTunnelService
}

// NewHandler creates a handler to proxy requests to external APIs.
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}

	h.Use(bouncer.AuthenticatedAccess)

	h.PathPrefix("/{id}/azure").Handler(httperror.LoggerHandler(h.proxyRequestsToAzureAPI))
	h.PathPrefix("/{id}/docker").Handler(httperror.LoggerHandler(h.proxyRequestsToDockerAPI))
	h.PathPrefix("/{id}/kubernetes").Handler(httperror.LoggerHandler(h.proxyRequestsToKubernetesAPI))
	h.PathPrefix("/{id}/agent/docker").Handler(httperror.LoggerHandler(h.proxyRequestsToDockerAPI))
	h.PathPrefix("/{id}/agent/kubernetes").Handler(httperror.LoggerHandler(h.proxyRequestsToKubernetesAPI))
	return h
}

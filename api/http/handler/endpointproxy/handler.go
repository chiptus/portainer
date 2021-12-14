package endpointproxy

import (
	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/security"
)

// Handler is the HTTP handler used to proxy requests to external APIs.
type Handler struct {
	*mux.Router
	DataStore            portainer.DataStore
	requestBouncer       *security.RequestBouncer
	ProxyManager         *proxy.Manager
	ReverseTunnelService portainer.ReverseTunnelService
}

// NewHandler creates a handler to proxy requests to external APIs.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
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
	h.PathPrefix("/{id}/storidge").Handler(httperror.LoggerHandler(h.proxyRequestsToStoridgeAPI))
	return h
}

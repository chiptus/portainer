package endpointedge

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"

	"github.com/gorilla/mux"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
)

// Handler is the HTTP handler used to handle edge environment(endpoint) operations.
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	DataStore            dataservices.DataStore
	FileService          portainer.FileService
	ReverseTunnelService portaineree.ReverseTunnelService
}

// NewHandler creates a handler to manage environment(endpoint) operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}

	h.Handle("/{id}/edge/status",
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStatusInspect))).Methods(http.MethodGet)
	h.Handle("/{id}/edge/stacks/{stackId}",
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStackInspect))).Methods(http.MethodGet)
	h.Handle("/{id}/edge/jobs/{jobID}/logs",
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeJobsLogs))).Methods(http.MethodPost)
	return h
}

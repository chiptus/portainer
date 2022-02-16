package status

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
)

// Handler is the HTTP handler used to handle status operations.
type Handler struct {
	*mux.Router
	Status    *portaineree.Status
	DataStore dataservices.DataStore
}

// NewHandler creates a handler to manage status operations.
func NewHandler(bouncer *security.RequestBouncer, status *portaineree.Status) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		Status: status,
	}
	h.Handle("/status",
		bouncer.PublicAccess(httperror.LoggerHandler(h.statusInspect))).Methods(http.MethodGet)
	h.Handle("/status/version",
		bouncer.AuthenticatedAccess(http.HandlerFunc(h.statusInspectVersion))).Methods(http.MethodGet)
	h.Handle("/status/nodes",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.statusNodesCount))).Methods(http.MethodGet)

	return h
}

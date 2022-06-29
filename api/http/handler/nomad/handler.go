package nomad

import (
	"github.com/portainer/portainer-ee/api/dataservices"
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/nomad/clientFactory"
)

// Handler - Nomad handler
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	nomadClientFactory   *clientFactory.ClientFactory
	AuthorizationService *authorization.Service
	JwtService           portaineree.JWTService
	userActivityService  portaineree.UserActivityService
	DataStore            dataservices.DataStore
}

// NewHandler creates a handler to manage Nomad operations.
func NewHandler(bouncer *security.RequestBouncer, nomadClientFactory *clientFactory.ClientFactory) *Handler {
	h := &Handler{
		Router:             mux.NewRouter(),
		nomadClientFactory: nomadClientFactory,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess)

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)

	authenticatedRouter.Handle("/nomad/allocation/{id}/events", httperror.LoggerHandler(h.getTaskEvents)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/nomad/allocation/{id}/logs", httperror.LoggerHandler(h.getTaskLogs)).Methods(http.MethodGet)

	authenticatedRouter.Handle("/nomad/leader", httperror.LoggerHandler(h.getLeader)).Methods(http.MethodGet)

	authenticatedRouter.Handle("/nomad/jobs", httperror.LoggerHandler(h.listJobs)).Methods(http.MethodGet)
	adminRouter.Handle("/nomad/jobs/{id}", httperror.LoggerHandler(h.deleteJob)).Methods(http.MethodDelete)

	authenticatedRouter.Handle("/nomad/dashboard", httperror.LoggerHandler(h.getDashboard)).Methods(http.MethodGet)

	return h
}

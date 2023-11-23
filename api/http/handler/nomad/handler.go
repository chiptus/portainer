package nomad

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/nomad/clientFactory"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler - Nomad handler
type Handler struct {
	*mux.Router
	nomadClientFactory   *clientFactory.ClientFactory
	authorizationService *authorization.Service
	dataStore            dataservices.DataStore
}

// NewHandler creates a handler to manage Nomad operations.
func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore, nomadClientFactory *clientFactory.ClientFactory, authorizationService *authorization.Service) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		dataStore:            dataStore,
		nomadClientFactory:   nomadClientFactory,
		authorizationService: authorizationService,
	}

	subrouter := h.PathPrefix("/nomad/endpoints/{endpointId}").Subrouter()
	subrouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "endpointId"))

	adminRouter := subrouter.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess)
	adminRouter.Handle("/jobs/{id}", httperror.LoggerHandler(h.deleteJob)).Methods(http.MethodDelete)

	authenticatedRouter := subrouter.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)
	authenticatedRouter.Handle("/allocation/{id}/events", httperror.LoggerHandler(h.getTaskEvents)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/allocation/{id}/logs", httperror.LoggerHandler(h.getTaskLogs)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/leader", httperror.LoggerHandler(h.getLeader)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/jobs", httperror.LoggerHandler(h.listJobs)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/dashboard", httperror.LoggerHandler(h.getDashboard)).Methods(http.MethodGet)

	return h
}

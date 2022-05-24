package endpointedge

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	portainer "github.com/portainer/portainer/api"
)

// Handler is the HTTP handler used to handle edge environment(endpoint) operations.
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	DataStore            dataservices.DataStore
	FileService          portainer.FileService
	ReverseTunnelService portaineree.ReverseTunnelService
	EdgeService          edge.Service
}

// NewHandler creates a handler to manage environment(endpoint) operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore, fileService portainer.FileService, reverseTunnelService portaineree.ReverseTunnelService, edgeService edge.Service) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		DataStore:            dataStore,
		FileService:          fileService,
		ReverseTunnelService: reverseTunnelService,
		EdgeService:          edgeService,
	}

	endpointRouter := h.PathPrefix("/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))

	endpointRouter.PathPrefix("/edge/status").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStatusInspect))).Methods(http.MethodGet)

	endpointRouter.PathPrefix("/edge/stacks/{stackId}").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStackInspect))).Methods(http.MethodGet)

	endpointRouter.PathPrefix("/edge/jobs/{jobID}/logs").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeJobsLogs))).Methods(http.MethodPost)

	endpointRouter.PathPrefix("/edge/trust").Handler(bouncer.AdminAccess(httperror.LoggerHandler(h.endpointTrust))).Methods(http.MethodPost)

	h.Handle("/edge/async", handlers.CompressHandler(bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeAsync)))).Methods(http.MethodPost)
	h.Handle("/edge/generate-key", bouncer.AdminAccess(httperror.LoggerHandler(h.endpointEdgeGenerateKey))).Methods(http.MethodPost)

	return h
}

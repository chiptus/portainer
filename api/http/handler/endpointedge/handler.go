package endpointedge

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/http/middlewares"

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
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore, fileService portainer.FileService, reverseTunnelService portaineree.ReverseTunnelService) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		DataStore:            dataStore,
		FileService:          fileService,
		ReverseTunnelService: reverseTunnelService,
	}

	endpointRouter := h.PathPrefix("/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))

	endpointRouter.PathPrefix("/edge/status").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStatusInspect))).Methods(http.MethodGet)

	endpointRouter.PathPrefix("/edge/stacks/{stackId}").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStackInspect))).Methods(http.MethodGet)

	endpointRouter.PathPrefix("/edge/jobs/{jobID}/logs").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeJobsLogs))).Methods(http.MethodPost)

	endpointRouter.PathPrefix("/edge/commands").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeAsyncCommands))).Methods(http.MethodGet)

	endpointRouter.PathPrefix("/edge/trust").Handler(bouncer.AdminAccess(httperror.LoggerHandler(h.endpointTrust))).Methods(http.MethodPost)

	h.Handle("/edge/generate-key", bouncer.AdminAccess(httperror.LoggerHandler(h.endpointEdgeGenerateKey))).Methods(http.MethodPost)
	return h
}

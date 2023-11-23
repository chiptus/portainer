package endpointedge

import (
	"net/http"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/edge/staggers"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	"github.com/portainer/portainer-ee/api/license"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

// Handler is the HTTP handler used to handle edge environment(endpoint) operations.
type Handler struct {
	*mux.Router
	requestBouncer       security.BouncerService
	DataStore            dataservices.DataStore
	FileService          portainer.FileService
	ReverseTunnelService portaineree.ReverseTunnelService
	EdgeService          *edgeasync.Service
	edgeUpdateService    updateschedules.EdgeUpdateService
	edgeStackCache       *cache.Cache
	staggerService       staggers.StaggerService
}

// NewHandler creates a handler to manage environment(endpoint) operations.
func NewHandler(bouncer security.BouncerService,
	dataStore dataservices.DataStore,
	fileService portainer.FileService,
	reverseTunnelService portaineree.ReverseTunnelService,
	edgeService *edgeasync.Service,
	licenseService portaineree.LicenseService,
	edgeUpdateService updateschedules.EdgeUpdateService,
	staggerService staggers.StaggerService) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		DataStore:            dataStore,
		FileService:          fileService,
		ReverseTunnelService: reverseTunnelService,
		EdgeService:          edgeService,
		edgeUpdateService:    edgeUpdateService,
		edgeStackCache: cache.New(
			portaineree.DefaultEdgeAgentCheckinIntervalInSeconds*time.Second,
			time.Minute,
		),
		staggerService: staggerService,
	}

	h.Handle("/api/endpoints/{id}/edge/status", bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStatusInspect))).Methods(http.MethodGet)

	h.Handle("/api/endpoints/edge/async", handlers.CompressHandler(bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeAsync)))).Methods(http.MethodPost)
	h.Handle("/api/endpoints/edge/generate-key", bouncer.PureAdminAccess(httperror.LoggerHandler(h.endpointEdgeGenerateKey))).Methods(http.MethodPost)

	endpointRouter := h.PathPrefix("/api/endpoints/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))

	endpointRouter.PathPrefix("/edge/stacks/{stackId}").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeStackInspect))).Methods(http.MethodGet)

	endpointRouter.PathPrefix("/edge/jobs/{jobID}/logs").Handler(
		bouncer.PublicAccess(httperror.LoggerHandler(h.endpointEdgeJobsLogs))).Methods(http.MethodPost)

	// used in waiting room
	endpointRouter.PathPrefix("/edge/trust").Handler(bouncer.AdminAccess(license.RecalculateLicenseUsage(licenseService, httperror.LoggerHandler(h.endpointTrust)))).Methods(http.MethodPost)

	endpointRouter.Handle("/edge/async/commands/stack/{stackId}", bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.createNormalStackCommand))).Methods(http.MethodPost)
	endpointRouter.Handle("/edge/async/commands/container", bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.createContainerCommand))).Methods(http.MethodPost)
	endpointRouter.Handle("/edge/async/commands/image", bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.createImageCommand))).Methods(http.MethodPost)
	endpointRouter.Handle("/edge/async/commands/volume", bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.createVolumeCommand))).Methods(http.MethodPost)

	return h
}

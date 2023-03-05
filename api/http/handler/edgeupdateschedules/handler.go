package edgeupdateschedules

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"

	"github.com/gorilla/mux"
	"github.com/portainer/portainer-ee/api/dataservices"

	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
)

const contextKey = "edgeUpdateSchedule_item"

// Handler is the HTTP handler used to handle edge environment update operations.
type Handler struct {
	*mux.Router
	requestBouncer    *security.RequestBouncer
	dataStore         dataservices.DataStore
	fileService       portainer.FileService
	updateService     *updateschedules.Service
	assetsPath        string
	edgeStacksService *edgestacks.Service
}

// NewHandler creates a handler to manage environment update operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore, fileService portainer.FileService, assetsPath string, edgeStacksService *edgestacks.Service, updateService *updateschedules.Service) *Handler {
	h := &Handler{
		Router:            mux.NewRouter(),
		requestBouncer:    bouncer,
		dataStore:         dataStore,
		fileService:       fileService,
		assetsPath:        assetsPath,
		edgeStacksService: edgeStacksService,
		updateService:     updateService,
	}

	router := h.PathPrefix("/edge_update_schedules").Subrouter()

	authenticatedRouter := router.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)

	authenticatedRouter.Handle("/agent_versions",
		httperror.LoggerHandler(h.agentVersions)).Methods(http.MethodGet)

	adminRouter := router.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess)

	adminRouter.Handle("",
		httperror.LoggerHandler(h.list)).Methods(http.MethodGet)

	adminRouter.Handle("",
		httperror.LoggerHandler(h.create)).Methods(http.MethodPost)

	adminRouter.Handle("/active",
		httperror.LoggerHandler(h.activeSchedules)).Methods(http.MethodPost)

	adminRouter.Handle("/previous_versions",
		httperror.LoggerHandler(h.previousVersions)).Methods(http.MethodGet)

	itemRouter := adminRouter.PathPrefix("/{id}").Subrouter()
	itemRouter.Use(middlewares.WithItem(updateService.Schedule, "id", contextKey))

	itemRouter.Handle("",
		httperror.LoggerHandler(h.inspect)).Methods(http.MethodGet)

	itemRouter.Handle("",
		httperror.LoggerHandler(h.update)).Methods(http.MethodPut)

	itemRouter.Handle("",
		httperror.LoggerHandler(h.delete)).Methods(http.MethodDelete)

	return h
}

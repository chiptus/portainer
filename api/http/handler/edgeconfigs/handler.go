package edgeconfigs

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
	dataStore           dataservices.DataStore
	bouncer             security.BouncerService
	userActivityService portaineree.UserActivityService
	edgeAsyncService    *edgeasync.Service
	fileService         portaineree.FileService
}

func NewHandler(dataStore dataservices.DataStore, bouncer security.BouncerService, userActivityService portaineree.UserActivityService, edgeAsyncService *edgeasync.Service, fileService portaineree.FileService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		dataStore:           dataStore,
		bouncer:             bouncer,
		edgeAsyncService:    edgeAsyncService,
		fileService:         fileService,
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))
	adminRouter.Handle("/edge_configurations", httperror.LoggerHandler(h.edgeConfigList)).Methods(http.MethodGet)
	adminRouter.Handle("/edge_configurations", httperror.LoggerHandler(h.edgeConfigCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/edge_configurations/{id}", httperror.LoggerHandler(h.edgeConfigInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/edge_configurations/{id}", httperror.LoggerHandler(h.edgeConfigUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/edge_configurations/{id}", httperror.LoggerHandler(h.edgeConfigDelete)).Methods(http.MethodDelete)

	// EE-6176 doc: restricted in handler to requests coming from remote edge agents
	h.Handle("/edge_configurations/{id}/files", httperror.LoggerHandler(h.edgeConfigFiles)).Methods(http.MethodGet)
	h.Handle("/edge_configurations/{id}/{state}", httperror.LoggerHandler(h.edgeConfigState)).Methods(http.MethodPut)

	return h
}

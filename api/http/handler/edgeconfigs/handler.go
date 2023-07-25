package edgeconfigs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
	dataStore           dataservices.DataStore
	userActivityService portaineree.UserActivityService
	edgeAsyncService    *edgeasync.Service
	fileService         portaineree.FileService
}

func NewHandler(dataStore dataservices.DataStore, bouncer security.BouncerService, userActivityService portaineree.UserActivityService, edgeAsyncService *edgeasync.Service, fileService portaineree.FileService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		dataStore:           dataStore,
		edgeAsyncService:    edgeAsyncService,
		fileService:         fileService,
		userActivityService: userActivityService,
	}

	authRoutes := h.NewRoute().Subrouter()
	authRoutes.Use(bouncer.AuthenticatedAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))

	authRoutes.Handle("/edge_configurations", httperror.LoggerHandler(h.edgeConfigList)).Methods(http.MethodGet)
	authRoutes.Handle("/edge_configurations", httperror.LoggerHandler(h.edgeConfigCreate)).Methods(http.MethodPost)
	authRoutes.Handle("/edge_configurations/{id}", httperror.LoggerHandler(h.edgeConfigInspect)).Methods(http.MethodGet)
	authRoutes.Handle("/edge_configurations/{id}", httperror.LoggerHandler(h.edgeConfigUpdate)).Methods(http.MethodPut)
	authRoutes.Handle("/edge_configurations/{id}", httperror.LoggerHandler(h.edgeConfigDelete)).Methods(http.MethodDelete)

	h.Handle("/edge_configurations/{id}/files", httperror.LoggerHandler(h.edgeConfigFiles)).Methods(http.MethodGet)
	h.Handle("/edge_configurations/{id}/{state}", httperror.LoggerHandler(h.edgeConfigState)).Methods(http.MethodPut)

	return h
}

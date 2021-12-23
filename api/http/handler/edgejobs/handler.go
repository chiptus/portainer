package edgejobs

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	portainer "github.com/portainer/portainer/api"
)

// Handler is the HTTP handler used to handle Edge job operations.
type Handler struct {
	*mux.Router
	DataStore            portaineree.DataStore
	FileService          portainer.FileService
	ReverseTunnelService portaineree.ReverseTunnelService
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage Edge job operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AdminAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/edge_jobs", httperror.LoggerHandler(h.edgeJobList)).Methods(http.MethodGet)
	h.Handle("/edge_jobs", httperror.LoggerHandler(h.edgeJobCreate)).Methods(http.MethodPost)
	h.Handle("/edge_jobs/{id}", httperror.LoggerHandler(h.edgeJobInspect)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}", httperror.LoggerHandler(h.edgeJobUpdate)).Methods(http.MethodPut)
	h.Handle("/edge_jobs/{id}", httperror.LoggerHandler(h.edgeJobDelete)).Methods(http.MethodDelete)
	h.Handle("/edge_jobs/{id}/file", httperror.LoggerHandler(h.edgeJobFile)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}/tasks", httperror.LoggerHandler(h.edgeJobTasksList)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}/tasks/{taskID}/logs", httperror.LoggerHandler(h.edgeJobTaskLogsInspect)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}/tasks/{taskID}/logs", httperror.LoggerHandler(h.edgeJobTasksCollect)).Methods(http.MethodPost)
	h.Handle("/edge_jobs/{id}/tasks/{taskID}/logs", httperror.LoggerHandler(h.edgeJobTasksClear)).Methods(http.MethodDelete)
	return h
}

package edgegroups

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	DataStore           portainer.DataStore
	userActivityService portainer.UserActivityService
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portainer.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AdminAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/edge_groups", httperror.LoggerHandler(h.edgeGroupCreate)).Methods(http.MethodPost)
	h.Handle("/edge_groups", httperror.LoggerHandler(h.edgeGroupList)).Methods(http.MethodGet)
	h.Handle("/edge_groups/{id}", httperror.LoggerHandler(h.edgeGroupInspect)).Methods(http.MethodGet)
	h.Handle("/edge_groups/{id}", httperror.LoggerHandler(h.edgeGroupUpdate)).Methods(http.MethodPut)
	h.Handle("/edge_groups/{id}", httperror.LoggerHandler(h.edgeGroupDelete)).Methods(http.MethodDelete)
	return h
}

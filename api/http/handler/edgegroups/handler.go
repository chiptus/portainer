package edgegroups

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	DataStore           portaineree.DataStore
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService) *Handler {
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

package endpointgroups

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            dataservices.DataStore
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/endpoint_groups", httperror.LoggerHandler(h.endpointGroupCreate)).Methods(http.MethodPost)
	h.Handle("/endpoint_groups", httperror.LoggerHandler(h.endpointGroupList)).Methods(http.MethodGet)
	h.Handle("/endpoint_groups/{id}", httperror.LoggerHandler(h.endpointGroupInspect)).Methods(http.MethodGet)
	h.Handle("/endpoint_groups/{id}", httperror.LoggerHandler(h.endpointGroupUpdate)).Methods(http.MethodPut)
	h.Handle("/endpoint_groups/{id}", httperror.LoggerHandler(h.endpointGroupDelete)).Methods(http.MethodDelete)
	h.Handle("/endpoint_groups/{id}/endpoints/{endpointId}", httperror.LoggerHandler(h.endpointGroupAddEndpoint)).Methods(http.MethodPut)
	h.Handle("/endpoint_groups/{id}/endpoints/{endpointId}", httperror.LoggerHandler(h.endpointGroupDeleteEndpoint)).Methods(http.MethodDelete)
	return h
}

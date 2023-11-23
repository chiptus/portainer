package endpointgroups

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/pendingactions"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	AuthorizationService  *authorization.Service
	DataStore             dataservices.DataStore
	userActivityService   portaineree.UserActivityService
	edgeAsyncService      *edgeasync.Service
	PendingActionsService *pendingactions.PendingActionsService
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService, edgeAsyncService *edgeasync.Service) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
		edgeAsyncService:    edgeAsyncService,
	}

	// admins + edge admins + roles that have the authorization are able to list endpoint groups
	restrictedRouter := h.NewRoute().Subrouter()
	restrictedRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))
	restrictedRouter.Handle("/endpoint_groups", httperror.LoggerHandler(h.endpointGroupList)).Methods(http.MethodGet)
	restrictedRouter.Handle("/endpoint_groups/{id}", httperror.LoggerHandler(h.endpointGroupInspect)).Methods(http.MethodGet)

	// only admins are able to manage endpoint groups
	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess, useractivity.LogUserActivity(h.userActivityService))
	adminRouter.Handle("/endpoint_groups", httperror.LoggerHandler(h.endpointGroupCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/endpoint_groups/{id}", httperror.LoggerHandler(h.endpointGroupUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/endpoint_groups/{id}", httperror.LoggerHandler(h.endpointGroupDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/endpoint_groups/{id}/endpoints/{endpointId}", httperror.LoggerHandler(h.endpointGroupAddEndpoint)).Methods(http.MethodPut)
	adminRouter.Handle("/endpoint_groups/{id}/endpoints/{endpointId}", httperror.LoggerHandler(h.endpointGroupDeleteEndpoint)).Methods(http.MethodDelete)
	return h
}

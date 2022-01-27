package webhooks

import (
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
)

// Handler is the HTTP handler used to handle webhook operations.
type Handler struct {
	*mux.Router
	requestBouncer      *security.RequestBouncer
	dataStore           portaineree.DataStore
	DockerClientFactory *docker.ClientFactory
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage webhooks operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore portaineree.DataStore, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
		dataStore:           dataStore,
		requestBouncer:      bouncer,
	}

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess, useractivity.LogUserActivity(h.userActivityService))

	authenticatedRouter.Handle("/webhooks", httperror.LoggerHandler(h.webhookCreate)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/webhooks", httperror.LoggerHandler(h.webhookList)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/webhooks/{id}", httperror.LoggerHandler(h.webhookUpdate)).Methods(http.MethodPut)
	authenticatedRouter.Handle("/webhooks/{id}", httperror.LoggerHandler(h.webhookDelete)).Methods(http.MethodDelete)
	publicRouter.Handle("/webhooks/{token}", httperror.LoggerHandler(h.webhookExecute)).Methods(http.MethodPost)
	return h
}

func (handler *Handler) checkResourceAccess(r *http.Request, resourceID string, resourceControlType portaineree.ResourceControlType) *httperror.HandlerError {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve user info from request context", Err: err}
	}
	// non-admins
	rc, err := handler.dataStore.ResourceControl().ResourceControlByResourceIDAndType(resourceID, resourceControlType)
	if rc == nil || err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve a resource control associated to the resource", Err: err}
	}
	userTeamIDs := make([]portaineree.TeamID, 0)
	for _, membership := range securityContext.UserMemberships {
		userTeamIDs = append(userTeamIDs, membership.TeamID)
	}
	canAccess := authorization.UserCanAccessResource(securityContext.UserID, userTeamIDs, rc)
	if !canAccess {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "This operation is disabled for non-admin users and unassigned access users"}
	}
	return nil
}

func (handler *Handler) checkAuthorization(r *http.Request, endpoint *portaineree.Endpoint, authorizations []portaineree.Authorization) (bool, *httperror.HandlerError) {
	err := handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return false, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Permission denied to access environment", Err: err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return false, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve user info from request context", Err: err}
	}

	authService := authorization.NewService(handler.dataStore)
	isAdminOrAuthorized, err := authService.UserIsAdminOrAuthorized(securityContext.UserID, endpoint.ID, authorizations)
	if err != nil {
		return false, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to get user authorizations", Err: err}
	}
	return isAdminOrAuthorized, nil
}

package resourcecontrols

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
)

// Handler is the HTTP handler used to handle resource control operations.
type Handler struct {
	*mux.Router
	dataStore           portaineree.DataStore
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage resource control operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore portaineree.DataStore, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		dataStore:           dataStore,
		userActivityService: userActivityService,
	}

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/resource_controls", httperror.LoggerHandler(h.resourceControlCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/resource_controls/{id}", httperror.LoggerHandler(h.resourceControlDelete)).Methods(http.MethodDelete)

	authenticatedRouter.Handle("/resource_controls/{id}", httperror.LoggerHandler(h.resourceControlUpdate)).Methods(http.MethodPut)

	return h
}

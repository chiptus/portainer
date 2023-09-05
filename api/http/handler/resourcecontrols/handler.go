package resourcecontrols

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle resource control operations.
type Handler struct {
	*mux.Router
	DataStore           dataservices.DataStore
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage resource control operations.
func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		DataStore:           dataStore,
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

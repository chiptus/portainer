package webhooks

import (
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

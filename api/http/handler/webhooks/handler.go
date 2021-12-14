package webhooks

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/docker"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
)

// Handler is the HTTP handler used to handle webhook operations.
type Handler struct {
	*mux.Router
	dataStore           portainer.DataStore
	DockerClientFactory *docker.ClientFactory
	userActivityService portainer.UserActivityService
}

// NewHandler creates a handler to manage webhooks operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore portainer.DataStore, userActivityService portainer.UserActivityService) *Handler {
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

package services

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/http/security"
)

type Handler struct {
	*mux.Router
	dockerClientFactory *client.ClientFactory
	dataStore           dataservices.DataStore
}

// NewHandler creates a handler to process non-proxied requests to docker APIs directly.
func NewHandler(routePrefix string, bouncer security.BouncerService, dataStore dataservices.DataStore, dockerClientFactory *client.ClientFactory) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		dataStore:           dataStore,
		dockerClientFactory: dockerClientFactory,
	}

	router := h.PathPrefix(routePrefix).Subrouter()
	router.Use(bouncer.AuthenticatedAccess)

	router.Handle("/{serviceID}/image_status", httperror.LoggerHandler(h.ServiceImageStatus)).Methods(http.MethodGet)
	return h
}

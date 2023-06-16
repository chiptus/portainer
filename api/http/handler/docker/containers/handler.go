package containers

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/http/security"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
	dockerClientFactory *client.ClientFactory
	dataStore           dataservices.DataStore
	containerService    *docker.ContainerService
	bouncer             security.BouncerService
}

// NewHandler creates a handler to process non-proxied requests to docker APIs directly.
func NewHandler(routePrefix string, bouncer security.BouncerService, dataStore dataservices.DataStore, dockerClientFactory *client.ClientFactory, containerService *docker.ContainerService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		dataStore:           dataStore,
		dockerClientFactory: dockerClientFactory,
		containerService:    containerService,
		bouncer:             bouncer,
	}

	router := h.PathPrefix(routePrefix).Subrouter()
	router.Use(bouncer.AuthenticatedAccess)

	router.Handle("/{containerId}/gpus", httperror.LoggerHandler(h.containerGpusInspect)).Methods(http.MethodGet)
	router.Handle("/{containerId}/image_status", httperror.LoggerHandler(h.containerImageStatus)).Methods(http.MethodGet)
	router.Handle("/{containerId}/recreate", httperror.LoggerHandler(h.recreate)).Methods(http.MethodPost)
	return h
}

package docker

import (
	"errors"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/http/handler/docker/services"
	"net/http"

	"github.com/portainer/portainer-ee/api/internal/endpointutils"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/handler/docker/containers"
	"github.com/portainer/portainer-ee/api/http/handler/dockersnapshot"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

// Handler is the HTTP handler which will natively deal with to external environments(endpoints).
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	dataStore            dataservices.DataStore
	dockerClientFactory  *client.ClientFactory
	authorizationService *authorization.Service
	containerService     *docker.ContainerService
}

// NewHandler creates a handler to process non-proxied requests to docker APIs directly.
func NewHandler(bouncer *security.RequestBouncer, authorizationService *authorization.Service, dataStore dataservices.DataStore, dockerClientFactory *client.ClientFactory, containerService *docker.ContainerService) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		authorizationService: authorizationService,
		dataStore:            dataStore,
		dockerClientFactory:  dockerClientFactory,
		containerService:     containerService,
	}

	// endpoints
	endpointRouter := h.PathPrefix("/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))
	endpointRouter.Use()
	endpointRouter.Use(dockerOnlyMiddleware)

	dockerSnapshotHandler := dockersnapshot.NewHandler("/{id}/snapshot", bouncer, dataStore)
	endpointRouter.PathPrefix("/snapshot").Handler(dockerSnapshotHandler)

	containersHandler := containers.NewHandler("/{id}/containers", bouncer, dataStore, dockerClientFactory, containerService)
	endpointRouter.PathPrefix("/containers").Handler(containersHandler)

	servicesHandler := services.NewHandler("/{id}/services", bouncer, dataStore, dockerClientFactory)
	endpointRouter.PathPrefix("/services").Handler(servicesHandler)
	return h
}

func dockerOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		endpoint, err := middlewares.FetchEndpoint(request)
		if err != nil {
			httperror.WriteError(rw, http.StatusInternalServerError, "Unable to find an environment on request context", err)
			return
		}

		if !endpointutils.IsDockerEndpoint(endpoint) {
			errMessage := "environment is not a docker environment"
			httperror.WriteError(rw, http.StatusBadRequest, errMessage, errors.New(errMessage))
			return
		}
		next.ServeHTTP(rw, request)
	})
}

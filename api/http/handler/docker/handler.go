package docker

import (
	"errors"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

// Handler is the HTTP handler which will natively deal with to external environments(endpoints).
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	DataStore            dataservices.DataStore
	DockerClientFactory  *docker.ClientFactory
	AuthorizationService *authorization.Service
}

// NewHandler creates a handler to process non-proxied requests to docker APIs directly.
func NewHandler(bouncer *security.RequestBouncer, authorizationService *authorization.Service, dataStore dataservices.DataStore, dockerClientFactory *docker.ClientFactory) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		AuthorizationService: authorizationService,
		DataStore:            dataStore,
		DockerClientFactory:  dockerClientFactory,
	}

	dockerRouter := h.PathPrefix("/docker").Subrouter()
	dockerRouter.Use(bouncer.AuthenticatedAccess)

	// endpoints
	endpointRouter := dockerRouter.PathPrefix("/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))
	endpointRouter.Use(dockerOnlyMiddleware)

	endpointRouter.PathPrefix("/images/status").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.imageStatus))).Methods(http.MethodPost)

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

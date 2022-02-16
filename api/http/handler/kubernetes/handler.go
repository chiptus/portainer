package kubernetes

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

// Handler is the HTTP handler which will natively deal with to external environments(endpoints).
type Handler struct {
	*mux.Router
	requestBouncer          *security.RequestBouncer
	DataStore               dataservices.DataStore
	KubernetesClientFactory *cli.ClientFactory
	AuthorizationService    *authorization.Service
	userActivityService     portaineree.UserActivityService
	JwtService              portaineree.JWTService
	BaseURL                 string
}

// NewHandler creates a handler to process pre-proxied requests to external APIs.
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore, baseURL string, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		requestBouncer:      bouncer,
		DataStore:           dataStore,
		BaseURL:             baseURL,
		userActivityService: userActivityService,
	}

	kubeRouter := h.PathPrefix("/kubernetes").Subrouter()
	kubeRouter.Use(bouncer.AuthenticatedAccess)
	kubeRouter.PathPrefix("/config").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.getKubernetesConfig))).Methods(http.MethodGet)

	// endpoints
	endpointRouter := kubeRouter.PathPrefix("/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))
	endpointRouter.Use(kubeOnlyMiddleware)

	endpointRouter.PathPrefix("/nodes_limits").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.getKubernetesNodesLimits))).Methods(http.MethodGet)

	// namespaces
	// in the future this piece of code might be in another package (or a few different packages - namespaces/namespace?)
	// to keep it simple, we've decided to leave it like this.
	namespaceRouter := endpointRouter.PathPrefix("/namespaces/{namespace}").Subrouter()
	namespaceRouter.Use(useractivity.LogUserActivity(h.userActivityService))
	namespaceRouter.Handle("/system", httperror.LoggerHandler(h.namespacesToggleSystem)).Methods(http.MethodPut)

	return h
}

func kubeOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		endpoint, err := middlewares.FetchEndpoint(request)
		if err != nil {
			httperror.WriteError(rw, http.StatusInternalServerError, "Unable to find an environment on request context", err)
			return
		}

		if !endpointutils.IsKubernetesEndpoint(endpoint) {
			errMessage := "environment is not a Kubernetes environment"
			httperror.WriteError(rw, http.StatusBadRequest, errMessage, errors.New(errMessage))
			return
		}

		next.ServeHTTP(rw, request)
	})
}

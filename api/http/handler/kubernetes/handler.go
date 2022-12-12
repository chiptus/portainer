package kubernetes

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

// Handler is the HTTP handler which will natively deal with to external environments(endpoints).
type Handler struct {
	*mux.Router
	opaOperationMutex        *sync.Mutex
	requestBouncer           *security.RequestBouncer
	DataStore                dataservices.DataStore
	KubernetesClientFactory  *cli.ClientFactory
	kubeClusterAccessService kubernetes.KubeClusterAccessService
	AuthorizationService     *authorization.Service
	userActivityService      portaineree.UserActivityService
	KubernetesDeployer       portaineree.KubernetesDeployer
	JwtService               portaineree.JWTService
	fileService              portainer.FileService
	baseFileDir              string
}

// NewHandler creates a handler to process pre-proxied requests to external APIs.
func NewHandler(bouncer *security.RequestBouncer, authorizationService *authorization.Service, dataStore dataservices.DataStore, jwtService portaineree.JWTService, kubeClusterAccessService kubernetes.KubeClusterAccessService, kubernetesClientFactory *cli.ClientFactory, kubernetesClient portaineree.KubeClient, userActivityService portaineree.UserActivityService, k8sDeployer portaineree.KubernetesDeployer, fileService portainer.FileService, assetsPath string) *Handler {

	h := &Handler{
		Router:                   mux.NewRouter(),
		opaOperationMutex:        &sync.Mutex{},
		requestBouncer:           bouncer,
		AuthorizationService:     authorizationService,
		DataStore:                dataStore,
		JwtService:               jwtService,
		kubeClusterAccessService: kubeClusterAccessService,
		KubernetesClientFactory:  kubernetesClientFactory,
		KubernetesDeployer:       k8sDeployer,
		userActivityService:      userActivityService,
		fileService:              fileService,
		baseFileDir:              assetsPath,
	}

	kubeRouter := h.PathPrefix("/kubernetes").Subrouter()
	kubeRouter.Use(bouncer.AuthenticatedAccess)
	kubeRouter.PathPrefix("/config").Handler(httperror.LoggerHandler(h.getKubernetesConfig)).Methods(http.MethodGet)

	// endpoints
	endpointRouter := kubeRouter.PathPrefix("/{id}").Subrouter()
	endpointRouter.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"))
	endpointRouter.Use(kubeOnlyMiddleware)
	endpointRouter.Use(h.kubeClient)

	endpointRouter.PathPrefix("/nodes_limits").Handler(httperror.LoggerHandler(h.getKubernetesNodesLimits)).Methods(http.MethodGet)
	endpointRouter.PathPrefix("/opa").Handler(httperror.LoggerHandler(h.getK8sPodSecurityRule)).Methods(http.MethodGet)
	endpointRouter.PathPrefix("/opa").Handler(httperror.LoggerHandler(h.updateK8sPodSecurityRule)).Methods(http.MethodPut)
	endpointRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.getKubernetesIngressControllers)).Methods(http.MethodGet)
	endpointRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.updateKubernetesIngressControllers)).Methods(http.MethodPut)
	endpointRouter.Handle("/ingresses/delete", httperror.LoggerHandler(h.deleteKubernetesIngresses)).Methods(http.MethodPost)
	endpointRouter.Handle("/services/delete", httperror.LoggerHandler(h.deleteKubernetesServices)).Methods(http.MethodPost)
	endpointRouter.Path("/rbac_enabled").Handler(httperror.LoggerHandler(h.isRBACEnabled)).Methods(http.MethodGet)
	endpointRouter.Path("/namespaces").Handler(httperror.LoggerHandler(h.createKubernetesNamespace)).Methods(http.MethodPost)
	endpointRouter.Path("/namespaces").Handler(httperror.LoggerHandler(h.updateKubernetesNamespace)).Methods(http.MethodPut)
	endpointRouter.Path("/namespaces").Handler(httperror.LoggerHandler(h.getKubernetesNamespaces)).Methods(http.MethodGet)
	endpointRouter.Path("/namespaces/{namespace}").Handler(httperror.LoggerHandler(h.deleteKubernetesNamespaces)).Methods(http.MethodDelete)
	endpointRouter.Path("/namespaces/{namespace}").Handler(httperror.LoggerHandler(h.getKubernetesNamespace)).Methods(http.MethodGet)

	// namespaces
	namespaceRouter := endpointRouter.PathPrefix("/namespaces/{namespace}").Subrouter()
	namespaceRouter.Use(useractivity.LogUserActivity(h.userActivityService))
	namespaceRouter.Handle("/system", httperror.LoggerHandler(h.namespacesToggleSystem)).Methods(http.MethodPut)
	namespaceRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.getKubernetesIngressControllersByNamespace)).Methods(http.MethodGet)
	namespaceRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.updateKubernetesIngressControllersByNamespace)).Methods(http.MethodPut)
	namespaceRouter.Handle("/configmaps", httperror.LoggerHandler(h.getKubernetesConfigMaps)).Methods(http.MethodGet)
	namespaceRouter.Handle("/ingresses", httperror.LoggerHandler(h.createKubernetesIngress)).Methods(http.MethodPost)
	namespaceRouter.Handle("/ingresses", httperror.LoggerHandler(h.updateKubernetesIngress)).Methods(http.MethodPut)
	namespaceRouter.Handle("/ingresses", httperror.LoggerHandler(h.getKubernetesIngresses)).Methods(http.MethodGet)
	namespaceRouter.Handle("/services", httperror.LoggerHandler(h.createKubernetesService)).Methods(http.MethodPost)
	namespaceRouter.Handle("/services", httperror.LoggerHandler(h.updateKubernetesService)).Methods(http.MethodPut)
	namespaceRouter.Handle("/services", httperror.LoggerHandler(h.getKubernetesServices)).Methods(http.MethodGet)
	namespaceRouter.Handle("/applications", httperror.LoggerHandler(h.getKubernetesApplications)).Methods(http.MethodGet)
	namespaceRouter.Handle("/applications/{kind}/{name}", httperror.LoggerHandler(h.getKubernetesApplication)).Methods(http.MethodGet)

	return h
}

func kubeOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		endpoint, err := middlewares.FetchEndpoint(request)
		if err != nil {
			httperror.WriteError(
				rw,
				http.StatusInternalServerError,
				"Unable to find an environment on request context",
				err,
			)
			return
		}

		if !endpointutils.IsKubernetesEndpoint(endpoint) {
			errMessage := "environment is not a Kubernetes environment"
			httperror.WriteError(
				rw,
				http.StatusBadRequest,
				errMessage,
				errors.New(errMessage),
			)
			return
		}

		next.ServeHTTP(rw, request)
	})
}

func (handler *Handler) kubeClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusBadRequest,
				"Invalid environment identifier route variable",
				err,
			)
			return
		}

		endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
		if err == portainerDsErrors.ErrObjectNotFound {
			httperror.WriteError(
				w,
				http.StatusNotFound,
				"Unable to find an environment with the specified identifier inside the database",
				err,
			)
			return
		} else if err != nil {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Unable to find an environment with the specified identifier inside the database",
				err,
			)
			return
		}

		if handler.KubernetesClientFactory == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Generate a proxied kubeconfig, then create a kubeclient using it.
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusForbidden,
				"Permission denied to access environment",
				err,
			)
			return
		}
		bearerToken, err := handler.JwtService.GenerateTokenForKubeconfig(tokenData)
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Unable to generate JWT token",
				err,
			)
			return
		}
		singleEndpointList := []portaineree.Endpoint{
			*endpoint,
		}
		config, handlerErr := handler.buildConfig(
			r,
			tokenData,
			bearerToken,
			singleEndpointList,
		)
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Unable to build endpoint kubeconfig",
				handlerErr.Err,
			)
			return
		}

		if len(config.Clusters) == 0 {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Unable build cluster kubeconfig",
				errors.New("Unable build cluster kubeconfig"),
			)
			return
		}

		// Manually setting the localhost to route
		// the request to proxy server
		serverURL, err := url.Parse(config.Clusters[0].Cluster.Server)
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Unable parse cluster's kubeconfig server URL",
				nil,
			)
			return
		}
		serverURL.Scheme = "https"
		serverURL.Host = "localhost" + handler.KubernetesClientFactory.AddrHTTPS
		config.Clusters[0].Cluster.Server = serverURL.String()

		yaml, err := cli.GenerateYAML(config)
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Unable to generate yaml from endpoint kubeconfig",
				err,
			)
			return
		}
		kubeCli, err := handler.KubernetesClientFactory.CreateKubeClientFromKubeConfig(endpoint.Name, []byte(yaml))
		if err != nil {
			httperror.WriteError(
				w,
				http.StatusInternalServerError,
				"Failed to create client from kubeconfig",
				err,
			)
			return
		}

		handler.KubernetesClientFactory.SetProxyKubeClient(strconv.Itoa(int(endpoint.ID)), r.Header.Get("Authorization"), kubeCli)
		next.ServeHTTP(w, r)
	})
}

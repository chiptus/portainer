package kubernetes

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"

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
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler which will natively deal with to external environments(endpoints).
type Handler struct {
	*mux.Router
	opaOperationMutex        *sync.Mutex
	requestBouncer           security.BouncerService
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
func NewHandler(bouncer security.BouncerService, authorizationService *authorization.Service, dataStore dataservices.DataStore, jwtService portaineree.JWTService, kubeClusterAccessService kubernetes.KubeClusterAccessService, kubernetesClientFactory *cli.ClientFactory, kubernetesClient portaineree.KubeClient, userActivityService portaineree.UserActivityService, k8sDeployer portaineree.KubernetesDeployer, fileService portainer.FileService, assetsPath string) *Handler {
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
	endpointRouter.Use(h.kubeClientMiddleware)

	endpointRouter.Handle("/nodes_limits", httperror.LoggerHandler(h.getKubernetesNodesLimits)).Methods(http.MethodGet)
	endpointRouter.Handle("/max_resource_limits", httperror.LoggerHandler(h.getKubernetesMaxResourceLimits)).Methods(http.MethodGet)
	endpointRouter.Handle("/opa", httperror.LoggerHandler(h.getK8sPodSecurityRule)).Methods(http.MethodGet)
	endpointRouter.Handle("/opa", httperror.LoggerHandler(h.updateK8sPodSecurityRule)).Methods(http.MethodPut)
	endpointRouter.Handle("/metrics/nodes", httperror.LoggerHandler(h.getKubernetesMetricsForAllNodes)).Methods(http.MethodGet)
	endpointRouter.Handle("/metrics/nodes/{name}", httperror.LoggerHandler(h.getKubernetesMetricsForNode)).Methods(http.MethodGet)
	endpointRouter.Handle("/metrics/pods/namespace/{namespace}", httperror.LoggerHandler(h.getKubernetesMetricsForAllPods)).Methods(http.MethodGet)
	endpointRouter.Handle("/metrics/pods/namespace/{namespace}/{name}", httperror.LoggerHandler(h.getKubernetesMetricsForPod)).Methods(http.MethodGet)
	endpointRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.getKubernetesIngressControllers)).Methods(http.MethodGet)
	endpointRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.updateKubernetesIngressControllers)).Methods(http.MethodPut)
	endpointRouter.Handle("/ingresses/delete", httperror.LoggerHandler(h.deleteKubernetesIngresses)).Methods(http.MethodPost)
	endpointRouter.Handle("/services/delete", httperror.LoggerHandler(h.deleteKubernetesServices)).Methods(http.MethodPost)
	endpointRouter.Handle("/service_accounts/delete", httperror.LoggerHandler(h.deleteKubernetesServiceAccounts)).Methods(http.MethodPost)
	endpointRouter.Handle("/roles/delete", httperror.LoggerHandler(h.deleteRoles)).Methods(http.MethodPost)
	endpointRouter.Handle("/role_bindings/delete", httperror.LoggerHandler(h.deleteRoleBindings)).Methods(http.MethodPost)
	endpointRouter.Handle("/rbac_enabled", httperror.LoggerHandler(h.isRBACEnabled)).Methods(http.MethodGet)
	endpointRouter.Handle("/namespaces", httperror.LoggerHandler(h.createKubernetesNamespace)).Methods(http.MethodPost)
	endpointRouter.Handle("/namespaces", httperror.LoggerHandler(h.updateKubernetesNamespace)).Methods(http.MethodPut)
	endpointRouter.Handle("/namespaces", httperror.LoggerHandler(h.getKubernetesNamespaces)).Methods(http.MethodGet)
	endpointRouter.Handle("/namespaces/{namespace}", httperror.LoggerHandler(h.deleteKubernetesNamespace)).Methods(http.MethodDelete)
	endpointRouter.Handle("/namespaces/{namespace}", httperror.LoggerHandler(h.getKubernetesNamespace)).Methods(http.MethodGet)

	/** Cluster Roles */
	endpointRouter.Path("/cluster_roles").Handler(httperror.LoggerHandler(h.getClusterRoles)).Methods(http.MethodGet)
	endpointRouter.Path("/cluster_roles/delete").Handler(httperror.LoggerHandler(h.deleteClusterRoles)).Methods(http.MethodPost)
	endpointRouter.Path("/cluster_role_bindings").Handler(httperror.LoggerHandler(h.getClusterRoleBindings)).Methods(http.MethodGet)
	endpointRouter.Path("/cluster_role_bindings/delete").Handler(httperror.LoggerHandler(h.deleteClusterRoleBindings)).Methods(http.MethodPost)

	// namespaces
	namespaceRouter := endpointRouter.PathPrefix("/namespaces/{namespace}").Subrouter()
	namespaceRouter.Use(useractivity.LogUserActivity(h.userActivityService))
	namespaceRouter.Handle("/system", httperror.LoggerHandler(h.namespacesToggleSystem)).Methods(http.MethodPut)
	namespaceRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.getKubernetesIngressControllersByNamespace)).Methods(http.MethodGet)
	namespaceRouter.Handle("/ingresscontrollers", httperror.LoggerHandler(h.updateKubernetesIngressControllersByNamespace)).Methods(http.MethodPut)
	namespaceRouter.Handle("/configuration", httperror.LoggerHandler(h.getKubernetesConfigMapsAndSecrets)).Methods(http.MethodGet)
	namespaceRouter.Handle("/ingresses", httperror.LoggerHandler(h.createKubernetesIngress)).Methods(http.MethodPost)
	namespaceRouter.Handle("/ingresses", httperror.LoggerHandler(h.updateKubernetesIngress)).Methods(http.MethodPut)
	namespaceRouter.Handle("/ingresses", httperror.LoggerHandler(h.getKubernetesIngresses)).Methods(http.MethodGet)
	namespaceRouter.Handle("/services", httperror.LoggerHandler(h.createKubernetesService)).Methods(http.MethodPost)
	namespaceRouter.Handle("/services", httperror.LoggerHandler(h.updateKubernetesService)).Methods(http.MethodPut)
	namespaceRouter.Handle("/services", httperror.LoggerHandler(h.getKubernetesServices)).Methods(http.MethodGet)
	namespaceRouter.Handle("/applications", httperror.LoggerHandler(h.getKubernetesApplications)).Methods(http.MethodGet)
	namespaceRouter.Handle("/applications/{kind}/{name}", httperror.LoggerHandler(h.getKubernetesApplication)).Methods(http.MethodGet)
	namespaceRouter.Handle("/service_accounts", httperror.LoggerHandler(h.getKubernetesServiceAccounts)).Methods(http.MethodGet)
	namespaceRouter.Handle("/roles", httperror.LoggerHandler(h.getRoles)).Methods(http.MethodGet)
	namespaceRouter.Handle("/role_bindings", httperror.LoggerHandler(h.getRoleBindings)).Methods(http.MethodGet)

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

func (h *Handler) getProxyKubeClient(r *http.Request) (*cli.KubeClient, *httperror.HandlerError) {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return nil, httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	cli, ok := h.KubernetesClientFactory.GetProxyKubeClient(strconv.Itoa(endpointID), r.Header.Get("Authorization"))
	if !ok {
		return nil, httperror.InternalServerError("Failed to lookup KubeClient", nil)
	}

	return cli, nil
}

func (handler *Handler) kubeClientMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler.KubernetesClientFactory == nil {
			next.ServeHTTP(w, r)
			return
		}

		endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
		if err != nil {
			httperror.WriteError(w, http.StatusBadRequest, "Invalid environment identifier route variable", err)
			return
		}

		// Check if we have a kubeclient against this auth token already, otherwise generate a new one
		_, ok := handler.KubernetesClientFactory.GetProxyKubeClient(strconv.Itoa(endpointID), r.Header.Get("Authorization"))
		if ok {
			next.ServeHTTP(w, r)
			return
		}

		endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
		if err != nil {
			if handler.DataStore.IsErrObjectNotFound(err) {
				httperror.WriteError(w, http.StatusNotFound,
					"Unable to find an environment with the specified identifier inside the database", err)
				return
			}

			httperror.WriteError(w, http.StatusInternalServerError, "Error reading from the database", err)
			return
		}

		// Generate a proxied kubeconfig, then create a kubeclient using it.
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			httperror.WriteError(w, http.StatusForbidden, "Permission denied to access environment", err)
			return
		}

		bearerToken, err := handler.JwtService.GenerateTokenForKubeconfig(tokenData)
		if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable to generate JWT token", err)
			return
		}

		config := handler.buildConfig(r, tokenData, bearerToken, []portaineree.Endpoint{*endpoint}, true)
		if len(config.Clusters) == 0 {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable build cluster kubeconfig", nil)
			return
		}

		// Manually setting the localhost to route
		// the request to proxy server
		serverURL, err := url.Parse(config.Clusters[0].Cluster.Server)
		if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable parse cluster's kubeconfig server URL", nil)
			return
		}
		serverURL.Scheme = "https"
		serverURL.Host = "localhost" + handler.KubernetesClientFactory.AddrHTTPS
		config.Clusters[0].Cluster.Server = serverURL.String()

		yaml, err := cli.GenerateYAML(config)
		if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable to generate yaml from endpoint kubeconfig", err)
			return
		}
		kubeCli, err := handler.KubernetesClientFactory.CreateKubeClientFromKubeConfig(endpoint.Name, []byte(yaml))
		if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Failed to create client from kubeconfig", err)
			return
		}

		handler.KubernetesClientFactory.SetProxyKubeClient(strconv.Itoa(int(endpoint.ID)), r.Header.Get("Authorization"), kubeCli)
		next.ServeHTTP(w, r)
	})
}

package helm

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/portainer/libhelm"
	"github.com/portainer/libhelm/options"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"net/http"
	"strings"
)

type requestBouncer interface {
	AuthenticatedAccess(h http.Handler) http.Handler
}

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	requestBouncer           requestBouncer
	dataStore                dataservices.DataStore
	jwtService               portaineree.JWTService
	kubeClusterAccessService kubernetes.KubeClusterAccessService
	kubernetesDeployer       portaineree.KubernetesDeployer
	helmPackageManager       libhelm.HelmPackageManager
	userActivityService      portaineree.UserActivityService
}

// NewHandler creates a handler to manage endpoint group operations.
func NewHandler(bouncer requestBouncer, dataStore dataservices.DataStore, jwtService portaineree.JWTService, kubernetesDeployer portaineree.KubernetesDeployer, helmPackageManager libhelm.HelmPackageManager, kubeClusterAccessService kubernetes.KubeClusterAccessService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:                   mux.NewRouter(),
		requestBouncer:           bouncer,
		dataStore:                dataStore,
		jwtService:               jwtService,
		kubeClusterAccessService: kubeClusterAccessService,
		kubernetesDeployer:       kubernetesDeployer,
		helmPackageManager:       helmPackageManager,
		userActivityService:      userActivityService,
	}

	h.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"),
		bouncer.AuthenticatedAccess,
		useractivity.LogUserActivity(h.userActivityService))

	// `helm list -o json`
	h.Handle("/{id}/kubernetes/helm",
		httperror.LoggerHandler(h.helmList)).Methods(http.MethodGet)

	// `helm delete RELEASE_NAME`
	h.Handle("/{id}/kubernetes/helm/{release}",
		httperror.LoggerHandler(h.helmDelete)).Methods(http.MethodDelete)

	// `helm install [NAME] [CHART] flags`
	h.Handle("/{id}/kubernetes/helm",
		httperror.LoggerHandler(h.helmInstall)).Methods(http.MethodPost)

	h.Handle("/{id}/kubernetes/helm/repositories",
		httperror.LoggerHandler(h.userGetHelmRepos)).Methods(http.MethodGet)
	h.Handle("/{id}/kubernetes/helm/repositories",
		httperror.LoggerHandler(h.userCreateHelmRepo)).Methods(http.MethodPost)

	return h
}

// NewTemplateHandler creates a template handler to manage environment(endpoint) group operations.
func NewTemplateHandler(bouncer requestBouncer, helmPackageManager libhelm.HelmPackageManager) *Handler {
	h := &Handler{
		Router:             mux.NewRouter(),
		requestBouncer:     bouncer,
		helmPackageManager: helmPackageManager,
	}

	h.Handle("/templates/helm",
		httperror.LoggerHandler(h.helmRepoSearch)).Methods(http.MethodGet)

	// helm show [COMMAND] [CHART] [REPO] flags
	h.Handle("/templates/helm/{command:chart|values|readme}",
		httperror.LoggerHandler(h.helmShow)).Methods(http.MethodGet)

	return h
}

// getHelmClusterAccess obtains the core k8s cluster access details from request.
// The cluster access includes the cluster server url, the user's bearer token and the tls certificate.
// The cluster access is passed in as kube config CLI params to helm binary.
func (handler *Handler) getHelmClusterAccess(r *http.Request) (*options.KubernetesClusterAccess, *httperror.HandlerError) {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment on request context", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user authentication token", err}
	}

	bearerToken, err := handler.jwtService.GenerateToken(tokenData)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusUnauthorized, "Unauthorized", err}
	}

	sslSettings, err := handler.dataStore.SSLSettings().Settings()
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve settings from the database", err}
	}

	hostURL := "localhost"
	if !sslSettings.SelfSigned {
		hostURL = strings.Split(r.Host, ":")[0]
	}

	kubeConfigInternal := handler.kubeClusterAccessService.GetData(hostURL, endpoint.ID)
	return &options.KubernetesClusterAccess{
		ClusterServerURL:         kubeConfigInternal.ClusterServerURL,
		CertificateAuthorityFile: kubeConfigInternal.CertificateAuthorityFile,
		AuthToken:                bearerToken,
	}, nil
}

// authoriseChartOperation verified whether the calling user can perform underlying helm operations based on authorization.
func (handler *Handler) authoriseHelmOperation(r *http.Request, authorization portaineree.Authorization) *httperror.HandlerError {
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user authentication token", err}
	}

	if tokenData.Role == portaineree.AdministratorRole {
		return nil
	}

	user, err := handler.dataStore.User().User(tokenData.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find user", err}
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment on request context", err}
	}
	authorizations := user.EndpointAuthorizations[endpoint.ID]
	if !authorizations[authorization] {
		errMsg := "Permission denied to perform helm operation"
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to perform helm operation", errors.New(errMsg)}
	}

	return nil
}

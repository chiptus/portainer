package helm

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/kubernetes"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/pkg/libhelm"
	"github.com/portainer/portainer/pkg/libhelm/options"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	requestBouncer           security.BouncerService
	dataStore                dataservices.DataStore
	jwtService               portainer.JWTService
	kubeClusterAccessService kubernetes.KubeClusterAccessService
	kubernetesDeployer       portaineree.KubernetesDeployer
	helmPackageManager       libhelm.HelmPackageManager
	userActivityService      portaineree.UserActivityService
	fileService              portaineree.FileService
}

// NewHandler creates a handler to manage endpoint group operations.
func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore, jwtService portainer.JWTService, kubernetesDeployer portaineree.KubernetesDeployer, helmPackageManager libhelm.HelmPackageManager, kubeClusterAccessService kubernetes.KubeClusterAccessService, userActivityService portaineree.UserActivityService, fileService portaineree.FileService) *Handler {
	h := &Handler{
		Router:                   mux.NewRouter(),
		requestBouncer:           bouncer,
		dataStore:                dataStore,
		jwtService:               jwtService,
		kubeClusterAccessService: kubeClusterAccessService,
		kubernetesDeployer:       kubernetesDeployer,
		helmPackageManager:       helmPackageManager,
		userActivityService:      userActivityService,
		fileService:              fileService,
	}

	// EE-6176 doc: RBAC performed inside handlers with direct checks on user's Authorizations
	// EE-6176 doc: see authoriseHelmOperation() function below
	h.Use(middlewares.WithEndpoint(dataStore.Endpoint(), "id"),
		bouncer.AuthenticatedAccess,
		useractivity.LogUserActivity(h.userActivityService))

	// `helm list -o json`
	h.Handle("/{id}/kubernetes/helm", httperror.LoggerHandler(h.helmList)).Methods(http.MethodGet)

	// `helm delete RELEASE_NAME`
	h.Handle("/{id}/kubernetes/helm/{release}", httperror.LoggerHandler(h.helmDelete)).Methods(http.MethodDelete)

	// `helm install [NAME] [CHART] flags`
	h.Handle("/{id}/kubernetes/helm", httperror.LoggerHandler(h.helmInstall)).Methods(http.MethodPost)

	// Deprecated
	h.Handle("/{id}/kubernetes/helm/repositories", httperror.LoggerHandler(h.userGetHelmRepos)).Methods(http.MethodGet)
	h.Handle("/{id}/kubernetes/helm/repositories", httperror.LoggerHandler(h.userCreateHelmRepo)).Methods(http.MethodPost)

	return h
}

// NewTemplateHandler creates a template handler to manage environment(endpoint) group operations.
func NewTemplateHandler(bouncer security.BouncerService, helmPackageManager libhelm.HelmPackageManager, fileService portaineree.FileService) *Handler {
	h := &Handler{
		Router:             mux.NewRouter(),
		helmPackageManager: helmPackageManager,
		requestBouncer:     bouncer,
		fileService:        fileService,
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
		return nil, httperror.NotFound("Unable to find an environment on request context", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	bearerToken, _, err := handler.jwtService.GenerateToken(tokenData)
	if err != nil {
		return nil, httperror.Unauthorized("Unauthorized", err)
	}

	sslSettings, err := handler.dataStore.SSLSettings().Settings()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	hostURL := "localhost"
	if !sslSettings.SelfSigned {
		hostURL = r.Host
	}

	kubeConfigInternal := handler.kubeClusterAccessService.GetClusterDetails(hostURL, endpoint.ID, true)

	return &options.KubernetesClusterAccess{
		ClusterServerURL:         kubeConfigInternal.ClusterServerURL,
		CertificateAuthorityFile: kubeConfigInternal.CertificateAuthorityFile,
		AuthToken:                bearerToken,
	}, nil
}

// authoriseChartOperation verified whether the calling user can perform underlying helm operations based on authorization.
//
// EE-6176 TODO later: move this check to RBAC layer performed before the handler exec
func (handler *Handler) authoriseHelmOperation(r *http.Request, authorization portainer.Authorization) *httperror.HandlerError {
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if security.IsAdminOrEdgeAdmin(tokenData.Role) {
		return nil
	}

	user, err := handler.dataStore.User().Read(tokenData.ID)
	if err != nil {
		return httperror.NotFound("Unable to find user", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.NotFound("Unable to find an environment on request context", err)
	}
	authorizations := user.EndpointAuthorizations[endpoint.ID]
	if !authorizations[authorization] {
		errMsg := "Permission denied to perform helm operation"
		return httperror.Forbidden("Permission denied to perform helm operation", errors.New(errMsg))
	}

	return nil
}

package registries

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/registryproxy"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
)

func hideFields(registry *portaineree.Registry, hideAccesses bool) {
	registry.Password = ""
	registry.ManagementConfiguration = nil
	if hideAccesses {
		registry.RegistryAccesses = nil
	}
}

// Handler is the HTTP handler used to handle registry operations.
type Handler struct {
	*mux.Router
	requestBouncer       accessGuard
	registryProxyService *registryproxy.Service

	DataStore           dataservices.DataStore
	FileService         portainer.FileService
	ProxyManager        *proxy.Manager
	userActivityService portaineree.UserActivityService
	K8sClientFactory    *cli.ClientFactory
}

// NewHandler creates a handler to manage registry operations.
func NewHandler(bouncer accessGuard, userActivityService portaineree.UserActivityService) *Handler {
	h := newHandler(bouncer, userActivityService)
	h.initRouter(bouncer)

	return h
}

func newHandler(bouncer accessGuard, userActivityService portaineree.UserActivityService) *Handler {
	return &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		registryProxyService: registryproxy.NewService(userActivityService),
		userActivityService:  userActivityService,
	}
}

func (h *Handler) initRouter(bouncer accessGuard) {
	h.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/registries", httperror.LoggerHandler(h.registryCreate)).Methods(http.MethodPost)
	h.Handle("/registries", httperror.LoggerHandler(h.registryList)).Methods(http.MethodGet)
	h.Handle("/registries/{id}", httperror.LoggerHandler(h.registryInspect)).Methods(http.MethodGet)
	h.Handle("/registries/{id}", httperror.LoggerHandler(h.registryUpdate)).Methods(http.MethodPut)
	h.Handle("/registries/{id}/configure", httperror.LoggerHandler(h.registryConfigure)).Methods(http.MethodPost)
	h.Handle("/registries/{id}", httperror.LoggerHandler(h.registryDelete)).Methods(http.MethodDelete)
	h.Handle("/registries/{id}/ecr/repositories/{repositoryName}", httperror.LoggerHandler(h.ecrDeleteRepository)).Methods(http.MethodDelete)
	h.Handle("/registries/{id}/ecr/repositories/{repositoryName}/tags", httperror.LoggerHandler(h.ecrDeleteTags)).Methods(http.MethodDelete)
	h.PathPrefix("/registries/{id}/v2").Handler(httperror.LoggerHandler(h.proxyRequestsToRegistryAPI))
	h.PathPrefix("/registries/{id}/proxies/gitlab").Handler(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithRegistry))
	h.PathPrefix("/registries/proxies/gitlab").Handler(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithoutRegistry))
}

type accessGuard interface {
	AdminAccess(h http.Handler) http.Handler
}

func (handler *Handler) registriesHaveSameURLAndCredentials(r1, r2 *portaineree.Registry) bool {
	hasSameUrl := r1.URL == r2.URL
	hasSameCredentials := r1.Authentication == r2.Authentication && (!r1.Authentication || (r1.Authentication && r1.Username == r2.Username))

	if r1.Type != portaineree.GitlabRegistry || r2.Type != portaineree.GitlabRegistry {
		return hasSameUrl && hasSameCredentials
	}

	return hasSameUrl && hasSameCredentials && r1.Gitlab.ProjectPath == r2.Gitlab.ProjectPath
}

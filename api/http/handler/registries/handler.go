package registries

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/registryproxy"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

const (
	handlerActivityContext = "Portainer"
)

func hideFields(registry *portainer.Registry, hideAccesses bool) {
	registry.Password = ""
	registry.ManagementConfiguration = nil
	if hideAccesses {
		registry.RegistryAccesses = nil
	}
}

// Handler is the HTTP handler used to handle registry operations.
type Handler struct {
	*mux.Router
	requestBouncer       *security.RequestBouncer
	registryProxyService *registryproxy.Service

	DataStore         portainer.DataStore
	FileService       portainer.FileService
	ProxyManager      *proxy.Manager
	UserActivityStore portainer.UserActivityStore
	K8sClientFactory  *cli.ClientFactory
}

// NewHandler creates a handler to manage registry operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityStore portainer.UserActivityStore) *Handler {
	h := newHandler(bouncer, userActivityStore)
	h.initRouter(bouncer)

	return h
}

func newHandler(bouncer *security.RequestBouncer, userActivityStore portainer.UserActivityStore) *Handler {
	return &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		registryProxyService: registryproxy.NewService(userActivityStore),
		UserActivityStore:    userActivityStore,
	}
}

func (h *Handler) initRouter(bouncer accessGuard) {
	h.Handle("/registries",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryCreate))).Methods(http.MethodPost) // admin
	h.Handle("/registries",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryList))).Methods(http.MethodGet) // admin
	h.Handle("/registries/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryInspect))).Methods(http.MethodGet) // filtered
	h.Handle("/registries/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryUpdate))).Methods(http.MethodPut) // admin
	h.Handle("/registries/{id}/configure",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryConfigure))).Methods(http.MethodPost) // admin
	h.Handle("/registries/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryDelete))).Methods(http.MethodDelete) // admin
	h.PathPrefix("/registries/{id}/v2").Handler(
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.proxyRequestsToRegistryAPI))) // admin
	h.PathPrefix("/registries/{id}/proxies/gitlab").Handler(
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithRegistry))) // admin
	h.PathPrefix("/registries/proxies/gitlab").Handler(
		bouncer.AdminAccess(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithoutRegistry)))
}

type accessGuard interface {
	AdminAccess(h http.Handler) http.Handler
	RestrictedAccess(h http.Handler) http.Handler
	AuthenticatedAccess(h http.Handler) http.Handler
}

func (handler *Handler) registriesHaveSameURLAndCredentials(r1, r2 *portainer.Registry) bool {
	hasSameUrl := r1.URL == r2.URL
	hasSameCredentials := r1.Authentication == r2.Authentication && (!r1.Authentication || (r1.Authentication && r1.Username == r2.Username))

	if r1.Type != portainer.GitlabRegistry || r2.Type != portainer.GitlabRegistry {
		return hasSameUrl && hasSameCredentials
	}

	return hasSameUrl && hasSameCredentials && r1.Gitlab.ProjectPath == r2.Gitlab.ProjectPath
}

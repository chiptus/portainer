package registries

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/registryproxy"
	"github.com/portainer/portainer-ee/api/http/security"
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

func (handler *Handler) initRouter(bouncer accessGuard) {
	logUserActivity := useractivity.LogUserActivity(handler.userActivityService)

	adminRouter := handler.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, logUserActivity)

	authenticatedRouter := handler.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, logUserActivity)

	adminRouter.Handle("/registries", httperror.LoggerHandler(handler.registryCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/registries", httperror.LoggerHandler(handler.registryList)).Methods(http.MethodGet)
	adminRouter.Handle("/registries/{id}", httperror.LoggerHandler(handler.registryUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/registries/{id}/configure", httperror.LoggerHandler(handler.registryConfigure)).Methods(http.MethodPost)
	adminRouter.Handle("/registries/{id}", httperror.LoggerHandler(handler.registryDelete)).Methods(http.MethodDelete)

	authenticatedRouter.Handle("/registries/{id}", httperror.LoggerHandler(handler.registryInspect)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/registries/{id}/ecr/repositories/{repositoryName}", httperror.LoggerHandler(handler.ecrDeleteRepository)).Methods(http.MethodDelete)
	authenticatedRouter.Handle("/registries/{id}/ecr/repositories/{repositoryName}/tags", httperror.LoggerHandler(handler.ecrDeleteTags)).Methods(http.MethodDelete)
	authenticatedRouter.PathPrefix("/registries/{id}/v2").Handler(httperror.LoggerHandler(handler.proxyRequestsToRegistryAPI))
	authenticatedRouter.PathPrefix("/registries/{id}/proxies/gitlab").Handler(httperror.LoggerHandler(handler.proxyRequestsToGitlabAPIWithRegistry))
	authenticatedRouter.PathPrefix("/registries/proxies/gitlab").Handler(httperror.LoggerHandler(handler.proxyRequestsToGitlabAPIWithoutRegistry))
}

type accessGuard interface {
	AdminAccess(h http.Handler) http.Handler
	AuthenticatedAccess(h http.Handler) http.Handler
	AuthorizedEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint, authorizationCheck bool) error
}

func (handler *Handler) registriesHaveSameURLAndCredentials(r1, r2 *portaineree.Registry) bool {
	hasSameUrl := r1.URL == r2.URL
	hasSameCredentials := r1.Authentication == r2.Authentication && (!r1.Authentication || (r1.Authentication && r1.Username == r2.Username))

	if r1.Type != portaineree.GitlabRegistry || r2.Type != portaineree.GitlabRegistry {
		return hasSameUrl && hasSameCredentials
	}

	return hasSameUrl && hasSameCredentials && r1.Gitlab.ProjectPath == r2.Gitlab.ProjectPath
}

func (handler *Handler) userHasRegistryAccess(r *http.Request) (hasAccess bool, isAdmin bool, err error) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return false, false, err
	}

	if securityContext.IsAdmin {
		return true, true, nil
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return false, false, err
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return false, false, err
	}
	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		return false, false, err
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err == security.ErrAuthorizationRequired {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	_, isEndpointAdmin := user.EndpointAuthorizations[portaineree.EndpointID(endpointID)][portaineree.EndpointResourcesAccess]

	return true, isEndpointAdmin, nil
}

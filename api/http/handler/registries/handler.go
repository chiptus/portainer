package registries

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/registryproxy"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/pendingactions"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/gorilla/mux"
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
	requestBouncer       security.BouncerService
	registryProxyService *registryproxy.Service

	DataStore             dataservices.DataStore
	FileService           portainer.FileService
	ProxyManager          *proxy.Manager
	userActivityService   portaineree.UserActivityService
	K8sClientFactory      *cli.ClientFactory
	PendingActionsService *pendingactions.PendingActionsService
}

// NewHandler creates a handler to manage registry operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		requestBouncer:       bouncer,
		registryProxyService: registryproxy.NewService(userActivityService),
		userActivityService:  userActivityService,
	}

	logUserActivity := useractivity.LogUserActivity(userActivityService)

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess, logUserActivity)
	adminRouter.Handle("/registries", httperror.LoggerHandler(h.registryCreate)).Methods(http.MethodPost)     // admin
	adminRouter.Handle("/registries/{id}", httperror.LoggerHandler(h.registryUpdate)).Methods(http.MethodPut) // admin
	adminRouter.Handle("/registries/{id}/configure", httperror.LoggerHandler(h.registryConfigure)).Methods(http.MethodPost)
	adminRouter.Handle("/registries/{id}", httperror.LoggerHandler(h.registryDelete)).Methods(http.MethodDelete)

	edgeAdminRouter := h.NewRoute().Subrouter()
	edgeAdminRouter.Use(bouncer.AdminAccess, logUserActivity)
	edgeAdminRouter.Handle("/registries", httperror.LoggerHandler(h.registryList)).Methods(http.MethodGet) // used in edge stack views + update schedules

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, logUserActivity)
	authenticatedRouter.Handle("/registries/{id}", httperror.LoggerHandler(h.registryInspect)).Methods(http.MethodGet)                                          // all roles
	authenticatedRouter.Handle("/registries/{id}/ecr/repositories/{repositoryName}", httperror.LoggerHandler(h.ecrDeleteRepository)).Methods(http.MethodDelete) // all roles
	authenticatedRouter.Handle("/registries/{id}/ecr/repositories/{repositoryName}/tags", httperror.LoggerHandler(h.ecrDeleteTags)).Methods(http.MethodDelete)  // all roles
	authenticatedRouter.PathPrefix("/registries/{id}/v2").Handler(httperror.LoggerHandler(h.proxyRequestsToRegistryAPI))
	authenticatedRouter.PathPrefix("/registries/{id}/proxies/gitlab").Handler(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithRegistry))
	authenticatedRouter.PathPrefix("/registries/proxies/gitlab").Handler(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithoutRegistry))

	return h
}

func (handler *Handler) registriesHaveSameURLAndCredentials(r1, r2 *portaineree.Registry) bool {
	hasSameUrl := r1.URL == r2.URL
	hasSameCredentials := r1.Authentication == r2.Authentication && (!r1.Authentication || (r1.Authentication && r1.Username == r2.Username))

	if r1.Type == portaineree.GithubRegistry && r2.Type == portaineree.GithubRegistry {
		org1 := ""
		if r1.Github.UseOrganisation {
			org1 = r1.Github.OrganisationName
		}

		org2 := ""
		if r2.Github.UseOrganisation {
			org2 = r2.Github.OrganisationName
		}

		hasSameOrg := org1 == org2

		return hasSameUrl && hasSameCredentials && hasSameOrg
	}

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

	if security.IsAdminOrEdgeAdminContext(securityContext) {
		return true, true, nil
	}

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return false, false, err
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return false, false, err
	}
	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err != nil {
		return false, false, err
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if errors.Is(err, security.ErrAuthorizationRequired) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	_, isEndpointAdmin := user.EndpointAuthorizations[portainer.EndpointID(endpointID)][portaineree.EndpointResourcesAccess]

	return true, isEndpointAdmin, nil
}

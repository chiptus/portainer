package websocket

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Handler is the HTTP handler used to handle websocket operations.
type Handler struct {
	*mux.Router
	DataStore                   dataservices.DataStore
	SignatureService            portainer.DigitalSignatureService
	ReverseTunnelService        portaineree.ReverseTunnelService
	KubernetesClientFactory     *cli.ClientFactory
	authorizationService        *authorization.Service
	requestBouncer              security.BouncerService
	connectionUpgrader          websocket.Upgrader
	userActivityService         portaineree.UserActivityService
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
}

// NewHandler creates a handler to manage websocket operations.
func NewHandler(kubernetesTokenCacheManager *kubernetes.TokenCacheManager, bouncer security.BouncerService, authorizationService *authorization.Service, dataStore dataservices.DataStore, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:                      mux.NewRouter(),
		connectionUpgrader:          websocket.Upgrader{},
		requestBouncer:              bouncer,
		authorizationService:        authorizationService,
		kubernetesTokenCacheManager: kubernetesTokenCacheManager,
		userActivityService:         userActivityService,
		DataStore:                   dataStore,
	}

	activityLogging := useractivity.LogUserActivityWithContext(h.userActivityService, middlewares.FindInQuery(dataStore.Endpoint(), "endpointId"))

	// EE-6176 doc: RBAC performed inside handlers with bouncer.AuthorizedEndpointOperation()
	h.Use(bouncer.AuthenticatedAccess, activityLogging)

	h.PathPrefix("/websocket/exec").Handler(httperror.LoggerHandler(h.websocketExec))
	h.PathPrefix("/websocket/attach").Handler(httperror.LoggerHandler(h.websocketAttach))
	h.PathPrefix("/websocket/pod").Handler(httperror.LoggerHandler(h.websocketPodExec))
	h.PathPrefix("/websocket/kubernetes-shell").Handler(httperror.LoggerHandler(h.websocketShellPodExec))
	h.PathPrefix("/websocket/microk8s-shell").Handler(httperror.LoggerHandler(h.websocketMicrok8sShell))
	return h
}

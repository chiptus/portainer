package websocket

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	httperror "github.com/portainer/libhttp/error"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

// Handler is the HTTP handler used to handle websocket operations.
type Handler struct {
	*mux.Router
	dataStore                   dataservices.DataStore
	SignatureService            portaineree.DigitalSignatureService
	ReverseTunnelService        portaineree.ReverseTunnelService
	KubernetesClientFactory     *cli.ClientFactory
	authorizationService        *authorization.Service
	requestBouncer              *security.RequestBouncer
	connectionUpgrader          websocket.Upgrader
	userActivityService         portaineree.UserActivityService
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
}

// NewHandler creates a handler to manage websocket operations.
func NewHandler(kubernetesTokenCacheManager *kubernetes.TokenCacheManager, bouncer *security.RequestBouncer, authorizationService *authorization.Service, dataStore dataservices.DataStore, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:                      mux.NewRouter(),
		connectionUpgrader:          websocket.Upgrader{},
		requestBouncer:              bouncer,
		authorizationService:        authorizationService,
		kubernetesTokenCacheManager: kubernetesTokenCacheManager,
		userActivityService:         userActivityService,
		dataStore:                   dataStore,
	}

	activityLogging := useractivity.LogUserActivityWithContext(h.userActivityService, middlewares.FindInQuery(dataStore.Endpoint(), "endpointId"))

	h.Use(bouncer.AuthenticatedAccess, activityLogging)

	h.PathPrefix("/websocket/exec").Handler(httperror.LoggerHandler(h.websocketExec))
	h.PathPrefix("/websocket/attach").Handler(httperror.LoggerHandler(h.websocketAttach))
	h.PathPrefix("/websocket/pod").Handler(httperror.LoggerHandler(h.websocketPodExec))
	h.PathPrefix("/websocket/kubernetes-shell").Handler(httperror.LoggerHandler(h.websocketShellPodExec))
	return h
}

package websocket

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	httperror "github.com/portainer/libhttp/error"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/middlewares"
	"github.com/portainer/portainer/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Handler is the HTTP handler used to handle websocket operations.
type Handler struct {
	*mux.Router
	dataStore                   portainer.DataStore
	SignatureService            portainer.DigitalSignatureService
	ReverseTunnelService        portainer.ReverseTunnelService
	KubernetesClientFactory     *cli.ClientFactory
	authorizationService        *authorization.Service
	requestBouncer              *security.RequestBouncer
	connectionUpgrader          websocket.Upgrader
	userActivityService         portainer.UserActivityService
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
}

// NewHandler creates a handler to manage websocket operations.
func NewHandler(kubernetesTokenCacheManager *kubernetes.TokenCacheManager, bouncer *security.RequestBouncer, authorizationService *authorization.Service, dataStore portainer.DataStore, userActivityService portainer.UserActivityService) *Handler {
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

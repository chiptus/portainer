package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/adminmonitor"
	backupOps "github.com/portainer/portainer/api/backup"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/docker"
	"github.com/portainer/portainer/api/http/handler"
	"github.com/portainer/portainer/api/http/handler/auth"
	"github.com/portainer/portainer/api/http/handler/backup"
	"github.com/portainer/portainer/api/http/handler/customtemplates"
	"github.com/portainer/portainer/api/http/handler/edgegroups"
	"github.com/portainer/portainer/api/http/handler/edgejobs"
	"github.com/portainer/portainer/api/http/handler/edgestacks"
	"github.com/portainer/portainer/api/http/handler/edgetemplates"
	"github.com/portainer/portainer/api/http/handler/endpointedge"
	"github.com/portainer/portainer/api/http/handler/endpointgroups"
	"github.com/portainer/portainer/api/http/handler/endpointproxy"
	"github.com/portainer/portainer/api/http/handler/endpoints"
	"github.com/portainer/portainer/api/http/handler/file"
	"github.com/portainer/portainer/api/http/handler/ldap"
	"github.com/portainer/portainer/api/http/handler/licenses"
	"github.com/portainer/portainer/api/http/handler/motd"
	"github.com/portainer/portainer/api/http/handler/registries"
	"github.com/portainer/portainer/api/http/handler/resourcecontrols"
	"github.com/portainer/portainer/api/http/handler/roles"
	"github.com/portainer/portainer/api/http/handler/settings"
	"github.com/portainer/portainer/api/http/handler/stacks"
	"github.com/portainer/portainer/api/http/handler/status"
	"github.com/portainer/portainer/api/http/handler/tags"
	"github.com/portainer/portainer/api/http/handler/teammemberships"
	"github.com/portainer/portainer/api/http/handler/teams"
	"github.com/portainer/portainer/api/http/handler/templates"
	"github.com/portainer/portainer/api/http/handler/upload"
	"github.com/portainer/portainer/api/http/handler/useractivity"
	"github.com/portainer/portainer/api/http/handler/users"
	"github.com/portainer/portainer/api/http/handler/webhooks"
	"github.com/portainer/portainer/api/http/handler/websocket"
	"github.com/portainer/portainer/api/http/offlinegate"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Server implements the portainer.Server interface
type Server struct {
	AuthorizationService        *authorization.Service
	BindAddress                 string
	AssetsPath                  string
	Status                      *portainer.Status
	ReverseTunnelService        portainer.ReverseTunnelService
	ComposeStackManager         portainer.ComposeStackManager
	CryptoService               portainer.CryptoService
	LicenseService              portainer.LicenseService
	SignatureService            portainer.DigitalSignatureService
	SnapshotService             portainer.SnapshotService
	FileService                 portainer.FileService
	DataStore                   portainer.DataStore
	GitService                  portainer.GitService
	JWTService                  portainer.JWTService
	LDAPService                 portainer.LDAPService
	OAuthService                portainer.OAuthService
	SwarmStackManager           portainer.SwarmStackManager
	UserActivityStore           portainer.UserActivityStore
	ProxyManager                *proxy.Manager
	KubernetesTokenCacheManager *kubernetes.TokenCacheManager
	Handler                     *handler.Handler
	SSL                         bool
	SSLCert                     string
	SSLKey                      string
	DockerClientFactory         *docker.ClientFactory
	KubernetesClientFactory     *cli.ClientFactory
	KubernetesDeployer          portainer.KubernetesDeployer
	ShutdownCtx                 context.Context
	ShutdownTrigger             context.CancelFunc
}

// Start starts the HTTP server
func (server *Server) Start() error {
	server.AuthorizationService.RegisterEventHandler("kubernetesTokenCacheManager", server.KubernetesTokenCacheManager)

	requestBouncer := security.NewRequestBouncer(server.DataStore, server.LicenseService, server.JWTService)

	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	offlineGate := offlinegate.NewOfflineGate()

	var authHandler = auth.NewHandler(requestBouncer, rateLimiter)
	authHandler.AuthorizationService = server.AuthorizationService
	authHandler.DataStore = server.DataStore
	authHandler.CryptoService = server.CryptoService
	authHandler.JWTService = server.JWTService
	authHandler.LDAPService = server.LDAPService
	authHandler.LicenseService = server.LicenseService
	authHandler.ProxyManager = server.ProxyManager
	authHandler.KubernetesTokenCacheManager = server.KubernetesTokenCacheManager
	authHandler.OAuthService = server.OAuthService
	authHandler.UserActivityStore = server.UserActivityStore

	adminMonitor := adminmonitor.New(5*time.Minute, server.DataStore, server.ShutdownCtx)
	adminMonitor.Start()

	backupScheduler := backupOps.NewBackupScheduler(offlineGate, server.DataStore, server.UserActivityStore, server.FileService.GetDatastorePath())
	if err := backupScheduler.Start(); err != nil {
		return errors.Wrap(err, "failed to start backup scheduler")
	}
	var backupHandler = backup.NewHandler(requestBouncer, server.DataStore, server.UserActivityStore, offlineGate, server.FileService.GetDatastorePath(), backupScheduler, server.ShutdownTrigger, adminMonitor)

	var roleHandler = roles.NewHandler(requestBouncer)
	roleHandler.DataStore = server.DataStore

	var customTemplatesHandler = customtemplates.NewHandler(requestBouncer)
	customTemplatesHandler.DataStore = server.DataStore
	customTemplatesHandler.FileService = server.FileService
	customTemplatesHandler.GitService = server.GitService
	customTemplatesHandler.UserActivityStore = server.UserActivityStore

	var edgeGroupsHandler = edgegroups.NewHandler(requestBouncer)
	edgeGroupsHandler.DataStore = server.DataStore
	edgeGroupsHandler.UserActivityStore = server.UserActivityStore

	var edgeJobsHandler = edgejobs.NewHandler(requestBouncer)
	edgeJobsHandler.DataStore = server.DataStore
	edgeJobsHandler.FileService = server.FileService
	edgeJobsHandler.ReverseTunnelService = server.ReverseTunnelService
	edgeJobsHandler.UserActivityStore = server.UserActivityStore

	var edgeStacksHandler = edgestacks.NewHandler(requestBouncer)
	edgeStacksHandler.DataStore = server.DataStore
	edgeStacksHandler.FileService = server.FileService
	edgeStacksHandler.GitService = server.GitService
	edgeStacksHandler.UserActivityStore = server.UserActivityStore

	var edgeTemplatesHandler = edgetemplates.NewHandler(requestBouncer)
	edgeTemplatesHandler.DataStore = server.DataStore

	var endpointHandler = endpoints.NewHandler(requestBouncer)
	endpointHandler.AuthorizationService = server.AuthorizationService
	endpointHandler.DataStore = server.DataStore
	endpointHandler.FileService = server.FileService
	endpointHandler.ProxyManager = server.ProxyManager
	endpointHandler.SnapshotService = server.SnapshotService
	endpointHandler.ReverseTunnelService = server.ReverseTunnelService
	endpointHandler.K8sClientFactory = server.KubernetesClientFactory
	endpointHandler.ComposeStackManager = server.ComposeStackManager
	endpointHandler.DockerClientFactory = server.DockerClientFactory
	endpointHandler.UserActivityStore = server.UserActivityStore

	var endpointEdgeHandler = endpointedge.NewHandler(requestBouncer)
	endpointEdgeHandler.DataStore = server.DataStore
	endpointEdgeHandler.FileService = server.FileService
	endpointEdgeHandler.ReverseTunnelService = server.ReverseTunnelService

	var endpointGroupHandler = endpointgroups.NewHandler(requestBouncer)
	endpointGroupHandler.AuthorizationService = server.AuthorizationService
	endpointGroupHandler.DataStore = server.DataStore
	endpointGroupHandler.UserActivityStore = server.UserActivityStore

	var endpointProxyHandler = endpointproxy.NewHandler(requestBouncer)
	endpointProxyHandler.DataStore = server.DataStore
	endpointProxyHandler.ProxyManager = server.ProxyManager
	endpointProxyHandler.ReverseTunnelService = server.ReverseTunnelService

	var licenseHandler = licenses.NewHandler(requestBouncer)
	licenseHandler.LicenseService = server.LicenseService
	licenseHandler.UserActivityStore = server.UserActivityStore

	var fileHandler = file.NewHandler(filepath.Join(server.AssetsPath, "public"))

	var ldapHandler = ldap.NewHandler(requestBouncer)
	ldapHandler.DataStore = server.DataStore
	ldapHandler.FileService = server.FileService
	ldapHandler.LDAPService = server.LDAPService

	var motdHandler = motd.NewHandler(requestBouncer)

	var registryHandler = registries.NewHandler(requestBouncer, server.UserActivityStore)
	registryHandler.DataStore = server.DataStore
	registryHandler.FileService = server.FileService
	registryHandler.ProxyManager = server.ProxyManager
	registryHandler.K8sClientFactory = server.KubernetesClientFactory

	var resourceControlHandler = resourcecontrols.NewHandler(requestBouncer)
	resourceControlHandler.DataStore = server.DataStore
	resourceControlHandler.UserActivityStore = server.UserActivityStore

	var settingsHandler = settings.NewHandler(requestBouncer)
	settingsHandler.AuthorizationService = server.AuthorizationService
	settingsHandler.DataStore = server.DataStore
	settingsHandler.FileService = server.FileService
	settingsHandler.JWTService = server.JWTService
	settingsHandler.LDAPService = server.LDAPService
	settingsHandler.SnapshotService = server.SnapshotService
	settingsHandler.UserActivityStore = server.UserActivityStore

	var stackHandler = stacks.NewHandler(requestBouncer)
	stackHandler.DataStore = server.DataStore
	stackHandler.DockerClientFactory = server.DockerClientFactory
	stackHandler.FileService = server.FileService
	stackHandler.SwarmStackManager = server.SwarmStackManager
	stackHandler.ComposeStackManager = server.ComposeStackManager
	stackHandler.KubernetesDeployer = server.KubernetesDeployer
	stackHandler.GitService = server.GitService
	stackHandler.DockerClientFactory = server.DockerClientFactory
	stackHandler.KubernetesClientFactory = server.KubernetesClientFactory
	stackHandler.AuthorizationService = server.AuthorizationService
	stackHandler.UserActivityStore = server.UserActivityStore

	var statusHandler = status.NewHandler(requestBouncer, server.Status)
	statusHandler.DataStore = server.DataStore

	var tagHandler = tags.NewHandler(requestBouncer)
	tagHandler.DataStore = server.DataStore
	tagHandler.UserActivityStore = server.UserActivityStore

	var teamHandler = teams.NewHandler(requestBouncer)
	teamHandler.AuthorizationService = server.AuthorizationService
	teamHandler.DataStore = server.DataStore
	teamHandler.K8sClientFactory = server.KubernetesClientFactory
	teamHandler.UserActivityStore = server.UserActivityStore

	var teamMembershipHandler = teammemberships.NewHandler(requestBouncer)
	teamMembershipHandler.AuthorizationService = server.AuthorizationService
	teamMembershipHandler.DataStore = server.DataStore
	teamMembershipHandler.UserActivityStore = server.UserActivityStore

	var templatesHandler = templates.NewHandler(requestBouncer)
	templatesHandler.DataStore = server.DataStore
	templatesHandler.FileService = server.FileService
	templatesHandler.GitService = server.GitService

	var uploadHandler = upload.NewHandler(requestBouncer)
	uploadHandler.FileService = server.FileService
	uploadHandler.UserActivityStore = server.UserActivityStore

	var userHandler = users.NewHandler(requestBouncer, rateLimiter)
	userHandler.AuthorizationService = server.AuthorizationService
	userHandler.DataStore = server.DataStore
	userHandler.CryptoService = server.CryptoService
	userHandler.K8sClientFactory = server.KubernetesClientFactory
	userHandler.UserActivityStore = server.UserActivityStore

	var userActivityHandler = useractivity.NewHandler(requestBouncer)
	userActivityHandler.UserActivityStore = server.UserActivityStore

	var websocketHandler = websocket.NewHandler(server.KubernetesTokenCacheManager, requestBouncer, server.AuthorizationService)
	websocketHandler.DataStore = server.DataStore
	websocketHandler.SignatureService = server.SignatureService
	websocketHandler.ReverseTunnelService = server.ReverseTunnelService
	websocketHandler.KubernetesClientFactory = server.KubernetesClientFactory
	websocketHandler.UserActivityStore = server.UserActivityStore

	var webhookHandler = webhooks.NewHandler(requestBouncer)
	webhookHandler.DataStore = server.DataStore
	webhookHandler.DockerClientFactory = server.DockerClientFactory
	webhookHandler.UserActivityStore = server.UserActivityStore

	server.Handler = &handler.Handler{
		RoleHandler:            roleHandler,
		AuthHandler:            authHandler,
		BackupHandler:          backupHandler,
		CustomTemplatesHandler: customTemplatesHandler,
		EdgeGroupsHandler:      edgeGroupsHandler,
		EdgeJobsHandler:        edgeJobsHandler,
		EdgeStacksHandler:      edgeStacksHandler,
		EdgeTemplatesHandler:   edgeTemplatesHandler,
		EndpointGroupHandler:   endpointGroupHandler,
		EndpointHandler:        endpointHandler,
		EndpointEdgeHandler:    endpointEdgeHandler,
		EndpointProxyHandler:   endpointProxyHandler,
		FileHandler:            fileHandler,
		LDAPHandler:            ldapHandler,
		LicenseHandler:         licenseHandler,
		MOTDHandler:            motdHandler,
		RegistryHandler:        registryHandler,
		ResourceControlHandler: resourceControlHandler,
		SettingsHandler:        settingsHandler,
		StatusHandler:          statusHandler,
		StackHandler:           stackHandler,
		TagHandler:             tagHandler,
		TeamHandler:            teamHandler,
		TeamMembershipHandler:  teamMembershipHandler,
		TemplatesHandler:       templatesHandler,
		UploadHandler:          uploadHandler,
		UserHandler:            userHandler,
		UserActivityHandler:    userActivityHandler,
		WebSocketHandler:       websocketHandler,
		WebhookHandler:         webhookHandler,
	}

	httpServer := &http.Server{
		Addr:    server.BindAddress,
		Handler: server.Handler,
	}
	httpServer.Handler = offlineGate.WaitingMiddleware(time.Minute, httpServer.Handler)

	if server.SSL {
		httpServer.TLSConfig = crypto.CreateServerTLSConfiguration()
		return httpServer.ListenAndServeTLS(server.SSLCert, server.SSLKey)
	}

	go server.shutdown(httpServer, backupScheduler)

	return httpServer.ListenAndServe()
}

func (server *Server) shutdown(httpServer *http.Server, backupScheduler *backupOps.BackupScheduler) {
	<-server.ShutdownCtx.Done()

	backupScheduler.Stop()

	log.Println("[DEBUG] Shutting down http server")
	shutdownTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := httpServer.Shutdown(shutdownTimeout)
	if err != nil {
		fmt.Printf("Failed shutdown http server: %s \n", err)
	}
}

package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/portainer/libhelm"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/adminmonitor"
	"github.com/portainer/portainer/api/apikey"
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
	"github.com/portainer/portainer/api/http/handler/helm"
	"github.com/portainer/portainer/api/http/handler/hostmanagement/openamt"
	kubehandler "github.com/portainer/portainer/api/http/handler/kubernetes"
	"github.com/portainer/portainer/api/http/handler/ldap"
	"github.com/portainer/portainer/api/http/handler/licenses"
	"github.com/portainer/portainer/api/http/handler/motd"
	"github.com/portainer/portainer/api/http/handler/registries"
	"github.com/portainer/portainer/api/http/handler/resourcecontrols"
	"github.com/portainer/portainer/api/http/handler/roles"
	"github.com/portainer/portainer/api/http/handler/settings"
	sslhandler "github.com/portainer/portainer/api/http/handler/ssl"
	"github.com/portainer/portainer/api/http/handler/stacks"
	"github.com/portainer/portainer/api/http/handler/status"
	"github.com/portainer/portainer/api/http/handler/storybook"
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
	"github.com/portainer/portainer/api/internal/ssl"
	k8s "github.com/portainer/portainer/api/kubernetes"
	"github.com/portainer/portainer/api/kubernetes/cli"
	"github.com/portainer/portainer/api/scheduler"
	stackdeloyer "github.com/portainer/portainer/api/stacks"
)

// Server implements the portainer.Server interface
type Server struct {
	AuthorizationService        *authorization.Service
	BindAddress                 string
	BindAddressHTTPS            string
	HTTPEnabled                 bool
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
	APIKeyService               apikey.APIKeyService
	OpenAMTService              portainer.OpenAMTService
	JWTService                  portainer.JWTService
	LDAPService                 portainer.LDAPService
	OAuthService                portainer.OAuthService
	SwarmStackManager           portainer.SwarmStackManager
	UserActivityStore           portainer.UserActivityStore
	ProxyManager                *proxy.Manager
	KubernetesTokenCacheManager *kubernetes.TokenCacheManager
	KubeConfigService           k8s.KubeConfigService
	Handler                     *handler.Handler
	SSLService                  *ssl.Service
	DockerClientFactory         *docker.ClientFactory
	KubernetesClientFactory     *cli.ClientFactory
	KubernetesDeployer          portainer.KubernetesDeployer
	HelmPackageManager          libhelm.HelmPackageManager
	Scheduler                   *scheduler.Scheduler
	ShutdownCtx                 context.Context
	ShutdownTrigger             context.CancelFunc
	StackDeployer               stackdeloyer.StackDeployer
	BaseURL                     string
}

// Start starts the HTTP server
func (server *Server) Start() error {
	server.AuthorizationService.RegisterEventHandler("kubernetesTokenCacheManager", server.KubernetesTokenCacheManager)

	requestBouncer := security.NewRequestBouncer(server.DataStore, server.LicenseService, server.JWTService, server.APIKeyService)

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
	edgeStacksHandler.KubernetesDeployer = server.KubernetesDeployer
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
	endpointHandler.BindAddress = server.BindAddress
	endpointHandler.BindAddressHTTPS = server.BindAddressHTTPS

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

	var kubernetesHandler = kubehandler.NewHandler(requestBouncer, server.DataStore, server.BaseURL)
	kubernetesHandler.AuthorizationService = server.AuthorizationService
	kubernetesHandler.KubernetesClientFactory = server.KubernetesClientFactory
	kubernetesHandler.UserActivityStore = server.UserActivityStore
	kubernetesHandler.JwtService = server.JWTService

	var licenseHandler = licenses.NewHandler(requestBouncer)
	licenseHandler.LicenseService = server.LicenseService
	licenseHandler.UserActivityStore = server.UserActivityStore

	var fileHandler = file.NewHandler(filepath.Join(server.AssetsPath, "public"))

	var endpointHelmHandler = helm.NewHandler(requestBouncer, server.DataStore, server.JWTService, server.KubernetesDeployer, server.HelmPackageManager, server.KubeConfigService, server.UserActivityStore)

	var helmTemplatesHandler = helm.NewTemplateHandler(requestBouncer, server.HelmPackageManager)

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

	var sslHandler = sslhandler.NewHandler(requestBouncer)
	sslHandler.SSLService = server.SSLService

	openAMTHandler, err := openamt.NewHandler(requestBouncer, server.DataStore)
	if err != nil {
		return err
	}
	if openAMTHandler != nil {
		openAMTHandler.OpenAMTService = server.OpenAMTService
		openAMTHandler.DataStore = server.DataStore
	}

	var stackHandler = stacks.NewHandler(requestBouncer)
	stackHandler.DockerClientFactory = server.DockerClientFactory
	stackHandler.FileService = server.FileService
	stackHandler.DataStore = server.DataStore
	stackHandler.KubernetesDeployer = server.KubernetesDeployer
	stackHandler.GitService = server.GitService
	stackHandler.DockerClientFactory = server.DockerClientFactory
	stackHandler.KubernetesClientFactory = server.KubernetesClientFactory
	stackHandler.AuthorizationService = server.AuthorizationService
	stackHandler.UserActivityStore = server.UserActivityStore
	stackHandler.Scheduler = server.Scheduler
	stackHandler.SwarmStackManager = server.SwarmStackManager
	stackHandler.ComposeStackManager = server.ComposeStackManager
	stackHandler.StackDeployer = server.StackDeployer

	var statusHandler = status.NewHandler(requestBouncer, server.Status)
	statusHandler.DataStore = server.DataStore

	var storybookHandler = storybook.NewHandler(server.AssetsPath)

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

	var userHandler = users.NewHandler(requestBouncer, rateLimiter, server.APIKeyService)
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
		EndpointHelmHandler:    endpointHelmHandler,
		EndpointEdgeHandler:    endpointEdgeHandler,
		EndpointProxyHandler:   endpointProxyHandler,
		HelmTemplatesHandler:   helmTemplatesHandler,
		KubernetesHandler:      kubernetesHandler,
		FileHandler:            fileHandler,
		LDAPHandler:            ldapHandler,
		LicenseHandler:         licenseHandler,
		MOTDHandler:            motdHandler,
		OpenAMTHandler:         openAMTHandler,
		RegistryHandler:        registryHandler,
		ResourceControlHandler: resourceControlHandler,
		SettingsHandler:        settingsHandler,
		SSLHandler:             sslHandler,
		StatusHandler:          statusHandler,
		StackHandler:           stackHandler,
		StorybookHandler:       storybookHandler,
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

	handler := offlineGate.WaitingMiddleware(time.Minute, server.Handler)

	if server.HTTPEnabled {
		go func() {
			log.Printf("[INFO] [http,server] [message: starting HTTP server on port %s]", server.BindAddress)
			httpServer := &http.Server{
				Addr:    server.BindAddress,
				Handler: handler,
			}

			go shutdown(server.ShutdownCtx, httpServer, backupScheduler)
			err := httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("[ERROR] [message: http server failed] [error: %s]", err)
			}
		}()
	}

	log.Printf("[INFO] [http,server] [message: starting HTTPS server on port %s]", server.BindAddressHTTPS)
	httpsServer := &http.Server{
		Addr:    server.BindAddressHTTPS,
		Handler: handler,
	}

	httpsServer.TLSConfig = crypto.CreateServerTLSConfiguration()
	httpsServer.TLSConfig.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return server.SSLService.GetRawCertificate(), nil
	}

	go shutdown(server.ShutdownCtx, httpsServer, backupScheduler)
	return httpsServer.ListenAndServeTLS("", "")
}

func shutdown(shutdownCtx context.Context, httpServer *http.Server, backupScheduler *backupOps.BackupScheduler) {
	<-shutdownCtx.Done()

	backupScheduler.Stop()

	log.Println("[DEBUG] [http,server] [message: shutting down http server]")
	shutdownTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := httpServer.Shutdown(shutdownTimeout)
	if err != nil {
		fmt.Printf("[ERROR] [http,server] [message: failed shutdown http server] [error: %s]", err)
	}
}

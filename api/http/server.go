package http

import (
	"context"
	"crypto/tls"
	"github.com/portainer/portainer-ee/api/docker/client"
	"net/http"
	"path/filepath"
	"time"

	"github.com/portainer/libhelm"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/adminmonitor"
	"github.com/portainer/portainer-ee/api/apikey"
	backupOps "github.com/portainer/portainer-ee/api/backup"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/dataservices/cloudcredential"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/handler"
	"github.com/portainer/portainer-ee/api/http/handler/auth"
	"github.com/portainer/portainer-ee/api/http/handler/backup"
	"github.com/portainer/portainer-ee/api/http/handler/cloudcredentials"
	"github.com/portainer/portainer-ee/api/http/handler/customtemplates"
	dockerhandler "github.com/portainer/portainer-ee/api/http/handler/docker"
	"github.com/portainer/portainer-ee/api/http/handler/edgegroups"
	"github.com/portainer/portainer-ee/api/http/handler/edgejobs"
	"github.com/portainer/portainer-ee/api/http/handler/edgestacks"
	"github.com/portainer/portainer-ee/api/http/handler/edgetemplates"
	"github.com/portainer/portainer-ee/api/http/handler/edgeupdateschedules"
	"github.com/portainer/portainer-ee/api/http/handler/endpointedge"
	"github.com/portainer/portainer-ee/api/http/handler/endpointgroups"
	"github.com/portainer/portainer-ee/api/http/handler/endpointproxy"
	"github.com/portainer/portainer-ee/api/http/handler/endpoints"
	"github.com/portainer/portainer-ee/api/http/handler/file"
	"github.com/portainer/portainer-ee/api/http/handler/gitops"
	"github.com/portainer/portainer-ee/api/http/handler/helm"
	"github.com/portainer/portainer-ee/api/http/handler/hostmanagement/fdo"
	"github.com/portainer/portainer-ee/api/http/handler/hostmanagement/openamt"
	"github.com/portainer/portainer-ee/api/http/handler/kaas"
	kubehandler "github.com/portainer/portainer-ee/api/http/handler/kubernetes"
	"github.com/portainer/portainer-ee/api/http/handler/ldap"
	"github.com/portainer/portainer-ee/api/http/handler/licenses"
	"github.com/portainer/portainer-ee/api/http/handler/motd"
	"github.com/portainer/portainer-ee/api/http/handler/nomad"
	"github.com/portainer/portainer-ee/api/http/handler/registries"
	"github.com/portainer/portainer-ee/api/http/handler/resourcecontrols"
	"github.com/portainer/portainer-ee/api/http/handler/roles"
	"github.com/portainer/portainer-ee/api/http/handler/settings"
	sslhandler "github.com/portainer/portainer-ee/api/http/handler/ssl"
	"github.com/portainer/portainer-ee/api/http/handler/stacks"
	"github.com/portainer/portainer-ee/api/http/handler/status"
	"github.com/portainer/portainer-ee/api/http/handler/storybook"
	"github.com/portainer/portainer-ee/api/http/handler/tags"
	"github.com/portainer/portainer-ee/api/http/handler/teammemberships"
	"github.com/portainer/portainer-ee/api/http/handler/teams"
	"github.com/portainer/portainer-ee/api/http/handler/templates"
	"github.com/portainer/portainer-ee/api/http/handler/upload"
	"github.com/portainer/portainer-ee/api/http/handler/useractivity"
	"github.com/portainer/portainer-ee/api/http/handler/users"
	"github.com/portainer/portainer-ee/api/http/handler/webhooks"
	"github.com/portainer/portainer-ee/api/http/handler/websocket"
	"github.com/portainer/portainer-ee/api/http/offlinegate"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	edgestackservice "github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	"github.com/portainer/portainer-ee/api/internal/ssl"
	k8s "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/nomad/clientFactory"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Server implements the portaineree.Server interface
type Server struct {
	AuthorizationService        *authorization.Service
	BindAddress                 string
	BindAddressHTTPS            string
	HTTPEnabled                 bool
	AssetsPath                  string
	Status                      *portaineree.Status
	ReverseTunnelService        portaineree.ReverseTunnelService
	ComposeStackManager         portaineree.ComposeStackManager
	CryptoService               portaineree.CryptoService
	LicenseService              portaineree.LicenseService
	SignatureService            portaineree.DigitalSignatureService
	EdgeAsyncService            *edgeasync.Service
	EdgeStacksService           *edgestackservice.Service
	SnapshotService             portaineree.SnapshotService
	FileService                 portainer.FileService
	DataStore                   dataservices.DataStore
	GitService                  portaineree.GitService
	APIKeyService               apikey.APIKeyService
	OpenAMTService              portainer.OpenAMTService
	JWTService                  portaineree.JWTService
	LDAPService                 portaineree.LDAPService
	OAuthService                portaineree.OAuthService
	SwarmStackManager           portaineree.SwarmStackManager
	UserActivityStore           portaineree.UserActivityStore
	UserActivityService         portaineree.UserActivityService
	ProxyManager                *proxy.Manager
	KubernetesTokenCacheManager *kubernetes.TokenCacheManager
	KubeClusterAccessService    k8s.KubeClusterAccessService
	Handler                     *handler.Handler
	SSLService                  *ssl.Service
	DockerClientFactory         *client.ClientFactory
	KubernetesClientFactory     *cli.ClientFactory
	KubernetesDeployer          portaineree.KubernetesDeployer
	NomadClientFactory          *clientFactory.ClientFactory
	HelmPackageManager          libhelm.HelmPackageManager
	Scheduler                   *scheduler.Scheduler
	ShutdownCtx                 context.Context
	ShutdownTrigger             context.CancelFunc
	CloudClusterSetupService    *cloud.CloudClusterSetupService
	CloudClusterInfoService     *cloud.CloudClusterInfoService
	CloudCredentialService      *cloudcredential.Service
	StackDeployer               deployments.StackDeployer
	DemoService                 *demo.Service
}

// Start starts the HTTP server
func (server *Server) Start() error {
	server.AuthorizationService.RegisterEventHandler("kubernetesTokenCacheManager", server.KubernetesTokenCacheManager)

	requestBouncer := security.NewRequestBouncer(server.DataStore, server.LicenseService, server.JWTService, server.APIKeyService, server.SSLService)

	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	offlineGate := offlinegate.NewOfflineGate()

	passwordStrengthChecker := security.NewPasswordStrengthChecker(server.DataStore.Settings())

	var authHandler = auth.NewHandler(requestBouncer, rateLimiter, passwordStrengthChecker)
	authHandler.AuthorizationService = server.AuthorizationService
	authHandler.DataStore = server.DataStore
	authHandler.CryptoService = server.CryptoService
	authHandler.JWTService = server.JWTService
	authHandler.LDAPService = server.LDAPService
	authHandler.LicenseService = server.LicenseService
	authHandler.ProxyManager = server.ProxyManager
	authHandler.KubernetesTokenCacheManager = server.KubernetesTokenCacheManager
	authHandler.OAuthService = server.OAuthService
	authHandler.UserActivityService = server.UserActivityService

	adminMonitor := adminmonitor.New(5*time.Minute, server.DataStore, server.ShutdownCtx)
	adminMonitor.Start()

	backupScheduler := backupOps.NewBackupScheduler(offlineGate, server.DataStore, server.UserActivityStore, server.FileService.GetDatastorePath())
	if err := backupScheduler.Start(); err != nil {
		return errors.Wrap(err, "failed to start backup scheduler")
	}
	var backupHandler = backup.NewHandler(
		requestBouncer,
		server.DataStore,
		server.UserActivityStore,
		offlineGate,
		server.FileService.GetDatastorePath(),
		backupScheduler,
		server.ShutdownTrigger,
		adminMonitor,
		server.DemoService,
	)

	var roleHandler = roles.NewHandler(requestBouncer)
	roleHandler.DataStore = server.DataStore

	var customTemplatesHandler = customtemplates.NewHandler(requestBouncer, server.UserActivityService)
	customTemplatesHandler.DataStore = server.DataStore
	customTemplatesHandler.FileService = server.FileService
	customTemplatesHandler.GitService = server.GitService

	var edgeGroupsHandler = edgegroups.NewHandler(requestBouncer, server.UserActivityService)
	edgeGroupsHandler.DataStore = server.DataStore

	var edgeJobsHandler = edgejobs.NewHandler(requestBouncer, server.UserActivityService, server.EdgeAsyncService)
	edgeJobsHandler.DataStore = server.DataStore
	edgeJobsHandler.FileService = server.FileService
	edgeJobsHandler.ReverseTunnelService = server.ReverseTunnelService

	edgeUpdateService, err := updateschedules.NewService(server.DataStore)
	if err != nil {
		return errors.WithMessage(err, "Unable to create EdgeUpdateService")
	}

	edgeUpdateScheduleHandler := edgeupdateschedules.NewHandler(requestBouncer, server.DataStore, server.FileService, server.AssetsPath, server.EdgeStacksService, edgeUpdateService)

	var edgeStacksHandler = edgestacks.NewHandler(requestBouncer, server.UserActivityService, server.DataStore, server.EdgeAsyncService, server.EdgeStacksService, edgeUpdateService)
	edgeStacksHandler.FileService = server.FileService
	edgeStacksHandler.GitService = server.GitService
	edgeStacksHandler.KubernetesDeployer = server.KubernetesDeployer

	var edgeTemplatesHandler = edgetemplates.NewHandler(requestBouncer)
	edgeTemplatesHandler.DataStore = server.DataStore

	var endpointHandler = endpoints.NewHandler(requestBouncer, server.UserActivityService, server.DataStore, server.EdgeAsyncService, server.DemoService, server.CloudClusterSetupService, server.LicenseService)
	endpointHandler.AuthorizationService = server.AuthorizationService
	endpointHandler.FileService = server.FileService
	endpointHandler.ProxyManager = server.ProxyManager
	endpointHandler.SnapshotService = server.SnapshotService
	endpointHandler.ReverseTunnelService = server.ReverseTunnelService
	endpointHandler.K8sClientFactory = server.KubernetesClientFactory
	endpointHandler.ComposeStackManager = server.ComposeStackManager
	endpointHandler.DockerClientFactory = server.DockerClientFactory
	endpointHandler.BindAddress = server.BindAddress
	endpointHandler.BindAddressHTTPS = server.BindAddressHTTPS
	endpointHandler.KubernetesTokenCacheManager = server.KubernetesTokenCacheManager

	var endpointEdgeHandler = endpointedge.NewHandler(requestBouncer, server.DataStore, server.FileService, server.ReverseTunnelService, server.EdgeAsyncService, server.LicenseService, edgeUpdateService)

	var endpointGroupHandler = endpointgroups.NewHandler(requestBouncer, server.UserActivityService)
	endpointGroupHandler.AuthorizationService = server.AuthorizationService
	endpointGroupHandler.DataStore = server.DataStore

	var endpointProxyHandler = endpointproxy.NewHandler(requestBouncer)
	endpointProxyHandler.DataStore = server.DataStore
	endpointProxyHandler.ProxyManager = server.ProxyManager
	endpointProxyHandler.ReverseTunnelService = server.ReverseTunnelService

	var kaasHandler = kaas.NewHandler(requestBouncer, server.DataStore, server.CloudClusterSetupService, server.CloudClusterInfoService, server.UserActivityService, server.LicenseService)

	var kubernetesHandler = kubehandler.NewHandler(requestBouncer, server.AuthorizationService, server.DataStore, server.JWTService, server.KubeClusterAccessService, server.KubernetesClientFactory, nil, server.UserActivityService, server.KubernetesDeployer, server.FileService, server.AssetsPath)

	containerService := docker.NewContainerService(server.DockerClientFactory, server.DataStore)

	var dockerHandler = dockerhandler.NewHandler(requestBouncer, server.AuthorizationService, server.DataStore, server.DockerClientFactory, containerService)

	var licenseHandler = licenses.NewHandler(requestBouncer, server.UserActivityService, server.DemoService)
	licenseHandler.LicenseService = server.LicenseService

	var fileHandler = file.NewHandler(filepath.Join(server.AssetsPath, "public"), adminMonitor.WasInstanceDisabled)

	var endpointHelmHandler = helm.NewHandler(requestBouncer, server.DataStore, server.JWTService, server.KubernetesDeployer, server.HelmPackageManager, server.KubeClusterAccessService, server.UserActivityService)

	var gitOperationHandler = gitops.NewHandler(requestBouncer, server.DataStore, server.GitService)

	var helmTemplatesHandler = helm.NewTemplateHandler(requestBouncer, server.HelmPackageManager)

	var ldapHandler = ldap.NewHandler(requestBouncer)
	ldapHandler.DataStore = server.DataStore
	ldapHandler.FileService = server.FileService
	ldapHandler.LDAPService = server.LDAPService

	var motdHandler = motd.NewHandler(requestBouncer)

	var registryHandler = registries.NewHandler(requestBouncer, server.UserActivityService)
	registryHandler.DataStore = server.DataStore
	registryHandler.FileService = server.FileService
	registryHandler.ProxyManager = server.ProxyManager
	registryHandler.K8sClientFactory = server.KubernetesClientFactory

	var resourceControlHandler = resourcecontrols.NewHandler(requestBouncer, server.DataStore, server.UserActivityService)

	var settingsHandler = settings.NewHandler(requestBouncer, server.UserActivityService, server.DemoService)
	settingsHandler.AuthorizationService = server.AuthorizationService
	settingsHandler.DataStore = server.DataStore
	settingsHandler.FileService = server.FileService
	settingsHandler.JWTService = server.JWTService
	settingsHandler.LDAPService = server.LDAPService
	settingsHandler.SnapshotService = server.SnapshotService

	var sslHandler = sslhandler.NewHandler(requestBouncer)
	sslHandler.SSLService = server.SSLService

	openAMTHandler := openamt.NewHandler(requestBouncer, server.DataStore)
	openAMTHandler.OpenAMTService = server.OpenAMTService
	openAMTHandler.DataStore = server.DataStore
	openAMTHandler.DockerClientFactory = server.DockerClientFactory

	fdoHandler := fdo.NewHandler(requestBouncer, server.DataStore, server.FileService)

	var stackHandler = stacks.NewHandler(requestBouncer, server.DataStore, server.UserActivityService)
	stackHandler.DockerClientFactory = server.DockerClientFactory
	stackHandler.FileService = server.FileService
	stackHandler.KubernetesDeployer = server.KubernetesDeployer
	stackHandler.GitService = server.GitService
	stackHandler.KubernetesClientFactory = server.KubernetesClientFactory
	stackHandler.AuthorizationService = server.AuthorizationService
	stackHandler.Scheduler = server.Scheduler
	stackHandler.SwarmStackManager = server.SwarmStackManager
	stackHandler.ComposeStackManager = server.ComposeStackManager
	stackHandler.StackDeployer = server.StackDeployer

	var statusHandler = status.NewHandler(requestBouncer, server.Status, server.DemoService, server.DataStore)

	var storybookHandler = storybook.NewHandler(server.AssetsPath)

	var tagHandler = tags.NewHandler(requestBouncer, server.UserActivityService)
	tagHandler.DataStore = server.DataStore

	var teamHandler = teams.NewHandler(requestBouncer, server.UserActivityService)
	teamHandler.AuthorizationService = server.AuthorizationService
	teamHandler.DataStore = server.DataStore
	teamHandler.K8sClientFactory = server.KubernetesClientFactory

	var teamMembershipHandler = teammemberships.NewHandler(requestBouncer, server.UserActivityService)
	teamMembershipHandler.AuthorizationService = server.AuthorizationService
	teamMembershipHandler.DataStore = server.DataStore

	var templatesHandler = templates.NewHandler(requestBouncer)
	templatesHandler.DataStore = server.DataStore
	templatesHandler.FileService = server.FileService
	templatesHandler.GitService = server.GitService

	var uploadHandler = upload.NewHandler(requestBouncer, server.UserActivityService)
	uploadHandler.FileService = server.FileService

	var userHandler = users.NewHandler(requestBouncer, rateLimiter, server.APIKeyService, server.UserActivityService, server.DemoService, passwordStrengthChecker)
	userHandler.AuthorizationService = server.AuthorizationService
	userHandler.DataStore = server.DataStore
	userHandler.CryptoService = server.CryptoService
	userHandler.K8sClientFactory = server.KubernetesClientFactory

	var userActivityHandler = useractivity.NewHandler(requestBouncer)
	userActivityHandler.UserActivityStore = server.UserActivityStore

	var websocketHandler = websocket.NewHandler(server.KubernetesTokenCacheManager, requestBouncer, server.AuthorizationService, server.DataStore, server.UserActivityService)
	websocketHandler.SignatureService = server.SignatureService
	websocketHandler.ReverseTunnelService = server.ReverseTunnelService
	websocketHandler.KubernetesClientFactory = server.KubernetesClientFactory

	var webhookHandler = webhooks.NewHandler(requestBouncer, server.DataStore, server.UserActivityService, containerService)
	webhookHandler.DockerClientFactory = server.DockerClientFactory

	var nomadHandler = nomad.NewHandler(requestBouncer, server.DataStore, server.NomadClientFactory, server.AuthorizationService)

	var cloudCredHandler = cloudcredentials.NewHandler(requestBouncer, server.UserActivityService)
	cloudCredHandler.DataStore = server.DataStore

	server.Handler = &handler.Handler{
		RoleHandler:               roleHandler,
		AuthHandler:               authHandler,
		BackupHandler:             backupHandler,
		CustomTemplatesHandler:    customTemplatesHandler,
		DockerHandler:             dockerHandler,
		EdgeGroupsHandler:         edgeGroupsHandler,
		EdgeJobsHandler:           edgeJobsHandler,
		EdgeUpdateScheduleHandler: edgeUpdateScheduleHandler,
		EdgeStacksHandler:         edgeStacksHandler,
		EdgeTemplatesHandler:      edgeTemplatesHandler,
		EndpointGroupHandler:      endpointGroupHandler,
		EndpointHandler:           endpointHandler,
		EndpointHelmHandler:       endpointHelmHandler,
		EndpointEdgeHandler:       endpointEdgeHandler,
		EndpointProxyHandler:      endpointProxyHandler,
		GitOperationHandler:       gitOperationHandler,
		HelmTemplatesHandler:      helmTemplatesHandler,
		KaasHandler:               kaasHandler,
		KubernetesHandler:         kubernetesHandler,
		FileHandler:               fileHandler,
		LDAPHandler:               ldapHandler,
		LicenseHandler:            licenseHandler,
		MOTDHandler:               motdHandler,
		OpenAMTHandler:            openAMTHandler,
		FDOHandler:                fdoHandler,
		RegistryHandler:           registryHandler,
		ResourceControlHandler:    resourceControlHandler,
		SettingsHandler:           settingsHandler,
		SSLHandler:                sslHandler,
		StatusHandler:             statusHandler,
		StackHandler:              stackHandler,
		StorybookHandler:          storybookHandler,
		TagHandler:                tagHandler,
		TeamHandler:               teamHandler,
		TeamMembershipHandler:     teamMembershipHandler,
		TemplatesHandler:          templatesHandler,
		UploadHandler:             uploadHandler,
		UserHandler:               userHandler,
		UserActivityHandler:       userActivityHandler,
		WebSocketHandler:          websocketHandler,
		WebhookHandler:            webhookHandler,
		NomadHandler:              nomadHandler,
		CloudCredentialsHandler:   cloudCredHandler,
	}

	handler := adminMonitor.WithRedirect(offlineGate.WaitingMiddleware(time.Minute, server.Handler))
	if server.HTTPEnabled {
		go func() {
			log.Info().Str("bind_address", server.BindAddress).Msg("starting HTTP server")
			httpServer := &http.Server{
				Addr:    server.BindAddress,
				Handler: handler,
			}

			go shutdown(server.ShutdownCtx, httpServer, backupScheduler)

			err := httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("HTTP server failed to start")
			}
		}()
	}

	log.Info().Str("bind_address", server.BindAddressHTTPS).Msg("starting HTTPS server")
	httpsServer := &http.Server{
		Addr:    server.BindAddressHTTPS,
		Handler: handler,
	}

	httpsServer.TLSConfig = crypto.CreateServerTLSConfiguration()
	httpsServer.TLSConfig.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return server.SSLService.GetRawCertificate(), nil
	}

	caCertPool := server.SSLService.GetCACertificatePool()
	if caCertPool != nil {
		log.Debug().Str("bind_address_HTTPS", server.BindAddressHTTPS).Msg("using CA certificate")

		httpsServer.TLSConfig.ClientCAs = caCertPool
		httpsServer.TLSConfig.ClientAuth = tls.VerifyClientCertIfGiven // can't use tls.RequireAndVerifyClientCert, this port is also used for the browser
	}

	go shutdown(server.ShutdownCtx, httpsServer, backupScheduler)

	return httpsServer.ListenAndServeTLS("", "")
}

func shutdown(shutdownCtx context.Context, httpServer *http.Server, backupScheduler *backupOps.BackupScheduler) {
	<-shutdownCtx.Done()

	backupScheduler.Stop()

	log.Debug().Msg("shutting down the HTTP server")
	shutdownTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := httpServer.Shutdown(shutdownTimeout)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to shut down the HTTP server")
	}
}

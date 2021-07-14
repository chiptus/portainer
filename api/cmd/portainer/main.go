package main

import (
	"context"
	"log"
	"os"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt"
	"github.com/portainer/portainer/api/chisel"
	"github.com/portainer/portainer/api/cli"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/docker"

	"github.com/portainer/portainer/api/exec"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/git"
	"github.com/portainer/portainer/api/http"
	"github.com/portainer/portainer/api/http/client"
	"github.com/portainer/portainer/api/http/proxy"
	kubeproxy "github.com/portainer/portainer/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/internal/edge"
	"github.com/portainer/portainer/api/internal/snapshot"
	"github.com/portainer/portainer/api/jwt"
	"github.com/portainer/portainer/api/kubernetes"
	kubecli "github.com/portainer/portainer/api/kubernetes/cli"
	"github.com/portainer/portainer/api/ldap"
	"github.com/portainer/portainer/api/libcompose"
	"github.com/portainer/portainer/api/license"
	"github.com/portainer/portainer/api/oauth"
	"github.com/portainer/portainer/api/useractivity"
)

func initCLI() *portainer.CLIFlags {
	var cliService portainer.CLIService = &cli.Service{}
	flags, err := cliService.ParseFlags(portainer.APIVersion)
	if err != nil {
		log.Fatalf("failed parsing flags: %s", err)
	}

	err = cliService.ValidateFlags(flags)
	if err != nil {
		log.Fatalf("failed validating flags:%s", err)
	}
	return flags
}

func initUserActivityStore(dataStorePath string) portainer.UserActivityStore {
	store, err := useractivity.NewUserActivityStore(dataStorePath)
	if err != nil {
		log.Fatalf("Failed initalizing user activity store: %s", err)
	}

	return store
}

func initFileService(dataStorePath string) portainer.FileService {
	fileService, err := filesystem.NewService(dataStorePath, "")
	if err != nil {
		log.Fatalf("failed creating file service: %s", err)
	}
	return fileService
}

func initDataStore(dataStorePath string, rollback bool, fileService portainer.FileService) portainer.DataStore {
	store, err := bolt.NewStore(dataStorePath, fileService)
	if err != nil {
		log.Fatalf("failed creating data store: %s", err)
	}

	err = store.Open()
	if err != nil {
		log.Fatalf("failed opening store: %s", err)
	}

	err = store.Init()
	if err != nil {
		log.Fatalf("failed initializing data store: %s", err)
	}

	if rollback {
		err := store.RollbackToCE()
		if err != nil {
			log.Fatalf("failed rolling back to CE: %s", err)
		}

		log.Println("Exiting rollback")
		os.Exit(0)
		return nil
	}

	err = store.MigrateData(false)
	if err != nil {
		log.Fatalf("failed migration: %s", err)
	}
	return store
}

func initComposeStackManager(assetsPath string, dataStorePath string, reverseTunnelService portainer.ReverseTunnelService, proxyManager *proxy.Manager) portainer.ComposeStackManager {
	composeWrapper, err := exec.NewComposeStackManager(assetsPath, dataStorePath, proxyManager)
	if err != nil {
		log.Printf("[INFO] [main,compose] [message: falling-back to libcompose] [error: %s]", err)
		return libcompose.NewComposeStackManager(dataStorePath, reverseTunnelService)
	}

	return composeWrapper
}

func initSwarmStackManager(assetsPath string, dataStorePath string, signatureService portainer.DigitalSignatureService, fileService portainer.FileService, reverseTunnelService portainer.ReverseTunnelService) (portainer.SwarmStackManager, error) {
	return exec.NewSwarmStackManager(assetsPath, dataStorePath, signatureService, fileService, reverseTunnelService)
}

func initKubernetesDeployer(dataStore portainer.DataStore, reverseTunnelService portainer.ReverseTunnelService, signatureService portainer.DigitalSignatureService, assetsPath string) portainer.KubernetesDeployer {
	return exec.NewKubernetesDeployer(dataStore, reverseTunnelService, signatureService, assetsPath)
}

func initJWTService(dataStore portainer.DataStore) (portainer.JWTService, error) {
	settings, err := dataStore.Settings().Settings()
	if err != nil {
		return nil, err
	}

	userSessionTimeout := settings.UserSessionTimeout
	if userSessionTimeout == "" {
		userSessionTimeout = portainer.DefaultUserSessionTimeout
	}
	jwtService, err := jwt.NewService(userSessionTimeout)
	if err != nil {
		return nil, err
	}
	return jwtService, nil
}

func initDigitalSignatureService() portainer.DigitalSignatureService {
	return crypto.NewECDSAService(os.Getenv("AGENT_SECRET"))
}

func initCryptoService() portainer.CryptoService {
	return &crypto.Service{}
}

func initLDAPService() portainer.LDAPService {
	return &ldap.Service{}
}

func initOAuthService() portainer.OAuthService {
	return oauth.NewService()
}

func initGitService() portainer.GitService {
	return git.NewService()
}

func initDockerClientFactory(signatureService portainer.DigitalSignatureService, reverseTunnelService portainer.ReverseTunnelService) *docker.ClientFactory {
	return docker.NewClientFactory(signatureService, reverseTunnelService)
}

func initKubernetesClientFactory(signatureService portainer.DigitalSignatureService, reverseTunnelService portainer.ReverseTunnelService, dataStore portainer.DataStore, instanceID string) *kubecli.ClientFactory {
	return kubecli.NewClientFactory(signatureService, reverseTunnelService, dataStore, instanceID)
}

func initSnapshotService(snapshotInterval string, dataStore portainer.DataStore, dockerClientFactory *docker.ClientFactory, kubernetesClientFactory *kubecli.ClientFactory, shutdownCtx context.Context) (portainer.SnapshotService, error) {
	dockerSnapshotter := docker.NewSnapshotter(dockerClientFactory)
	kubernetesSnapshotter := kubernetes.NewSnapshotter(kubernetesClientFactory)

	snapshotService, err := snapshot.NewService(snapshotInterval, dataStore, dockerSnapshotter, kubernetesSnapshotter, shutdownCtx)
	if err != nil {
		return nil, err
	}

	return snapshotService, nil
}

func initStatus(flags *portainer.CLIFlags) *portainer.Status {
	return &portainer.Status{
		Version: portainer.APIVersion,
	}
}

func updateSettingsFromFlags(dataStore portainer.DataStore, flags *portainer.CLIFlags) error {
	settings, err := dataStore.Settings().Settings()
	if err != nil {
		return err
	}

	settings.LogoURL = *flags.Logo
	settings.SnapshotInterval = *flags.SnapshotInterval
	settings.EnableEdgeComputeFeatures = *flags.EnableEdgeComputeFeatures
	settings.EnableTelemetry = true
	settings.OAuthSettings.SSO = true
	settings.OAuthSettings.HideInternalAuth = true

	if *flags.Templates != "" {
		settings.TemplatesURL = *flags.Templates
	}

	if *flags.Labels != nil {
		settings.BlackListedLabels = *flags.Labels
	}

	return dataStore.Settings().UpdateSettings(settings)
}

func loadAndParseKeyPair(fileService portainer.FileService, signatureService portainer.DigitalSignatureService) error {
	private, public, err := fileService.LoadKeyPair()
	if err != nil {
		return err
	}
	return signatureService.ParseKeyPair(private, public)
}

func generateAndStoreKeyPair(fileService portainer.FileService, signatureService portainer.DigitalSignatureService) error {
	private, public, err := signatureService.GenerateKeyPair()
	if err != nil {
		return err
	}
	privateHeader, publicHeader := signatureService.PEMHeaders()
	return fileService.StoreKeyPair(private, public, privateHeader, publicHeader)
}

func initKeyPair(fileService portainer.FileService, signatureService portainer.DigitalSignatureService) error {
	existingKeyPair, err := fileService.KeyPairFilesExist()
	if err != nil {
		log.Fatalf("failed checking for existing key pair: %s", err)
	}

	if existingKeyPair {
		return loadAndParseKeyPair(fileService, signatureService)
	}
	return generateAndStoreKeyPair(fileService, signatureService)
}

func createTLSSecuredEndpoint(flags *portainer.CLIFlags, dataStore portainer.DataStore, snapshotService portainer.SnapshotService) error {
	tlsConfiguration := portainer.TLSConfiguration{
		TLS:           *flags.TLS,
		TLSSkipVerify: *flags.TLSSkipVerify,
	}

	if *flags.TLS {
		tlsConfiguration.TLSCACertPath = *flags.TLSCacert
		tlsConfiguration.TLSCertPath = *flags.TLSCert
		tlsConfiguration.TLSKeyPath = *flags.TLSKey
	} else if !*flags.TLS && *flags.TLSSkipVerify {
		tlsConfiguration.TLS = true
	}

	endpointID := dataStore.Endpoint().GetNextIdentifier()
	endpoint := &portainer.Endpoint{
		ID:                 portainer.EndpointID(endpointID),
		Name:               "primary",
		URL:                *flags.EndpointURL,
		GroupID:            portainer.EndpointGroupID(1),
		Type:               portainer.DockerEnvironment,
		TLSConfig:          tlsConfiguration,
		UserAccessPolicies: portainer.UserAccessPolicies{},
		TeamAccessPolicies: portainer.TeamAccessPolicies{},
		Extensions:         []portainer.EndpointExtension{},
		TagIDs:             []portainer.TagID{},
		Status:             portainer.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portainer.KubernetesDefault(),

		SecuritySettings: portainer.EndpointSecuritySettings{
			AllowVolumeBrowserForRegularUsers: false,
			EnableHostManagementFeatures:      false,

			AllowSysctlSettingForRegularUsers:         true,
			AllowBindMountsForRegularUsers:            true,
			AllowPrivilegedModeForRegularUsers:        true,
			AllowHostNamespaceForRegularUsers:         true,
			AllowContainerCapabilitiesForRegularUsers: true,
			AllowDeviceMappingForRegularUsers:         true,
			AllowStackManagementForRegularUsers:       true,
		},
	}

	if strings.HasPrefix(endpoint.URL, "tcp://") {
		tlsConfig, err := crypto.CreateTLSConfigurationFromDisk(tlsConfiguration.TLSCACertPath, tlsConfiguration.TLSCertPath, tlsConfiguration.TLSKeyPath, tlsConfiguration.TLSSkipVerify)
		if err != nil {
			return err
		}

		agentOnDockerEnvironment, err := client.ExecutePingOperation(endpoint.URL, tlsConfig)
		if err != nil {
			return err
		}

		if agentOnDockerEnvironment {
			endpoint.Type = portainer.AgentOnDockerEnvironment
		}
	}

	err := snapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		log.Printf("http error: endpoint snapshot error (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
	}

	return dataStore.Endpoint().CreateEndpoint(endpoint)
}

func createUnsecuredEndpoint(endpointURL string, dataStore portainer.DataStore, snapshotService portainer.SnapshotService) error {
	if strings.HasPrefix(endpointURL, "tcp://") {
		_, err := client.ExecutePingOperation(endpointURL, nil)
		if err != nil {
			return err
		}
	}

	endpointID := dataStore.Endpoint().GetNextIdentifier()
	endpoint := &portainer.Endpoint{
		ID:                 portainer.EndpointID(endpointID),
		Name:               "primary",
		URL:                endpointURL,
		GroupID:            portainer.EndpointGroupID(1),
		Type:               portainer.DockerEnvironment,
		TLSConfig:          portainer.TLSConfiguration{},
		UserAccessPolicies: portainer.UserAccessPolicies{},
		TeamAccessPolicies: portainer.TeamAccessPolicies{},
		Extensions:         []portainer.EndpointExtension{},
		TagIDs:             []portainer.TagID{},
		Status:             portainer.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portainer.KubernetesDefault(),

		SecuritySettings: portainer.EndpointSecuritySettings{
			AllowVolumeBrowserForRegularUsers: false,
			EnableHostManagementFeatures:      false,

			AllowSysctlSettingForRegularUsers:         true,
			AllowBindMountsForRegularUsers:            true,
			AllowPrivilegedModeForRegularUsers:        true,
			AllowHostNamespaceForRegularUsers:         true,
			AllowContainerCapabilitiesForRegularUsers: true,
			AllowDeviceMappingForRegularUsers:         true,
			AllowStackManagementForRegularUsers:       true,
		},
	}

	err := snapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		log.Printf("http error: endpoint snapshot error (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
	}

	return dataStore.Endpoint().CreateEndpoint(endpoint)
}

func initEndpoint(flags *portainer.CLIFlags, dataStore portainer.DataStore, snapshotService portainer.SnapshotService) error {
	if *flags.EndpointURL == "" {
		return nil
	}

	endpoints, err := dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	if len(endpoints) > 0 {
		log.Println("Instance already has defined endpoints. Skipping the endpoint defined via CLI.")
		return nil
	}

	if *flags.TLS || *flags.TLSSkipVerify {
		return createTLSSecuredEndpoint(flags, dataStore, snapshotService)
	}
	return createUnsecuredEndpoint(*flags.EndpointURL, dataStore, snapshotService)
}

func buildServer(flags *portainer.CLIFlags) portainer.Server {
	shutdownCtx, shutdownTrigger := context.WithCancel(context.Background())

	fileService := initFileService(*flags.Data)

	dataStore := initDataStore(*flags.Data, *flags.RollbackToCE, fileService)

	jwtService, err := initJWTService(dataStore)
	if err != nil {
		log.Fatalf("failed initializing JWT service: %s", err)
	}

	licenseService := license.NewService(dataStore.License(), shutdownCtx)
	if err = licenseService.Init(); err != nil {
		log.Fatalf("failed initializing license service: %s", err)
	}

	ldapService := initLDAPService()

	oauthService := initOAuthService()

	gitService := initGitService()

	cryptoService := initCryptoService()

	digitalSignatureService := initDigitalSignatureService()

	err = initKeyPair(fileService, digitalSignatureService)
	if err != nil {
		log.Fatalf("failed initializing key pair: %s", err)
	}

	reverseTunnelService := chisel.NewService(dataStore, shutdownCtx)

	instanceID, err := dataStore.Version().InstanceID()
	if err != nil {
		log.Fatalf("failed to get datastore version: %s", err)
	}

	dockerClientFactory := initDockerClientFactory(digitalSignatureService, reverseTunnelService)
	kubernetesClientFactory := initKubernetesClientFactory(digitalSignatureService, reverseTunnelService, dataStore, instanceID)

	snapshotService, err := initSnapshotService(*flags.SnapshotInterval, dataStore, dockerClientFactory, kubernetesClientFactory, shutdownCtx)
	if err != nil {
		log.Fatalf("failed initializing snapshot service: %s", err)
	}
	snapshotService.Start()

	authorizationService := authorization.NewService(dataStore)
	authorizationService.K8sClientFactory = kubernetesClientFactory

	swarmStackManager, err := initSwarmStackManager(*flags.Assets, *flags.Data, digitalSignatureService, fileService, reverseTunnelService)
	if err != nil {
		log.Fatalf("failed initializing swarm stack manager: %s", err)
	}
	kubernetesTokenCacheManager := kubeproxy.NewTokenCacheManager()

	userActivityStore := initUserActivityStore(*flags.Data)

	proxyManager := proxy.NewManager(dataStore, digitalSignatureService, reverseTunnelService, dockerClientFactory, kubernetesClientFactory, kubernetesTokenCacheManager, authorizationService, userActivityStore)

	composeStackManager := initComposeStackManager(*flags.Assets, *flags.Data, reverseTunnelService, proxyManager)

	kubernetesDeployer := initKubernetesDeployer(dataStore, reverseTunnelService, digitalSignatureService, *flags.Assets)

	if dataStore.IsNew() {
		err = updateSettingsFromFlags(dataStore, flags)
		if err != nil {
			log.Fatalf("failed updating settings from flags: %s", err)
		}
	}

	err = edge.LoadEdgeJobs(dataStore, reverseTunnelService)
	if err != nil {
		log.Fatalf("failed loading edge jobs from database: %s", err)
	}

	applicationStatus := initStatus(flags)

	err = initEndpoint(flags, dataStore, snapshotService)
	if err != nil {
		log.Fatalf("failed initializing endpoint: %s", err)
	}

	adminPasswordHash := ""
	if *flags.AdminPasswordFile != "" {
		content, err := fileService.GetFileContent(*flags.AdminPasswordFile)
		if err != nil {
			log.Fatalf("failed getting admin password file: %s", err)
		}
		adminPasswordHash, err = cryptoService.Hash(strings.TrimSuffix(string(content), "\n"))
		if err != nil {
			log.Fatalf("failed hashing admin password: %s", err)
		}
	} else if *flags.AdminPassword != "" {
		adminPasswordHash = *flags.AdminPassword
	}

	if adminPasswordHash != "" {
		users, err := dataStore.User().UsersByRole(portainer.AdministratorRole)
		if err != nil {
			log.Fatalf("failed getting admin user: %s", err)
		}

		if len(users) == 0 {
			log.Println("Created admin user with the given password.")
			user := &portainer.User{
				Username:                "admin",
				Role:                    portainer.AdministratorRole,
				Password:                adminPasswordHash,
				PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
			}
			err := dataStore.User().CreateUser(user)
			if err != nil {
				log.Fatalf("failed creating admin user: %s", err)
			}
		} else {
			log.Println("Instance already has an administrator user defined. Skipping admin password related flags.")
		}
	}

	err = reverseTunnelService.StartTunnelServer(*flags.TunnelAddr, *flags.TunnelPort, snapshotService)
	if err != nil {
		log.Fatalf("failed starting tunnel server: %s", err)
	}

	err = licenseService.Start()
	if err != nil {
		log.Fatalf("failed starting license service: %s", err)
	}

	return &http.Server{
		AuthorizationService:        authorizationService,
		ReverseTunnelService:        reverseTunnelService,
		Status:                      applicationStatus,
		BindAddress:                 *flags.Addr,
		AssetsPath:                  *flags.Assets,
		DataStore:                   dataStore,
		LicenseService:              licenseService,
		SwarmStackManager:           swarmStackManager,
		ComposeStackManager:         composeStackManager,
		KubernetesDeployer:          kubernetesDeployer,
		CryptoService:               cryptoService,
		JWTService:                  jwtService,
		FileService:                 fileService,
		LDAPService:                 ldapService,
		OAuthService:                oauthService,
		GitService:                  gitService,
		ProxyManager:                proxyManager,
		KubernetesTokenCacheManager: kubernetesTokenCacheManager,
		SignatureService:            digitalSignatureService,
		SnapshotService:             snapshotService,
		SSL:                         *flags.SSL,
		SSLCert:                     *flags.SSLCert,
		SSLKey:                      *flags.SSLKey,
		DockerClientFactory:         dockerClientFactory,
		UserActivityStore:           userActivityStore,
		KubernetesClientFactory:     kubernetesClientFactory,
		ShutdownCtx:                 shutdownCtx,
		ShutdownTrigger:             shutdownTrigger,
	}
}

func main() {
	flags := initCLI()

	for {
		server := buildServer(flags)
		log.Printf("Starting Portainer %s on %s\n", portainer.APIVersion, *flags.Addr)
		err := server.Start()
		log.Printf("Http server exited: %s\n", err)
	}
}

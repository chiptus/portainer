package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/portainer/libhelm"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/bolt"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/cli"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/exec"
	"github.com/portainer/portainer-ee/api/http"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/http/proxy"
	kubeproxy "github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	"github.com/portainer/portainer-ee/api/internal/ssl"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer-ee/api/kubernetes"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/ldap"
	"github.com/portainer/portainer-ee/api/license"
	"github.com/portainer/portainer-ee/api/oauth"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks"
	"github.com/portainer/portainer-ee/api/useractivity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/git"
	"github.com/portainer/portainer/api/hostmanagement/openamt"
)

func initCLI() *portaineree.CLIFlags {
	var cliService portaineree.CLIService = &cli.Service{}
	flags, err := cliService.ParseFlags(portaineree.APIVersion)
	if err != nil {
		log.Fatalf("failed parsing flags: %s", err)
	}

	err = cliService.ValidateFlags(flags)
	if err != nil {
		log.Fatalf("failed validating flags:%s", err)
	}
	return flags
}

func initUserActivity(dataStorePath string, shutdownCtx context.Context) (portaineree.UserActivityService, portaineree.UserActivityStore) {
	store, err := useractivity.NewStore(dataStorePath)
	if err != nil {
		log.Fatalf("Failed initalizing user activity store: %s", err)
	}

	service := useractivity.NewService(store)

	go shutdownUserActivityStore(shutdownCtx, store)

	return service, store
}

func shutdownUserActivityStore(shutdownCtx context.Context, store portaineree.UserActivityStore) {
	<-shutdownCtx.Done()
	store.Close()
}

func initFileService(dataStorePath string) portainer.FileService {
	fileService, err := filesystem.NewService(dataStorePath, "")
	if err != nil {
		log.Fatalf("failed creating file service: %s", err)
	}
	return fileService
}

func initDataStore(dataStorePath string, rollback bool, rollbackToCE bool, fileService portainer.FileService, shutdownCtx context.Context) portaineree.DataStore {
	store := bolt.NewStore(dataStorePath, fileService)
	err := store.Open()
	if err != nil {
		log.Fatalf("failed opening store: %s", err)
	}

	if rollback {
		err := store.Rollback(false)
		if err != nil {
			log.Fatalf("failed rolling back: %s", err)
		}

		log.Println("Exiting rollback")
		os.Exit(0)
		return nil
	}

	err = store.Init()
	if err != nil {
		log.Fatalf("failed initializing data store: %s", err)
	}

	if rollbackToCE {
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

	go shutdownDatastore(shutdownCtx, store)
	return store
}

func shutdownDatastore(shutdownCtx context.Context, datastore portaineree.DataStore) {
	<-shutdownCtx.Done()
	datastore.Close()
}

func initComposeStackManager(assetsPath string, configPath string, reverseTunnelService portaineree.ReverseTunnelService, proxyManager *proxy.Manager) portaineree.ComposeStackManager {
	composeWrapper, err := exec.NewComposeStackManager(assetsPath, configPath, proxyManager)
	if err != nil {
		log.Fatalf("failed creating compose manager: %s", err)
	}

	return composeWrapper
}

func initSwarmStackManager(
	assetsPath string,
	configPath string,
	signatureService portaineree.DigitalSignatureService,
	fileService portainer.FileService,
	reverseTunnelService portaineree.ReverseTunnelService,
	dataStore portaineree.DataStore,
) (portaineree.SwarmStackManager, error) {
	return exec.NewSwarmStackManager(assetsPath, configPath, signatureService, fileService, reverseTunnelService, dataStore)
}

func initKubernetesDeployer(authService *authorization.Service, kubernetesTokenCacheManager *kubeproxy.TokenCacheManager, kubernetesClientFactory *kubecli.ClientFactory, dataStore portaineree.DataStore, reverseTunnelService portaineree.ReverseTunnelService, signatureService portaineree.DigitalSignatureService, proxyManager *proxy.Manager, assetsPath string) portaineree.KubernetesDeployer {
	return exec.NewKubernetesDeployer(authService, kubernetesTokenCacheManager, kubernetesClientFactory, dataStore, reverseTunnelService, signatureService, proxyManager, assetsPath)
}

func initHelmPackageManager(assetsPath string) (libhelm.HelmPackageManager, error) {
	return libhelm.NewHelmPackageManager(libhelm.HelmConfig{BinaryPath: assetsPath})
}

func initAPIKeyService(datastore portaineree.DataStore) apikey.APIKeyService {
	return apikey.NewAPIKeyService(datastore.APIKeyRepository(), datastore.User())
}

func initJWTService(dataStore portaineree.DataStore) (portaineree.JWTService, error) {
	settings, err := dataStore.Settings().Settings()
	if err != nil {
		return nil, err
	}

	userSessionTimeout := settings.UserSessionTimeout
	if userSessionTimeout == "" {
		userSessionTimeout = portaineree.DefaultUserSessionTimeout
	}
	jwtService, err := jwt.NewService(userSessionTimeout, dataStore)
	if err != nil {
		return nil, err
	}

	return jwtService, nil
}

func initDigitalSignatureService() portaineree.DigitalSignatureService {
	return crypto.NewECDSAService(os.Getenv("AGENT_SECRET"))
}

func initCryptoService() portaineree.CryptoService {
	return &crypto.Service{}
}

func initLDAPService() portaineree.LDAPService {
	return &ldap.Service{}
}

func initOAuthService() portaineree.OAuthService {
	return oauth.NewService()
}

func initGitService() portaineree.GitService {
	return git.NewService()
}

func initSSLService(addr, dataPath, certPath, keyPath string, fileService portainer.FileService, dataStore portaineree.DataStore, shutdownTrigger context.CancelFunc) (*ssl.Service, error) {
	slices := strings.Split(addr, ":")
	host := slices[0]
	if host == "" {
		host = "0.0.0.0"
	}

	sslService := ssl.NewService(fileService, dataStore, shutdownTrigger)

	err := sslService.Init(host, certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return sslService, nil
}

func initDockerClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService) *docker.ClientFactory {
	return docker.NewClientFactory(signatureService, reverseTunnelService)
}

func initKubernetesClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore portaineree.DataStore, instanceID string) *kubecli.ClientFactory {
	return kubecli.NewClientFactory(signatureService, reverseTunnelService, dataStore, instanceID)
}

func initSnapshotService(snapshotIntervalFromFlag string, dataStore portaineree.DataStore, dockerClientFactory *docker.ClientFactory, kubernetesClientFactory *kubecli.ClientFactory, shutdownCtx context.Context) (portaineree.SnapshotService, error) {
	dockerSnapshotter := docker.NewSnapshotter(dockerClientFactory)
	kubernetesSnapshotter := kubernetes.NewSnapshotter(kubernetesClientFactory)

	snapshotService, err := snapshot.NewService(snapshotIntervalFromFlag, dataStore, dockerSnapshotter, kubernetesSnapshotter, shutdownCtx)
	if err != nil {
		return nil, err
	}

	return snapshotService, nil
}

func initStatus(instanceID string) *portaineree.Status {
	return &portaineree.Status{
		Version:    portaineree.APIVersion,
		InstanceID: instanceID,
	}
}

func updateSettingsFromFlags(dataStore portaineree.DataStore, flags *portaineree.CLIFlags) error {
	settings, err := dataStore.Settings().Settings()
	if err != nil {
		return err
	}

	if *flags.SnapshotInterval != "" {
		settings.SnapshotInterval = *flags.SnapshotInterval
	}

	if *flags.Logo != "" {
		settings.LogoURL = *flags.Logo
	}

	if *flags.EnableEdgeComputeFeatures {
		settings.EnableEdgeComputeFeatures = *flags.EnableEdgeComputeFeatures
	}

	if *flags.Templates != "" {
		settings.TemplatesURL = *flags.Templates
	}

	if *flags.Labels != nil {
		settings.BlackListedLabels = *flags.Labels
	}

	err = dataStore.Settings().UpdateSettings(settings)
	if err != nil {
		return err
	}

	sslSettings, err := dataStore.SSLSettings().Settings()
	if err != nil {
		return err
	}

	if *flags.HTTPDisabled {
		sslSettings.HTTPEnabled = false
	} else {
		sslSettings.HTTPEnabled = *flags.HTTPEnabled || sslSettings.HTTPEnabled
	}

	err = dataStore.SSLSettings().UpdateSettings(sslSettings)
	if err != nil {
		return err
	}

	return nil
}

// enableFeaturesFromFlags turns on or off feature flags
// e.g.  portainer --feat open-amt --feat fdo=true ... (defaults to true)
// note, settings are persisted to the DB. To turn off `--feat open-amt=false`
func enableFeaturesFromFlags(dataStore portaineree.DataStore, flags *portaineree.CLIFlags) error {
	settings, err := dataStore.Settings().Settings()
	if err != nil {
		return err
	}

	if settings.FeatureFlagSettings == nil {
		settings.FeatureFlagSettings = make(map[portaineree.Feature]bool)
	}

	// loop through feature flags to check if they are supported
	for _, feat := range *flags.FeatureFlags {
		var correspondingFeature *portaineree.Feature
		for i, supportedFeat := range portaineree.SupportedFeatureFlags {
			if strings.EqualFold(feat.Name, string(supportedFeat)) {
				correspondingFeature = &portaineree.SupportedFeatureFlags[i]
			}
		}

		if correspondingFeature == nil {
			return fmt.Errorf("unknown feature flag '%s'", feat.Name)
		}

		featureState, err := strconv.ParseBool(feat.Value)
		if err != nil {
			return fmt.Errorf("feature flag's '%s' value should be true or false", feat.Name)
		}

		if featureState {
			log.Printf("Feature %v : on", *correspondingFeature)
		} else {
			log.Printf("Feature %v : off", *correspondingFeature)
		}

		settings.FeatureFlagSettings[*correspondingFeature] = featureState
	}

	return dataStore.Settings().UpdateSettings(settings)
}

func loadAndParseKeyPair(fileService portainer.FileService, signatureService portaineree.DigitalSignatureService) error {
	private, public, err := fileService.LoadKeyPair()
	if err != nil {
		return err
	}
	return signatureService.ParseKeyPair(private, public)
}

func generateAndStoreKeyPair(fileService portainer.FileService, signatureService portaineree.DigitalSignatureService) error {
	private, public, err := signatureService.GenerateKeyPair()
	if err != nil {
		return err
	}
	privateHeader, publicHeader := signatureService.PEMHeaders()
	return fileService.StoreKeyPair(private, public, privateHeader, publicHeader)
}

func initKeyPair(fileService portainer.FileService, signatureService portaineree.DigitalSignatureService) error {
	existingKeyPair, err := fileService.KeyPairFilesExist()
	if err != nil {
		log.Fatalf("failed checking for existing key pair: %s", err)
	}

	if existingKeyPair {
		return loadAndParseKeyPair(fileService, signatureService)
	}
	return generateAndStoreKeyPair(fileService, signatureService)
}

func createTLSSecuredEndpoint(flags *portaineree.CLIFlags, dataStore portaineree.DataStore, snapshotService portaineree.SnapshotService) error {
	tlsConfiguration := portaineree.TLSConfiguration{
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
	endpoint := &portaineree.Endpoint{
		ID:                 portaineree.EndpointID(endpointID),
		Name:               "primary",
		URL:                *flags.EndpointURL,
		GroupID:            portaineree.EndpointGroupID(1),
		Type:               portaineree.DockerEnvironment,
		TLSConfig:          tlsConfiguration,
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		Extensions:         []portaineree.EndpointExtension{},
		TagIDs:             []portaineree.TagID{},
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portaineree.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),

		SecuritySettings: portaineree.EndpointSecuritySettings{
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

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
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
			endpoint.Type = portaineree.AgentOnDockerEnvironment
		}
	}

	err := snapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		log.Printf("http error: environment snapshot error (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
	}

	return dataStore.Endpoint().CreateEndpoint(endpoint)
}

func createUnsecuredEndpoint(endpointURL string, dataStore portaineree.DataStore, snapshotService portaineree.SnapshotService) error {
	if strings.HasPrefix(endpointURL, "tcp://") {
		_, err := client.ExecutePingOperation(endpointURL, nil)
		if err != nil {
			return err
		}
	}

	endpointID := dataStore.Endpoint().GetNextIdentifier()
	endpoint := &portaineree.Endpoint{
		ID:                 portaineree.EndpointID(endpointID),
		Name:               "primary",
		URL:                endpointURL,
		GroupID:            portaineree.EndpointGroupID(1),
		Type:               portaineree.DockerEnvironment,
		TLSConfig:          portaineree.TLSConfiguration{},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		Extensions:         []portaineree.EndpointExtension{},
		TagIDs:             []portaineree.TagID{},
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portaineree.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),

		SecuritySettings: portaineree.EndpointSecuritySettings{
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

		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	}

	err := snapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		log.Printf("http error: environment snapshot error (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
	}

	return dataStore.Endpoint().CreateEndpoint(endpoint)
}

func initEndpoint(flags *portaineree.CLIFlags, dataStore portaineree.DataStore, snapshotService portaineree.SnapshotService) error {
	if *flags.EndpointURL == "" {
		return nil
	}

	endpoints, err := dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	if len(endpoints) > 0 {
		log.Println("Instance already has defined environments. Skipping the environment defined via CLI.")
		return nil
	}

	if *flags.TLS || *flags.TLSSkipVerify {
		return createTLSSecuredEndpoint(flags, dataStore, snapshotService)
	}
	return createUnsecuredEndpoint(*flags.EndpointURL, dataStore, snapshotService)
}

func buildServer(flags *portaineree.CLIFlags) portaineree.Server {
	shutdownCtx, shutdownTrigger := context.WithCancel(context.Background())

	fileService := initFileService(*flags.Data)

	dataStore := initDataStore(*flags.Data, *flags.Rollback, *flags.RollbackToCE, fileService, shutdownCtx)

	apiKeyService := initAPIKeyService(dataStore)

	jwtService, err := initJWTService(dataStore)
	if err != nil {
		log.Fatalf("failed initializing JWT service: %s", err)
	}

	licenseService := license.NewService(dataStore.License(), shutdownCtx)
	if err = licenseService.Init(); err != nil {
		log.Fatalf("failed initializing license service: %s", err)
	}

	err = enableFeaturesFromFlags(dataStore, flags)
	if err != nil {
		log.Fatalf("failed enabling feature flag: %v", err)
	}

	ldapService := initLDAPService()

	oauthService := initOAuthService()

	gitService := initGitService()

	openAMTService := openamt.NewService()

	cryptoService := initCryptoService()

	digitalSignatureService := initDigitalSignatureService()

	sslService, err := initSSLService(*flags.AddrHTTPS, *flags.Data, *flags.SSLCert, *flags.SSLKey, fileService, dataStore, shutdownTrigger)
	if err != nil {
		log.Fatal(err)
	}

	sslSettings, err := sslService.GetSSLSettings()
	if err != nil {
		log.Fatalf("failed to get ssl settings: %s", err)
	}

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

	kubernetesTokenCacheManager := kubeproxy.NewTokenCacheManager()

	kubeConfigService := kubernetes.NewKubeConfigCAService(*flags.AddrHTTPS, sslSettings.CertPath)

	userActivityService, userActivityStore := initUserActivity(*flags.Data, shutdownCtx)

	proxyManager := proxy.NewManager(dataStore, digitalSignatureService, reverseTunnelService, dockerClientFactory, kubernetesClientFactory, kubernetesTokenCacheManager, authorizationService, userActivityService)

	reverseTunnelService.ProxyManager = proxyManager

	dockerConfigPath := fileService.GetDockerConfigPath()

	composeStackManager := initComposeStackManager(*flags.Assets, dockerConfigPath, reverseTunnelService, proxyManager)

	swarmStackManager, err := initSwarmStackManager(*flags.Assets, dockerConfigPath, digitalSignatureService, fileService, reverseTunnelService, dataStore)
	if err != nil {
		log.Fatalf("failed initializing swarm stack manager: %s", err)
	}

	kubernetesDeployer := initKubernetesDeployer(authorizationService, kubernetesTokenCacheManager, kubernetesClientFactory, dataStore, reverseTunnelService, digitalSignatureService, proxyManager, *flags.Assets)

	helmPackageManager, err := initHelmPackageManager(*flags.Assets)
	if err != nil {
		log.Fatalf("failed initializing helm package manager: %s", err)
	}

	err = updateSettingsFromFlags(dataStore, flags)
	if err != nil {
		log.Fatalf("failed updating settings from flags: %s", err)
	}

	err = edge.LoadEdgeJobs(dataStore, reverseTunnelService)
	if err != nil {
		log.Fatalf("failed loading edge jobs from database: %s", err)
	}

	applicationStatus := initStatus(instanceID)

	err = initEndpoint(flags, dataStore, snapshotService)
	if err != nil {
		log.Fatalf("failed initializing environment: %s", err)
	}

	adminPasswordHash := ""
	if *flags.AdminPasswordFile != "" {
		content, err := fileService.GetFileContent(*flags.AdminPasswordFile, "")
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
		users, err := dataStore.User().UsersByRole(portaineree.AdministratorRole)
		if err != nil {
			log.Fatalf("failed getting admin user: %s", err)
		}

		if len(users) == 0 {
			log.Println("Created admin user with the given password.")
			user := &portaineree.User{
				Username:                "admin",
				Role:                    portaineree.AdministratorRole,
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

	scheduler := scheduler.NewScheduler(shutdownCtx)
	stackDeployer := stacks.NewStackDeployer(swarmStackManager, composeStackManager, kubernetesDeployer)
	stacks.StartStackSchedules(scheduler, stackDeployer, dataStore, gitService, userActivityService)

	sslDBSettings, err := dataStore.SSLSettings().Settings()
	if err != nil {
		log.Fatalf("failed to fetch ssl settings from DB")
	}

	return &http.Server{
		AuthorizationService:        authorizationService,
		ReverseTunnelService:        reverseTunnelService,
		Status:                      applicationStatus,
		BindAddress:                 *flags.Addr,
		BindAddressHTTPS:            *flags.AddrHTTPS,
		HTTPEnabled:                 sslDBSettings.HTTPEnabled,
		AssetsPath:                  *flags.Assets,
		DataStore:                   dataStore,
		LicenseService:              licenseService,
		SwarmStackManager:           swarmStackManager,
		ComposeStackManager:         composeStackManager,
		KubernetesDeployer:          kubernetesDeployer,
		HelmPackageManager:          helmPackageManager,
		APIKeyService:               apiKeyService,
		CryptoService:               cryptoService,
		JWTService:                  jwtService,
		FileService:                 fileService,
		LDAPService:                 ldapService,
		OAuthService:                oauthService,
		GitService:                  gitService,
		OpenAMTService:              openAMTService,
		ProxyManager:                proxyManager,
		KubernetesTokenCacheManager: kubernetesTokenCacheManager,
		KubeConfigService:           kubeConfigService,
		SignatureService:            digitalSignatureService,
		SnapshotService:             snapshotService,
		SSLService:                  sslService,
		DockerClientFactory:         dockerClientFactory,
		UserActivityService:         userActivityService,
		UserActivityStore:           userActivityStore,
		KubernetesClientFactory:     kubernetesClientFactory,
		Scheduler:                   scheduler,
		ShutdownCtx:                 shutdownCtx,
		ShutdownTrigger:             shutdownTrigger,
		StackDeployer:               stackDeployer,
		BaseURL:                     *flags.BaseURL,
	}
}

func main() {
	flags := initCLI()

	configureLogger()

	for {
		server := buildServer(flags)
		log.Printf("[INFO] [cmd,main] Starting Portainer version %s\n", portaineree.APIVersion)
		err := server.Start()
		log.Printf("[INFO] [cmd,main] Http server exited: %s\n", err)
	}
}

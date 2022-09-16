package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/portainer/libhelm"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/build"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/cli"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/database"
	"github.com/portainer/portainer-ee/api/database/boltdb"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/exec"
	"github.com/portainer/portainer-ee/api/filesystem"
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
	"github.com/portainer/portainer-ee/api/nomad/clientFactory"
	nomadSnapshot "github.com/portainer/portainer-ee/api/nomad/snapshot"
	"github.com/portainer/portainer-ee/api/oauth"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks"
	"github.com/portainer/portainer-ee/api/useractivity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/git"
	"github.com/portainer/portainer/api/hostmanagement/openamt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func initCLI() *portaineree.CLIFlags {
	var cliService portaineree.CLIService = &cli.Service{}
	flags, err := cliService.ParseFlags(portaineree.APIVersion)
	if err != nil {
		log.Fatal().Err(err).Msg("failed parsing flags")
	}

	err = cliService.ValidateFlags(flags)
	if err != nil {
		log.Fatal().Err(err).Msg("failed validating flags")
	}

	return flags
}

func initUserActivity(dataStorePath string, maxBatchSize int, maxBatchDelay time.Duration, initialMmapSize int, shutdownCtx context.Context) (portaineree.UserActivityService, portaineree.UserActivityStore) {
	store, err := useractivity.NewStore(dataStorePath, maxBatchSize, maxBatchDelay, initialMmapSize)
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating file service")
	}

	service := useractivity.NewService(store)

	go shutdownUserActivityStore(shutdownCtx, store)

	return service, store
}

func shutdownUserActivityStore(shutdownCtx context.Context, store portaineree.UserActivityStore) {
	<-shutdownCtx.Done()
	store.Close()
}

func initFileService(dataStorePath string) portaineree.FileService {
	fileService, err := filesystem.NewService(dataStorePath, "")
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating file service")
	}
	return fileService
}

func initDataStore(flags *portaineree.CLIFlags, secretKey []byte, fileService portainer.FileService, shutdownCtx context.Context) dataservices.DataStore {
	connection, err := database.NewDatabase("boltdb", *flags.Data, secretKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating database connection")
	}

	if bconn, ok := connection.(*boltdb.DbConnection); ok {
		bconn.MaxBatchSize = *flags.MaxBatchSize
		bconn.MaxBatchDelay = *flags.MaxBatchDelay
		bconn.InitialMmapSize = *flags.InitialMmapSize
	} else {
		log.Fatal().Msg("failed creating database connection: expecting a boltdb database type but a different one was received")
	}

	store := datastore.NewStore(*flags.Data, fileService, connection)
	isNew, err := store.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("failed opening store")
	}

	if *flags.Rollback {
		err := store.Rollback(false)
		if err != nil {
			log.Fatal().Err(err).Msg("failed rolling back")
		}

		log.Info().Msg("exiting rollback")
		os.Exit(0)
		return nil
	}

	if *flags.RollbackToCE {
		err := store.RollbackToCE()
		if err != nil {
			log.Fatal().Err(err).Msg("failed rolling back to CE")
		}

		log.Info().Msg("Exiting rollback")
		os.Exit(0)
		return nil
	}

	// Init sets some defaults - it's basically a migration
	err = store.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing data store")
	}

	if isNew {
		// from MigrateData
		store.VersionService.StoreDBVersion(portaineree.DBVersion)
		store.VersionService.StoreEdition(portaineree.PortainerEE)

		err := updateSettingsFromFlags(store, flags)
		if err != nil {
			log.Fatal().Err(err).Msg("failed updating settings from flags")
		}
	} else {
		err = store.MigrateData()
		if err != nil {
			log.Fatal().Err(err).Msg("failure during creation of new database")
		}
	}

	// this is for the db restore functionality - needs more tests.
	go func() {
		<-shutdownCtx.Done()
		defer connection.Close()
	}()

	return store
}

func initComposeStackManager(assetsPath string, configPath string, reverseTunnelService portaineree.ReverseTunnelService, proxyManager *proxy.Manager) portaineree.ComposeStackManager {
	composeWrapper, err := exec.NewComposeStackManager(assetsPath, configPath, proxyManager)
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating compose manager")
	}

	return composeWrapper
}

func initSwarmStackManager(
	assetsPath string,
	configPath string,
	signatureService portaineree.DigitalSignatureService,
	fileService portainer.FileService,
	reverseTunnelService portaineree.ReverseTunnelService,
	dataStore dataservices.DataStore,
) (portaineree.SwarmStackManager, error) {
	return exec.NewSwarmStackManager(assetsPath, configPath, signatureService, fileService, reverseTunnelService, dataStore)
}

func initKubernetesDeployer(authService *authorization.Service, kubernetesTokenCacheManager *kubeproxy.TokenCacheManager, kubernetesClientFactory *kubecli.ClientFactory, dataStore dataservices.DataStore, reverseTunnelService portaineree.ReverseTunnelService, signatureService portaineree.DigitalSignatureService, proxyManager *proxy.Manager, assetsPath string) portaineree.KubernetesDeployer {
	return exec.NewKubernetesDeployer(authService, kubernetesTokenCacheManager, kubernetesClientFactory, dataStore, reverseTunnelService, signatureService, proxyManager, assetsPath)
}

func initHelmPackageManager(assetsPath string) (libhelm.HelmPackageManager, error) {
	return libhelm.NewHelmPackageManager(libhelm.HelmConfig{BinaryPath: assetsPath})
}

func initAPIKeyService(datastore dataservices.DataStore) apikey.APIKeyService {
	return apikey.NewAPIKeyService(datastore.APIKeyRepository(), datastore.User())
}

func initJWTService(dataStore dataservices.DataStore) (dataservices.JWTService, error) {
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

func initSSLService(addr, certPath, keyPath, caCertPath string, fileService portainer.FileService, dataStore dataservices.DataStore, shutdownTrigger context.CancelFunc) (*ssl.Service, error) {
	slices := strings.Split(addr, ":")
	host := slices[0]
	if host == "" {
		host = "0.0.0.0"
	}

	sslService := ssl.NewService(fileService, dataStore, shutdownTrigger)

	err := sslService.Init(host, certPath, keyPath, caCertPath)
	if err != nil {
		return nil, err
	}

	return sslService, nil
}

func initDockerClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService) *docker.ClientFactory {
	return docker.NewClientFactory(signatureService, reverseTunnelService)
}

func initKubernetesClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID string) *kubecli.ClientFactory {
	return kubecli.NewClientFactory(signatureService, reverseTunnelService, dataStore, instanceID)
}

func initNomadClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID string) *clientFactory.ClientFactory {
	return clientFactory.NewClientFactory(signatureService, reverseTunnelService, dataStore, instanceID)
}

func initSnapshotService(
	snapshotInterval string,
	dataStore dataservices.DataStore,
	dockerClientFactory *docker.ClientFactory,
	kubernetesClientFactory *kubecli.ClientFactory,
	nomadClientFactory *clientFactory.ClientFactory,
	shutdownCtx context.Context,
) (portaineree.SnapshotService, error) {
	dockerSnapshotter := docker.NewSnapshotter(dockerClientFactory)
	kubernetesSnapshotter := kubernetes.NewSnapshotter(kubernetesClientFactory)
	nomadSnapshotter := nomadSnapshot.NewSnapshotter(nomadClientFactory)

	snapshotService, err := snapshot.NewService(snapshotInterval, dataStore, dockerSnapshotter, kubernetesSnapshotter, nomadSnapshotter, shutdownCtx)
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

func updateSettingsFromFlags(dataStore dataservices.DataStore, flags *portaineree.CLIFlags) error {
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

	if agentKey, ok := os.LookupEnv("AGENT_SECRET"); ok {
		settings.AgentSecret = agentKey
	} else {
		settings.AgentSecret = ""
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
	} else if *flags.HTTPEnabled {
		sslSettings.HTTPEnabled = true
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
func enableFeaturesFromFlags(dataStore dataservices.DataStore, flags *portaineree.CLIFlags) error {
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

		log.Info().Str("feature", string(*correspondingFeature)).Bool("state", featureState).Msg("")

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
		log.Fatal().Err(err).Msg("failed checking for existing key pair")
	}

	if existingKeyPair {
		return loadAndParseKeyPair(fileService, signatureService)
	}
	return generateAndStoreKeyPair(fileService, signatureService)
}

func createTLSSecuredEndpoint(flags *portaineree.CLIFlags, dataStore dataservices.DataStore, snapshotService portaineree.SnapshotService) error {
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
		TagIDs:             []portaineree.TagID{},
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
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
		log.Error().
			Str("endpoint", endpoint.Name).
			Str("URL", endpoint.URL).
			Err(err).
			Msg("environment snapshot error")
	}

	return dataStore.Endpoint().Create(endpoint)
}

func createUnsecuredEndpoint(endpointURL string, dataStore dataservices.DataStore, snapshotService portaineree.SnapshotService) error {
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
		TagIDs:             []portaineree.TagID{},
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
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
		log.Error().
			Str("endpoint", endpoint.Name).
			Str("URL", endpoint.URL).Err(err).
			Msg("environment snapshot error")
	}

	return dataStore.Endpoint().Create(endpoint)
}

func initEndpoint(flags *portaineree.CLIFlags, dataStore dataservices.DataStore, snapshotService portaineree.SnapshotService) error {
	if *flags.EndpointURL == "" {
		return nil
	}

	endpoints, err := dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	if len(endpoints) > 0 {
		log.Info().Msg("instance already has defined environments, skipping the environment defined via CLI")

		return nil
	}

	if *flags.TLS || *flags.TLSSkipVerify {
		return createTLSSecuredEndpoint(flags, dataStore, snapshotService)
	}
	return createUnsecuredEndpoint(*flags.EndpointURL, dataStore, snapshotService)
}

func updateLicenseKeyFromFlags(licenseService portaineree.LicenseService, licenseKey *string) error {
	if licenseKey == nil || *licenseKey == "" {
		return nil
	}

	_, err := licenseService.AddLicense(*licenseKey)
	return errors.WithMessage(err, "failed to add license")
}

func loadEncryptionSecretKey(keyfilename string) []byte {
	content, err := os.ReadFile(path.Join("/run/secrets", keyfilename))
	if err != nil {
		if os.IsNotExist(err) {
			log.Info().Str("filename", keyfilename).Msg("encryption key file not present")
		} else {
			log.Info().Err(err).Msg("error reading encryption key file")
		}

		return nil
	}

	// return a 32 byte hash of the secret (required for AES)
	hash := sha256.Sum256(content)
	return hash[:]
}

func buildServer(flags *portaineree.CLIFlags) portainer.Server {
	shutdownCtx, shutdownTrigger := context.WithCancel(context.Background())

	fileService := initFileService(*flags.Data)
	encryptionKey := loadEncryptionSecretKey(*flags.SecretKeyName)
	if encryptionKey == nil {
		log.Info().Msg("proceeding without encryption key")
	}

	dataStore := initDataStore(flags, encryptionKey, fileService, shutdownCtx)

	instanceID, err := dataStore.Version().InstanceID()
	if err != nil {
		log.Fatal().Err(err).Msg("failed getting instance id")
	}

	apiKeyService := initAPIKeyService(dataStore)

	_, err = dataStore.Settings().Settings()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	jwtService, err := initJWTService(dataStore)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing JWT service")
	}

	licenseService := license.NewService(dataStore, shutdownCtx)
	if err = licenseService.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed initializing license service")
	}

	err = enableFeaturesFromFlags(dataStore, flags)
	if err != nil {
		log.Fatal().Err(err).Msg("failed enabling feature flag")
	}

	ldapService := initLDAPService()

	oauthService := initOAuthService()

	gitService := initGitService()

	openAMTService := openamt.NewService()

	cryptoService := initCryptoService()

	digitalSignatureService := initDigitalSignatureService()

	edgeService := edge.NewService(dataStore, fileService)

	sslService, err := initSSLService(*flags.AddrHTTPS, *flags.SSLCert, *flags.SSLKey, *flags.SSLCACert, fileService, dataStore, shutdownTrigger)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	sslSettings, err := sslService.GetSSLSettings()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get SSL settings")
	}

	err = initKeyPair(fileService, digitalSignatureService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing key pair")
	}

	reverseTunnelService := chisel.NewService(dataStore, shutdownCtx)

	instanceID, err = dataStore.Version().InstanceID()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get datastore version")
	}

	dockerClientFactory := initDockerClientFactory(digitalSignatureService, reverseTunnelService)
	kubernetesClientFactory := initKubernetesClientFactory(digitalSignatureService, reverseTunnelService, dataStore, instanceID)
	nomadClientFactory := initNomadClientFactory(digitalSignatureService, reverseTunnelService, dataStore, instanceID)

	snapshotService, err := initSnapshotService(*flags.SnapshotInterval, dataStore, dockerClientFactory, kubernetesClientFactory, nomadClientFactory, shutdownCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing snapshot service")
	}
	snapshotService.Start()

	authorizationService := authorization.NewService(dataStore)
	authorizationService.K8sClientFactory = kubernetesClientFactory

	cloudClusterSetupService := cloud.NewCloudClusterSetupService(dataStore, fileService, kubernetesClientFactory, snapshotService, authorizationService, shutdownCtx)
	cloudClusterSetupService.Start()

	cloudClusterInfoService := cloud.NewCloudInfoService(dataStore, shutdownCtx)
	cloudClusterInfoService.Start()

	kubernetesTokenCacheManager := kubeproxy.NewTokenCacheManager()

	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService(*flags.BaseURL, *flags.AddrHTTPS, sslSettings.CertPath)

	userActivityService, userActivityStore := initUserActivity(*flags.Data, *flags.MaxBatchSize, *flags.MaxBatchDelay, *flags.InitialMmapSize, shutdownCtx)

	proxyManager := proxy.NewManager(dataStore, digitalSignatureService, reverseTunnelService, dockerClientFactory, kubernetesClientFactory, kubernetesTokenCacheManager, authorizationService, userActivityService, gitService)

	reverseTunnelService.ProxyManager = proxyManager

	dockerConfigPath := fileService.GetDockerConfigPath()

	composeStackManager := initComposeStackManager(*flags.Assets, dockerConfigPath, reverseTunnelService, proxyManager)

	swarmStackManager, err := initSwarmStackManager(*flags.Assets, dockerConfigPath, digitalSignatureService, fileService, reverseTunnelService, dataStore)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing swarm stack manager")
	}

	kubernetesDeployer := initKubernetesDeployer(authorizationService, kubernetesTokenCacheManager, kubernetesClientFactory, dataStore, reverseTunnelService, digitalSignatureService, proxyManager, *flags.Assets)

	helmPackageManager, err := initHelmPackageManager(*flags.Assets)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing helm package manager")
	}

	err = updateSettingsFromFlags(dataStore, flags)
	if err != nil {
		log.Fatal().Err(err).Msg("failed updating settings from flags")
	}

	err = updateLicenseKeyFromFlags(licenseService, flags.LicenseKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed updating license key from flags")
	}

	err = edge.LoadEdgeJobs(dataStore, reverseTunnelService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed loading edge jobs from database")
	}

	applicationStatus := initStatus(instanceID)

	demoService := demo.NewService()
	if *flags.DemoEnvironment {
		err := demoService.Init(dataStore, cryptoService)
		if err != nil {
			log.Fatal().Err(err).Msg("failed initializing demo environment")
		}
	}

	err = initEndpoint(flags, dataStore, snapshotService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing environment")
	}

	adminPasswordHash := ""
	if *flags.AdminPasswordFile != "" {
		content, err := fileService.GetFileContent(*flags.AdminPasswordFile, "")
		if err != nil {
			log.Fatal().Err(err).Msg("failed getting admin password file")
		}

		adminPasswordHash, err = cryptoService.Hash(strings.TrimSuffix(string(content), "\n"))
		if err != nil {
			log.Fatal().Err(err).Msg("failed hashing admin password")
		}
	} else if *flags.AdminPassword != "" {
		adminPasswordHash = *flags.AdminPassword
	}

	if adminPasswordHash != "" {
		users, err := dataStore.User().UsersByRole(portaineree.AdministratorRole)
		if err != nil {
			log.Fatal().Err(err).Msg("failed getting admin user")
		}

		if len(users) == 0 {
			log.Info().Msg("created admin user with the given password.")
			user := &portaineree.User{
				Username:                "admin",
				Role:                    portaineree.AdministratorRole,
				Password:                adminPasswordHash,
				PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
			}

			err := dataStore.User().Create(user)
			if err != nil {
				log.Fatal().Err(err).Msg("failed creating admin user")
			}
		} else {
			log.Info().Msg("instance already has an administrator user defined, skipping admin password related flags.")
		}
	}

	err = reverseTunnelService.StartTunnelServer(*flags.TunnelAddr, *flags.TunnelPort, snapshotService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed starting tunnel server")
	}

	if !*flags.DemoEnvironment {
		err = licenseService.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("failed starting license service")
		}
	}

	scheduler := scheduler.NewScheduler(shutdownCtx)
	stackDeployer := stacks.NewStackDeployer(swarmStackManager, composeStackManager, kubernetesDeployer)
	stacks.StartStackSchedules(scheduler, stackDeployer, dataStore, gitService, userActivityService)

	sslDBSettings, err := dataStore.SSLSettings().Settings()
	if err != nil {
		log.Fatal().Msg("failed to fetch SSL settings from DB")
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
		EdgeService:                 *edgeService,
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
		KubeClusterAccessService:    kubeClusterAccessService,
		SignatureService:            digitalSignatureService,
		SnapshotService:             snapshotService,
		SSLService:                  sslService,
		DockerClientFactory:         dockerClientFactory,
		UserActivityService:         userActivityService,
		UserActivityStore:           userActivityStore,
		KubernetesClientFactory:     kubernetesClientFactory,
		NomadClientFactory:          nomadClientFactory,
		Scheduler:                   scheduler,
		ShutdownCtx:                 shutdownCtx,
		ShutdownTrigger:             shutdownTrigger,
		StackDeployer:               stackDeployer,
		CloudClusterSetupService:    cloudClusterSetupService,
		CloudClusterInfoService:     cloudClusterInfoService,
		DemoService:                 demoService,
	}
}

func main() {
	configureLogger()

	flags := initCLI()

	setLoggingLevel(*flags.LogLevel)

	for {
		server := buildServer(flags)
		log.Info().
			Str("version", portainer.APIVersion).
			Str("build_number", build.BuildNumber).
			Str("image_tag", build.ImageTag).
			Str("nodejs_version", build.NodejsVersion).
			Str("yarn_version", build.YarnVersion).
			Str("webpack_version", build.WebpackVersion).
			Str("go_version", build.GoVersion).
			Msg("starting Portainer")

		err := server.Start()
		log.Info().Err(err).Msg("HTTP server exited")
	}
}

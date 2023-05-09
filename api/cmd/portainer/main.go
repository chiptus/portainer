package main

import (
	"context"
	"crypto/sha256"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	libstack "github.com/portainer/docker-compose-wrapper"
	"github.com/portainer/docker-compose-wrapper/compose"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/build"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/cli"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/database"
	"github.com/portainer/portainer-ee/api/database/boltdb"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/docker"
	dockerclient "github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/exec"
	"github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/http"
	"github.com/portainer/portainer-ee/api/http/proxy"
	kubeproxy "github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	"github.com/portainer/portainer-ee/api/internal/ssl"
	"github.com/portainer/portainer-ee/api/internal/update"
	"github.com/portainer/portainer-ee/api/jwt"
	"github.com/portainer/portainer-ee/api/kubernetes"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/ldap"
	"github.com/portainer/portainer-ee/api/license"
	"github.com/portainer/portainer-ee/api/nomad/clientFactory"
	nomadSnapshot "github.com/portainer/portainer-ee/api/nomad/snapshot"
	"github.com/portainer/portainer-ee/api/oauth"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/useractivity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/git"
	"github.com/portainer/portainer/api/hostmanagement/openamt"
	"github.com/portainer/portainer/api/platform"
	"github.com/portainer/portainer/pkg/featureflags"
	"github.com/portainer/portainer/pkg/libhelm"

	"github.com/gofrs/uuid"
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
		if isNew {
			log.Fatal().Msg("cannot rollback to CE, no previous version found")
		}

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
		instanceId, err := uuid.NewV4()
		if err != nil {
			log.Fatal().Err(err).Msg("failed generating instance id")
		}

		v := models.Version{
			SchemaVersion: portaineree.APIVersion,
			Edition:       int(portaineree.PortainerEE),
			InstanceID:    instanceId.String(),
		}
		store.VersionService.UpdateVersion(&v)

		err = updateSettingsFromFlags(store, flags)
		if err != nil {
			log.Fatal().Err(err).Msg("failed updating settings from flags")
		}
	} else {
		err = store.MigrateData()
		if err != nil {
			log.Fatal().Err(err).Msg("failed migration")
		}
	}

	// this is for the db restore functionality - needs more tests.
	go func() {
		<-shutdownCtx.Done()
		defer connection.Close()
	}()

	return store
}

func initComposeStackManager(composeDeployer libstack.Deployer, reverseTunnelService portaineree.ReverseTunnelService, proxyManager *proxy.Manager) portaineree.ComposeStackManager {
	composeWrapper, err := exec.NewComposeStackManager(composeDeployer, proxyManager)
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
	return libhelm.NewHelmPackageManager(
		libhelm.HelmConfig{BinaryPath: assetsPath},
	)
}

func initAPIKeyService(datastore dataservices.DataStore) apikey.APIKeyService {
	return apikey.NewAPIKeyService(datastore.APIKeyRepository(), datastore.User())
}

func initJWTService(userSessionTimeout string, dataStore dataservices.DataStore) (dataservices.JWTService, error) {
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

func initGitService(shutdownCtx context.Context) portainer.GitService {
	return git.NewService(shutdownCtx)
}

func initSSLService(addr, certPath, keyPath, caCertPath, mTLSCertPath, mTLSKeyPath, mTLSCaCertPath string, fileService portaineree.FileService, dataStore dataservices.DataStore, shutdownTrigger context.CancelFunc) (*ssl.Service, error) {
	slices := strings.Split(addr, ":")
	host := slices[0]
	if host == "" {
		host = "0.0.0.0"
	}

	if mTLSCertPath != "" && mTLSKeyPath != "" && mTLSCaCertPath != "" {
		err := dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			settings, err := tx.Settings().Settings()
			if err != nil {
				return err
			}

			settings.Edge.MTLS.UseSeparateCert = true

			return tx.Settings().UpdateSettings(settings)
		})
		if err != nil {
			log.Error().Err(err).Msg("could not update the settings")
		}
	}

	sslService := ssl.NewService(fileService, dataStore, shutdownTrigger)

	err := sslService.Init(host, certPath, keyPath, caCertPath, mTLSCertPath, mTLSKeyPath, mTLSCaCertPath)
	if err != nil {
		return nil, err
	}

	return sslService, nil
}

func initDockerClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService) *dockerclient.ClientFactory {
	return dockerclient.NewClientFactory(signatureService, reverseTunnelService)
}

func initKubernetesClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID, addrHTTPS, userSessionTimeout string) (*kubecli.ClientFactory, error) {
	return kubecli.NewClientFactory(signatureService, reverseTunnelService, dataStore, instanceID, addrHTTPS, userSessionTimeout)
}

func initNomadClientFactory(signatureService portaineree.DigitalSignatureService, reverseTunnelService portaineree.ReverseTunnelService, dataStore dataservices.DataStore, instanceID string) *clientFactory.ClientFactory {
	return clientFactory.NewClientFactory(signatureService, reverseTunnelService, dataStore, instanceID)
}

func initSnapshotService(
	snapshotInterval string,
	dataStore dataservices.DataStore,
	dockerClientFactory *dockerclient.ClientFactory,
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

	return dataStore.SSLSettings().UpdateSettings(sslSettings)
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

func cleanUpGhostUpdaterStacks(ctx context.Context) {
	// retry three times to make sure that the updater container exits by itself.
	// It is because if the updater container is forced to remove, the previous agent
	// container can be skipped to be removed by updater container, which will result
	// in the previous ce container being a ghost container.
	err := docker.Retry(ctx, 3, 30*time.Second, docker.ScanAndCleanUpGhostUpdaterContainers)
	if err != nil {
		log.Warn().Err(err).Msg("unable to clean up ghost updater stack")
	}
}

func buildServer(flags *portaineree.CLIFlags) portainer.Server {
	shutdownCtx, shutdownTrigger := context.WithCancel(context.Background())

	if flags.FeatureFlags != nil {
		featureflags.Parse(*flags.FeatureFlags, portaineree.SupportedFeatureFlags)
	}

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

	settings, err := dataStore.Settings().Settings()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	jwtService, err := initJWTService(settings.UserSessionTimeout, dataStore)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing JWT service")
	}

	ldapService := initLDAPService()

	oauthService := initOAuthService()

	gitService := initGitService(shutdownCtx)

	openAMTService := openamt.NewService()

	cryptoService := initCryptoService()

	digitalSignatureService := initDigitalSignatureService()

	edgeAsyncService := edgeasync.NewService(dataStore, fileService)

	edgeStacksService := edgestacks.NewService(dataStore, edgeAsyncService)

	sslService, err := initSSLService(*flags.AddrHTTPS,
		*flags.SSLCert, *flags.SSLKey, *flags.SSLCACert,
		*flags.MTLSCert, *flags.MTLSKey, *flags.MTLSCACert,
		fileService, dataStore, shutdownTrigger)
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

	dockerClientFactory := initDockerClientFactory(digitalSignatureService, reverseTunnelService)
	kubernetesClientFactory, err := initKubernetesClientFactory(digitalSignatureService, reverseTunnelService, dataStore, instanceID, *flags.AddrHTTPS, settings.UserSessionTimeout)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing Kubernetes Client Factory service")
	}
	nomadClientFactory := initNomadClientFactory(digitalSignatureService, reverseTunnelService, dataStore, instanceID)

	snapshotService, err := initSnapshotService(*flags.SnapshotInterval, dataStore, dockerClientFactory, kubernetesClientFactory, nomadClientFactory, shutdownCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing snapshot service")
	}
	snapshotService.Start()

	licenseService := license.NewService(dataStore, shutdownCtx, snapshotService)
	if err = licenseService.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed initializing license service")
	}

	authorizationService := authorization.NewService(dataStore)
	authorizationService.K8sClientFactory = kubernetesClientFactory

	kubernetesTokenCacheManager := kubeproxy.NewTokenCacheManager()

	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService(*flags.BaseURL, *flags.AddrHTTPS, sslSettings.CertPath)

	userActivityService, userActivityStore := initUserActivity(*flags.Data, *flags.MaxBatchSize, *flags.MaxBatchDelay, *flags.InitialMmapSize, shutdownCtx)

	proxyManager := proxy.NewManager(dataStore, digitalSignatureService, reverseTunnelService, dockerClientFactory, kubernetesClientFactory, kubernetesTokenCacheManager, authorizationService, userActivityService, gitService)

	reverseTunnelService.ProxyManager = proxyManager

	kubernetesDeployer := initKubernetesDeployer(authorizationService, kubernetesTokenCacheManager, kubernetesClientFactory, dataStore, reverseTunnelService, digitalSignatureService, proxyManager, *flags.Assets)

	cloudClusterSetupService := cloud.NewCloudClusterSetupService(dataStore, fileService, kubernetesClientFactory, snapshotService, authorizationService, shutdownCtx, kubernetesDeployer)
	cloudClusterSetupService.Start()

	cloudClusterInfoService := cloud.NewCloudInfoService(dataStore, shutdownCtx)
	cloudClusterInfoService.Start()

	dockerConfigPath := fileService.GetDockerConfigPath()

	composeDeployer, err := compose.NewComposeDeployer(*flags.Assets, dockerConfigPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing compose deployer")
	}

	composeStackManager := initComposeStackManager(composeDeployer, reverseTunnelService, proxyManager)

	swarmStackManager, err := initSwarmStackManager(*flags.Assets, dockerConfigPath, digitalSignatureService, fileService, reverseTunnelService, dataStore)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing swarm stack manager")
	}

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
		log.Warn().Err(err).Msg("failed updating license key from flags")
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

	// channel to control when the admin user is created
	adminCreationDone := make(chan struct{}, 1)

	go endpointutils.InitEndpoint(shutdownCtx, adminCreationDone, flags, dataStore, snapshotService)

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

			// notify the admin user is created, the endpoint initialization can start
			adminCreationDone <- struct{}{}
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
	stackDeployer := deployments.NewStackDeployer(swarmStackManager, composeStackManager, kubernetesDeployer, dockerClientFactory, dataStore)
	deployments.StartStackSchedules(scheduler, stackDeployer, dataStore, gitService, userActivityService)

	sslDBSettings, err := dataStore.SSLSettings().Settings()
	if err != nil {
		log.Fatal().Msg("failed to fetch SSL settings from DB")
	}

	updateService, err := update.NewService(*flags.Assets, composeDeployer, kubernetesClientFactory)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing update service")
	}

	// Our normal migrations run as part of the database initialization
	// but some more complex migrations require access to a kubernetes or docker
	// client. Therefore we run a separate migration process just before
	// starting the server.
	postInitMigrator := datastore.NewPostInitMigrator(
		kubernetesClientFactory,
		dockerClientFactory,
		dataStore,
		*flags.Assets,
		kubernetesDeployer,
	)
	if err := postInitMigrator.PostInitMigrate(); err != nil {
		log.Fatal().Err(err).Msg("failure during post init migrations")
	}

	currentPlatform, err := platform.DetermineContainerPlatform()
	if err != nil {
		log.Warn().Err(err).Msg("failed to determine the current container platform")
	} else {
		switch currentPlatform {
		case platform.PlatformDockerStandalone, platform.PlatformDockerSwarm:
			// if the current container is upgraded from CE version, the below goroutine
			// will try to remove the ghost updater stack. Otherwise, it will exit automatically
			go cleanUpGhostUpdaterStacks(shutdownCtx)
			break
		}
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
		EdgeAsyncService:            edgeAsyncService,
		EdgeStacksService:           edgeStacksService,
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
		UpdateService:               updateService,
		AdminCreationDone:           adminCreationDone,
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	configureLogger()
	setLoggingMode("PRETTY")

	flags := initCLI()

	setLoggingLevel(*flags.LogLevel)
	setLoggingMode(*flags.LogMode)

	for {
		server := buildServer(flags)
		log.Info().
			Str("version", portaineree.APIVersion).
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

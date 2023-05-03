package endpointutils

import (
	"context"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/client"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	log "github.com/rs/zerolog/log"
)

// InitEndpoint controls the workflow to initialize the primary endpoint.
// When installing Portainer in a Docker Cluster using the yaml file provided in the official
// documentation (https://docs.portainer.io/start/install/server/swarm/linux), the "primary"
// endpoint is initialized before the admin user is created. This triggers the creation of a
// snapshot of the environment in the background, which includes the agent saving the signature
// from the first request made by the server. However, if a user restores Portainer from a backup
// instead of creating a new admin user, the server will not be able to connect to the agent because
// the saved signature will not match. To solve this issue, this solution proposes to wait for
// the admin user to be created before initializing the primary endpoint. This way, the agent
// will save the signature from the first request after the admin user is created, ensuring that
// it matches in the event of a backup restoration.
func InitEndpoint(shutdownCtx context.Context, adminCreationDone <-chan struct{}, flags *portaineree.CLIFlags, dataStore dataservices.DataStore, snapshotService portaineree.SnapshotService) {

	select {
	case <-shutdownCtx.Done():
		log.Debug().Msg("shutdown endpoint initalization")

	case <-adminCreationDone:
		// Wait for the admin user to be created before initializing the primary endpoint
		// The admin user can be created in two ways:
		// 1. Using the CLI with the --admin-password flag
		// 2. Using the API with the /api/users/admin/init endpoint
		log.Debug().Msg("init primary endpoint")

		err := initEndpoint(flags, dataStore, snapshotService)
		if err != nil {
			log.Fatal().Err(err).Msg("failed initializing environment")
		}
	}
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

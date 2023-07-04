package migrator

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	mk8s "github.com/portainer/portainer-ee/api/cloud/microk8s"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/internal/url"
	"github.com/rs/zerolog/log"
)

func (migrator *Migrator) assignEdgeGroupsToEdgeUpdatesForDB100() error {
	updates, err := migrator.edgeUpdateService.List()
	if err != nil {
		return err
	}

	for idx := range updates {
		update := updates[idx]
		edgeStack, err := migrator.edgeStackService.EdgeStack(update.EdgeStackID)
		if err != nil {
			return err
		}

		environmentIds := make([]portaineree.EndpointID, len(update.EnvironmentsPreviousVersions))
		i := 0
		for id := range update.EnvironmentsPreviousVersions {
			environmentIds[i] = id
			i++
		}

		edgeGroup := &portaineree.EdgeGroup{
			Name:         edgeStack.Name,
			Endpoints:    environmentIds,
			EdgeUpdateID: int(update.ID),
		}

		err = migrator.edgeGroupService.Create(edgeGroup)
		if err != nil {
			return err
		}

		update.EdgeGroupIDs = edgeStack.EdgeGroups
		err = migrator.edgeUpdateService.Update(update.ID, &update)
		if err != nil {
			return err
		}

		edgeStack.EdgeGroups = []portaineree.EdgeGroupID{edgeGroup.ID}
		err = migrator.edgeStackService.UpdateEdgeStack(edgeStack.ID, edgeStack)
		if err != nil {
			return err
		}
	}

	return nil
}

// rebuildEdgeStackFileSystemWithVersionForDB100 creates the edge stack version folder if needed.
// This is needed for backward compatibility with edge stacks created before the
// edge stack version folder was introduced.
func (migrator *Migrator) rebuildEdgeStackFileSystemWithVersionForDB100() error {
	edgeStacks, err := migrator.edgeStackService.EdgeStacks()
	if err != nil {
		return err
	}

	for _, edgeStack := range edgeStacks {
		commitHash := ""
		if edgeStack.GitConfig != nil {
			commitHash = edgeStack.GitConfig.ConfigHash
		}

		edgeStackIdentifier := strconv.Itoa(int(edgeStack.ID))
		edgeStackVersionFolder := migrator.fileService.GetEdgeStackProjectPathByVersion(edgeStackIdentifier, edgeStack.Version, commitHash)

		// Conduct the source folder checks to avoid unnecessary error return
		// In the normal case, the source folder should exist, However, there is a chance that
		// the edge stack folder was deleted by the user, but the edge stack id is still in the
		// database. In this case, we should skip folder migration
		sourceExists, err := migrator.fileService.FileExists(edgeStack.ProjectPath)
		if err != nil {
			log.Warn().
				Err(err).
				Int("edgeStackID", int(edgeStack.ID)).
				Msg("failed to check if edge stack project folder exists")
			continue
		}
		if !sourceExists {
			log.Debug().
				Int("edgeStackID", int(edgeStack.ID)).
				Msg("edge stack project folder does not exist, skipping")
			continue
		}

		/*
			We do not need to check if the target folder exists or not, because
			1. There is a chance the edge stack folder already included a version folder that matches
			with our version folder name. But it was added by user or existed in git repository originally.
			In that case, we should still add our version folder as the parent folder. For example:

			Original:                                       After migration:

			└── edge-stacks                                     └── edge-stacks
				└── 1                                               └── 1
					├── docker-compose.yml                              └── v1
					└── v1                                                  ├── docker-compose.yml
																			└── v1
			 2. As the migration function will be only invoked once when the database is upgraded
			 from lower version to 100, we do not need to worry about nested subfolders being created
			 multiple times. For example: /edge-stacks/2/v1/v1/v1/v1/docker-compose.yml
		*/

		err = migrator.fileService.SafeMoveDirectory(edgeStack.ProjectPath, edgeStackVersionFolder)
		if err != nil {
			return fmt.Errorf("failed to copy edge stack %d project folder: %w", edgeStack.ID, err)
		}
	}
	return nil
}

func (migrator *Migrator) updateTunnelServerAddressForDB100() error {
	settings, err := migrator.settingsService.Settings()
	if err != nil {
		return err
	}

	if settings.EdgePortainerURL != "" && settings.Edge.TunnelServerAddress == "" {
		u, err := url.ParseURL(settings.EdgePortainerURL)
		if err != nil {
			return err
		}

		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}

		settings.Edge.TunnelServerAddress = net.JoinHostPort(host, *migrator.flags.TunnelPort)
		log.
			Info().
			Str("EdgePortainerURL", settings.EdgePortainerURL).
			Str("TunnelServerAddress", settings.Edge.TunnelServerAddress).
			Msg("TunnelServerAddress updated")
	}

	return migrator.settingsService.UpdateSettings(settings)
}

func (m *Migrator) updateCloudProviderForDB100() error {
	// get all environments
	environments, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	// The Name field which is used for display was stored, but not the provider.
	// We need to store an unchangeable provider label whose value should never be changed.
	// Also Digital Ocean was missing a space between the words.
	for _, env := range environments {
		if env.CloudProvider != nil {
			switch env.CloudProvider.Name {
			case "Civo":
				env.CloudProvider.Provider = portaineree.CloudProviderCivo
			case "Linode":
				env.CloudProvider.Provider = portaineree.CloudProviderLinode
			case "DigitalOcean":
				env.CloudProvider.Name = "Digital Ocean"
				env.CloudProvider.Provider = portaineree.CloudProviderDigitalOcean
			case "Google Cloud Platform":
				env.CloudProvider.Provider = portaineree.CloudProviderGKE
			case "Azure":
				env.CloudProvider.Provider = portaineree.CloudProviderAzure
			case "Amazon":
				env.CloudProvider.Provider = portaineree.CloudProviderAmazon
			case "MicroK8s":
				env.CloudProvider.Provider = portaineree.CloudProviderMicrok8s
			}

			err = m.endpointService.UpdateEndpoint(env.ID, &env)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Migrator) enableCommunityAddonForDB100() error {
	log.Info().Msg("enable `community` addon on all MicroK8s master nodes")

	// get all environments
	environments, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, env := range environments {
		if env.CloudProvider != nil {
			if env.CloudProvider.Provider == portaineree.CloudProviderMicrok8s {
				nodeIP, _, _ := strings.Cut(env.URL, ":")

				credential, err := m.cloudCredentialService.Read(env.CloudProvider.CredentialID)
				if err != nil {
					log.Error().Err(err).Msgf("unable to retrieve SSH credential information for MicroK8s environment %s", env.URL)
					continue
				}

				// Get all the Nodes
				// Create ssh client with one of the master nodes.
				sshClient, err := sshUtil.NewConnection(
					credential.Credentials["username"],
					credential.Credentials["password"],
					credential.Credentials["passphrase"],
					credential.Credentials["privateKey"],
					nodeIP,
				)
				if err != nil {
					log.Error().Err(err).Msgf("failed creating ssh client for node %s", nodeIP)
					continue
				}
				defer sshClient.Close()

				var respNodes bytes.Buffer
				if err = sshClient.RunCommand(
					"microk8s kubectl get nodes -o json",
					&respNodes,
				); err != nil {
					log.Error().Err(err).Msg("failed to run ssh command on node")
					continue
				}
				nodeIps, err := mk8s.ParseKubernetesNodes(respNodes.Bytes())
				if err != nil {
					log.Error().Err(err).Msg("failed to get the kubernetes node addresses")
					continue
				}

				for _, node := range nodeIps {
					if node.IsMaster {
						sshClientForNode, err := sshUtil.NewConnection(
							credential.Credentials["username"],
							credential.Credentials["password"],
							credential.Credentials["passphrase"],
							credential.Credentials["privateKey"],
							node.IP,
						)
						if err != nil {
							log.Error().Err(err).Msgf("failed to create ssh client for node %s (IP %s)", node.HostName, node.IP)
							continue
						}
						if err = sshClientForNode.RunCommand(
							"microk8s enable community",
							os.Stdout,
						); err != nil {
							log.Error().Err(err).Msgf("while disabling addon community on node %s (IP %s)", node.HostName, node.IP)
						}
						sshClientForNode.Close()
					}
				}
			}
		}
	}

	return nil
}

package migrator

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/chisel/crypto"
	mk8s "github.com/portainer/portainer-ee/api/cloud/microk8s"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/url"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

func (migrator *Migrator) assignEdgeGroupsToEdgeUpdatesForDB100() error {
	updates, err := migrator.edgeUpdateService.ReadAll()
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
		err = migrator.edgeStackService.UpdateEdgeStack(edgeStack.ID, edgeStack, true)
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

		edgeStack.StackFileVersion = edgeStack.Version
		edgeStackVersionFolder := migrator.fileService.GetEdgeStackProjectPathByVersion(edgeStackIdentifier, edgeStack.StackFileVersion, commitHash)

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

		err = migrator.edgeStackService.UpdateEdgeStackFunc(edgeStack.ID, func(edgeStack *portaineree.EdgeStack) {
			edgeStack.StackFileVersion = edgeStack.Version
		})
		if err != nil {
			return fmt.Errorf("failed to update edge stack %d file version: %w", edgeStack.ID, err)
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
				sshClient, err := sshUtil.NewConnectionWithCredentials(
					nodeIP,
					credential,
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
						sshClientForNode, err := sshUtil.NewConnectionWithCredentials(
							node.IP,
							credential,
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

func (m *Migrator) migrateCloudProviderAddonsForDB100() error {
	log.Info().Msg("migrating addons to addons with args for all MicroK8s environments")

	// get all environments
	environments, err := m.endpointService.Endpoints()

	for _, env := range environments {
		if env.CloudProvider != nil {
			if env.CloudProvider.Provider == portaineree.CloudProviderMicrok8s {

				if env.CloudProvider.Addons != nil {
					addons := strings.Split(*env.CloudProvider.Addons, ",")
					for _, addon := range addons {
						env.CloudProvider.AddonsWithArgs = append(env.CloudProvider.AddonsWithArgs, portaineree.MicroK8sAddon{
							Name: addon,
						})
					}

					err = m.endpointService.UpdateEndpoint(env.ID, &env)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// rebuildStackFileSystemWithVersionForDB100 creates the regular stack version folder if needed.
// This is needed for backward compatibility with regular stacks created before the
// regular stack version folder was introduced.
func (migrator *Migrator) rebuildStackFileSystemWithVersionForDB100() error {
	stacks, err := migrator.stackService.ReadAll()
	if err != nil {
		return err
	}

	for _, stack := range stacks {
		commitHash := ""
		if stack.GitConfig != nil {
			commitHash = stack.GitConfig.ConfigHash
		}

		stackIdentifier := strconv.Itoa(int(stack.ID))

		stack.StackFileVersion = 1
		stackVersionFolder := migrator.fileService.GetStackProjectPathByVersion(stackIdentifier, stack.StackFileVersion, commitHash)

		// Conduct the source folder checks to avoid unnecessary error return, same
		// as the above edge stack migration.
		sourceExists, err := migrator.fileService.FileExists(stack.ProjectPath)
		if err != nil {
			log.Warn().
				Err(err).
				Int("stackID", int(stack.ID)).
				Msg("failed to check if stack project folder exists")
			continue
		}
		if !sourceExists {
			log.Debug().
				Int("stackID", int(stack.ID)).
				Msg("stack project folder does not exist, skipping")
			continue
		}

		err = migrator.fileService.SafeMoveDirectory(stack.ProjectPath, stackVersionFolder)
		if err != nil {
			return fmt.Errorf("failed to copy stack %d project folder: %w", stack.ID, err)
		}

		err = migrator.stackService.Update(stack.ID, &stack)
		if err != nil {
			return fmt.Errorf("failed to update stack %d file version: %w", stack.ID, err)
		}

	}
	return nil
}

func (m *Migrator) convertSeedToPrivateKeyForDB100() error {
	var serverInfo *portaineree.TunnelServerInfo

	serverInfo, err := m.TunnelServerService.Info()
	if err != nil {
		if dataservices.IsErrObjectNotFound(err) {
			log.Info().Msg("ServerInfo object not found")
			return nil
		}
		log.Error().
			Err(err).
			Msg("Failed to read ServerInfo from DB")
		return err
	}

	if serverInfo.PrivateKeySeed != "" {
		key, err := crypto.GenerateGo119CompatibleKey(serverInfo.PrivateKeySeed)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to read ServerInfo from DB")
			return err
		}

		err = m.fileService.StoreChiselPrivateKey(key)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to save Chisel private key to disk")
			return err
		}
	} else {
		log.Info().Msg("PrivateKeySeed is blank")
	}

	serverInfo.PrivateKeySeed = ""
	err = m.TunnelServerService.UpdateInfo(serverInfo)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to clean private key seed in DB")
	} else {
		log.Info().Msg("Success to migrate private key seed to private key file")
	}
	return err
}

func (m *Migrator) updateEdgeStackStatusForDB100() error {
	log.Info().Msg("update edge stack status to have deployment steps")

	edgeStacks, err := m.edgeStackService.EdgeStacks()
	if err != nil {
		return err
	}

	for _, edgeStack := range edgeStacks {

		for environmentID, environmentStatus := range edgeStack.Status {
			// skip if status is already updated
			if len(environmentStatus.Status) > 0 {
				continue
			}

			statusArray := []portainer.EdgeStackDeploymentStatus{}
			if environmentStatus.Details.Pending {
				statusArray = append(statusArray, portainer.EdgeStackDeploymentStatus{
					Type: portainer.EdgeStackStatusPending,
					Time: time.Now().Unix(),
				})
			}

			if environmentStatus.Details.Acknowledged {
				statusArray = append(statusArray, portainer.EdgeStackDeploymentStatus{
					Type: portainer.EdgeStackStatusAcknowledged,
					Time: time.Now().Unix(),
				})
			}

			if environmentStatus.Details.Error {
				statusArray = append(statusArray, portainer.EdgeStackDeploymentStatus{
					Type:  portainer.EdgeStackStatusError,
					Error: environmentStatus.Error,
					Time:  time.Now().Unix(),
				})
			}

			if environmentStatus.Details.Ok {
				statusArray = append(statusArray,
					portainer.EdgeStackDeploymentStatus{
						Type: portainer.EdgeStackStatusDeploymentReceived,
						Time: time.Now().Unix(),
					},
					portainer.EdgeStackDeploymentStatus{
						Type: portainer.EdgeStackStatusRunning,
						Time: time.Now().Unix(),
					},
				)
			}

			if environmentStatus.Details.ImagesPulled {
				statusArray = append(statusArray, portainer.EdgeStackDeploymentStatus{
					Type: portainer.EdgeStackStatusImagesPulled,
					Time: time.Now().Unix(),
				})
			}

			if environmentStatus.Details.Remove {
				statusArray = append(statusArray, portainer.EdgeStackDeploymentStatus{
					Type: portainer.EdgeStackStatusRemoving,
					Time: time.Now().Unix(),
				})
			}

			environmentStatus.Status = statusArray

			edgeStack.Status[environmentID] = environmentStatus
		}

		err = m.edgeStackService.UpdateEdgeStack(edgeStack.ID, &edgeStack, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) fixPotentialUpdateScheduleDBCorruptionForDB100() error {
	edgeStacks, err := m.edgeStackService.EdgeStacks()
	if err != nil {
		return err
	}

	edgeStackIDs := make(map[portaineree.EdgeStackID]bool, len(edgeStacks))

	for _, edgeStack := range edgeStacks {
		edgeStackIDs[edgeStack.ID] = true
	}

	endpointRelations, err := m.endpointRelationService.EndpointRelations()
	if err != nil {
		return err
	}

	for _, endpointRelation := range endpointRelations {
		hasCorruptedEdgeStack := false
		for edgeStackID := range endpointRelation.EdgeStacks {
			if _, ok := edgeStackIDs[edgeStackID]; ok {
				continue
			}

			// remove the edge stack ID from the endpoint relation if
			// the edge stack does not exist
			delete(endpointRelation.EdgeStacks, edgeStackID)
			hasCorruptedEdgeStack = true
		}

		if hasCorruptedEdgeStack {
			err = m.endpointRelationService.UpdateEndpointRelation(endpointRelation.EndpointID, &endpointRelation)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

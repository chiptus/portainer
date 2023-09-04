package migrator

import (
	"bytes"
	"net"
	"os"
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

		settings.Edge.TunnelServerAddress = net.JoinHostPort(u.Hostname(), *migrator.flags.TunnelPort)
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

package microk8s

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/rs/zerolog/log"
)

type (
	MicroK8sMasterWorkerNode struct {
		IP            string
		HostName      string
		IsMaster      bool
		Status        string // node status
		UpgradeStatus string // pending,updating,updated,failed
		Error         error
		Unschedulable bool // true if node is unschedulable (drained)
	}

	Microk8sUpgrade struct {
		endpoint  *portaineree.Endpoint
		dataStore dataservices.DataStore

		addons         []string
		endpointIP     string
		nodeIps        []MicroK8sMasterWorkerNode
		credentials    *models.CloudCredential
		currentVersion string
		nextVersion    string
	}
)

func (e MicroK8sMasterWorkerNode) String() string {
	return fmt.Sprintf("HostName: %s (IP: %s), UpgradeStatus: %s, Error: %v", e.HostName, e.IP, e.UpgradeStatus, e.Error)
}

func (e Microk8sUpgrade) Len() int {
	return len(e.nodeIps)
}

func (e Microk8sUpgrade) Swap(i, j int) {
	e.nodeIps[i], e.nodeIps[j] = e.nodeIps[j], e.nodeIps[i]
}

func (e Microk8sUpgrade) Less(i, j int) bool {
	return !e.nodeIps[i].IsMaster && e.nodeIps[j].IsMaster
}

func NewMicrok8sUpgrade(endpoint *portaineree.Endpoint, dataStore dataservices.DataStore) *Microk8sUpgrade {
	return &Microk8sUpgrade{
		endpoint:  endpoint,
		dataStore: dataStore,
	}
}

func (service *Microk8sUpgrade) getNextVersion(current string) string {
	if current == "1.27/stable" {
		return current
	}

	// Fallback to "updating" to current version.
	// This will be used if there are no newer versions.
	previous := current

	for _, v := range MicroK8sVersions {
		if v.Value == current {
			break
		}
		previous = v.Value
	}

	return previous
}

func (u *Microk8sUpgrade) Upgrade() (string, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msg("processing upgrade request")

	u.setMessage(u.endpoint.ID, "Upgrading cluster", "Gathering information about cluster.", "processing")

	nodeIP, _, _ := strings.Cut(u.endpoint.URL, ":")
	u.endpointIP = nodeIP

	credential, err := u.dataStore.CloudCredential().Read(u.endpoint.CloudProvider.CredentialID)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve SSH credential information: %w", err)
	}
	u.credentials = credential

	// Create ssh client with one of the master nodes.
	sshClient, err := sshUtil.NewConnection(
		credential.Credentials["username"],
		credential.Credentials["password"],
		credential.Credentials["passphrase"],
		credential.Credentials["privateKey"],
		nodeIP,
	)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating ssh client")
		return "", err
	}
	defer sshClient.Close()

	// We can't use the microk8s version command as it was added in 1.25.
	// Instead we parse the output from snap.
	var resp bytes.Buffer
	if err = sshClient.RunCommand(
		"snap list",
		&resp,
	); err != nil {
		return "", fmt.Errorf("failed to run ssh command: %w", err)
	}

	if u.currentVersion, err = ParseSnapInstalledVersion(resp.String()); err != nil {
		return "", err
	}

	// Find next version.
	u.nextVersion = u.getNextVersion(u.currentVersion)

	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("current version %s, upgrading to %s", u.currentVersion, u.nextVersion)

	if u.nextVersion == u.currentVersion {
		log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("cluster is already on latest version %s", u.currentVersion)
		return u.currentVersion, nil
	}

	u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Current MicroK8s version: %s, upgrading to: %s", u.currentVersion, u.nextVersion), "processing")

	// Get all the nodes in the cluster.
	// We need to know which nodes are masters and which are workers.
	var respNodes bytes.Buffer
	if err = sshClient.RunCommand(
		"microk8s kubectl get nodes -o json",
		&respNodes,
	); err != nil {
		return "", fmt.Errorf("failed to run ssh command: %w", err)
	}
	if u.nodeIps, err = ParseKubernetesNodes(respNodes.Bytes()); err != nil {
		return "", fmt.Errorf("failed to get the kubernetes node addresses: %w", err)
	}
	sort.Stable(sort.Reverse(u))

	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Number of nodes in the cluster %d", len(u.nodeIps))

	var respAddons bytes.Buffer
	if err = sshClient.RunCommand("microk8s status --format yaml", &respAddons); err != nil {
		return "", fmt.Errorf("failed to get addons: %w", err)
	}
	if u.addons, err = GetEnabledAddons(respAddons.String()); err != nil {
		return "", fmt.Errorf("failed to parse addons: %w", err)
	}

	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("installed addons on cluster %v", u.addons)

	isSingleNodeCluster := len(u.nodeIps) == 1

	for index, node := range u.nodeIps {
		if node.Status == "Ready" {
			u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Upgrading node %s (IP %s).", node.HostName, node.IP), "processing")
			log.Info().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Upgrading node %s (IP %s)", node.HostName, node.IP)

			// Upgrade node
			u.nodeIps[index].UpgradeStatus = "upgrading"
			u.nodeIps[index].Error = nil

			if !isSingleNodeCluster {
				u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Draining node %s (IP %s).", node.HostName, node.IP), "processing")
				log.Info().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Draining node %s (IP %s).", node.HostName, node.IP)
				// Step 1: drain node
				if err = sshClient.RunCommand(
					"microk8s kubectl drain "+node.HostName+" --ignore-daemonsets --delete-emptydir-data",
					os.Stdout,
				); err != nil {
					u.nodeIps[index].UpgradeStatus = "failed"
					u.nodeIps[index].Error = err

					log.Error().Str("provider", portaineree.CloudProviderMicrok8s).Err(err).Msgf("Error in draining node %s (IP %s). Continuing to next node.", node.HostName, node.IP)
					u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Error in draining node %s (IP %s). Continuing to next node.", node.HostName, node.IP), "processing")
					continue
				}

				u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Checking status, node %s (IP %s).", node.HostName, node.IP), "processing")
				log.Info().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Checking status node %s (IP %s).", node.HostName, node.IP)

				count := 0
				// Get the status again
				for {
					time.Sleep(5 * time.Second)

					var respNodes bytes.Buffer
					if err = sshClient.RunCommand(
						"microk8s kubectl get nodes -o json",
						&respNodes,
					); err != nil {
						return "", fmt.Errorf("failed to run ssh command: %w", err)
					}
					u.nodeIps[index].Unschedulable, err = ParseAndCheckIfNodeUnschedulable(respNodes.Bytes(), node.HostName)
					if err != nil {
						log.Debug().Err(err).Msgf("failed to get node status after drain. checkoing again in 5 seconds")
					} else if !u.nodeIps[index].Unschedulable {
						log.Debug().Err(err).Msgf("Node is not set to SchedulingDisabled. checking again in 5 seconds")
					} else {
						u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Node %s (IP %s) status is SchedulingDisabled.", node.HostName, node.IP), "processing")
						log.Info().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Node %s (IP %s) status is SchedulingDisabled.", node.HostName, node.IP)
						break
					}
					count++
					log.Debug().Msgf("Trying node %s (IP %s) status again. count: %d", node.HostName, node.IP, count)
					if count == 5 {
						log.Error().Msgf("failed to get node status after drain. continuing to next node")
						break
					}
				}
			}

			u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Upgrading MicroK8s version on node %s (IP %s).", node.HostName, node.IP), "processing")
			log.Info().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Upgrading MicroK8s version on node %s (IP %s).", node.HostName, node.IP)
			// Step 2: refresh node
			sshClientPerNode, err := sshUtil.NewConnection(
				credential.Credentials["username"],
				credential.Credentials["password"],
				credential.Credentials["passphrase"],
				credential.Credentials["privateKey"],
				node.IP,
			)
			if err != nil {
				log.Debug().Err(err).Msgf("failed creating ssh client for node %s (IP %s)", node.HostName, node.IP)
			} else {
				if err = sshClientPerNode.RunCommand(
					"snap refresh microk8s --channel="+u.nextVersion,
					os.Stdout,
				); err != nil {
					u.nodeIps[index].UpgradeStatus = "failed"
					u.nodeIps[index].Error = err

					log.Error().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Error in upgrading MicroK8s version on node %s (IP %s). Trying to revert MicroK8s version on this node.", node.HostName, node.IP)
					u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Error in upgrading MicroK8s version on node %s (IP %s). Trying to revert MicroK8s version on this node.", node.HostName, node.IP), "processing")

					// Try reverting to previous version
					if err = sshClientPerNode.RunCommand(
						"snap revert microk8s",
						os.Stdout,
					); err != nil {
						u.nodeIps[index].Error = err

						log.Error().Str("provider", portaineree.CloudProviderMicrok8s).Err(err).Msgf("Error when reverting MicroK8s on node %s (IP %s). Continuing to next node.", node.HostName, node.IP)
						u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Error when reverting MicroK8s on node %s (IP %s). Continuing to next node.", node.HostName, node.IP), "processing")
						continue
					}
				}
			}
			sshClientPerNode.Close()

			// Added waiting to allow the microk8s refresh to complete/settle.
			// TODO: check if really required as the dev testing was done with it.
			time.Sleep(4 * time.Second)

			if !isSingleNodeCluster {
				log.Info().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Resuming pod scheduling on node %s (IP %s).", node.HostName, node.IP)
				u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Resuming pod scheduling on node %s (IP %s).", node.HostName, node.IP), "processing")

				// Step 3: uncordon node
				if err = sshClient.RunCommand(
					"microk8s kubectl uncordon "+node.HostName,
					os.Stdout,
				); err != nil {
					u.nodeIps[index].UpgradeStatus = "failed"
					u.nodeIps[index].Error = err

					log.Error().Str("provider", portaineree.CloudProviderMicrok8s).Err(err).Msgf("Error when resuming pod scheduling on node %s (IP %s). Continuing to next node.", node.HostName, node.IP)
					u.setMessage(u.endpoint.ID, "Upgrading cluster", fmt.Sprintf("Error when resuming pod scheduling on node %s (IP %s). Continuing to next node.", node.HostName, node.IP), "processing")
					continue
				}
			}
		}
		u.nodeIps[index].UpgradeStatus = "updated"
		u.nodeIps[index].Error = nil
	}

	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Upgrading addons %v", u.addons)

	allInstallableAddons := types.Microk8sAddonsPayload{
		Addons: GetAllAvailableAddons().GetNames(),
	}

	u.setMessage(u.endpoint.ID, "Upgrading cluster", "Disabling addons", "processing")
	// disable addons
	for _, addon := range u.addons {
		if allInstallableAddons.IndexOf(addon) != -1 {
			log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Disabling addon %s", addon)
			if err = sshClient.RunCommand(
				"microk8s disable "+addon,
				os.Stdout,
			); err != nil {
				log.Error().Str("provider", portaineree.CloudProviderMicrok8s).Err(err).Msgf("while disabling addon %s", addon)
			}
		}
	}

	u.setMessage(u.endpoint.ID, "Upgrading cluster", "Enabling addons", "processing")
	// enable addons
	for _, addon := range u.addons {
		if allInstallableAddons.IndexOf(addon) != -1 {
			log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Enabling addon %s", addon)
			if err = sshClient.RunCommand(
				"microk8s enable "+addon,
				os.Stdout,
			); err != nil {
				log.Error().Str("provider", portaineree.CloudProviderMicrok8s).Err(err).Msgf("while enabling addon %s", addon)
			}
		}
	}

	isError := false
	messages := []string{}
	for _, node := range u.nodeIps {
		if node.Error != nil {
			isError = isError || true
		}
		messages = append(messages, node.String())
	}

	operationStatus := ""
	if isError {
		operationStatus = "error"
	}

	u.setMessage(u.endpoint.ID, "Upgrade completed", "Check Portainer logs for more details<br/><br/>"+strings.Join(messages, "<br/>"), operationStatus)

	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msgf("Upgrade status: %+v", u.nodeIps)

	return u.nextVersion, err
}

func (service *Microk8sUpgrade) setMessage(id portaineree.EndpointID, summary, detail string, operationStatus string) error {
	status := portaineree.EndpointStatusMessage{Summary: summary, Detail: detail, OperationStatus: operationStatus, Operation: "upgrade"}
	err := service.dataStore.Endpoint().SetMessage(id, status)
	if err != nil {
		return fmt.Errorf("unable to update endpoint in database")
	}
	return nil
}

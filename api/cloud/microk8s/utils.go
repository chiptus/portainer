package microk8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// ParseSnapInstalledVersion reads the command line response of `snap list`
// and returns the current installed version of microk8s.
func ParseSnapInstalledVersion(s string) (string, error) {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		words := strings.Fields(line)
		for _, word := range words {
			if word == "microk8s" {
				if len(words) < 4 {
					return "", fmt.Errorf("invalid snap list output: %v", line)
				}
				return words[3], nil
			}
		}
	}
	return "", fmt.Errorf("microk8s not found in snap list: %v", lines)
}

func ParseKubernetesNodesResponse(s string) ([]string, error) {
	var nodeList v1.NodeList

	err := json.Unmarshal([]byte(s), &nodeList)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for _, node := range nodeList.Items {
		ips = append(ips, node.Status.Addresses[0].Address)
	}
	return ips, nil
}

func ParseKubernetesNodes(s []byte) ([]MicroK8sMasterWorkerNode, error) {
	var nodeList v1.NodeList

	err := json.Unmarshal(s, &nodeList)
	if err != nil {
		return nil, err
	}

	microK8sMasterWorkerNode := make([]MicroK8sMasterWorkerNode, 0)
	for _, node := range nodeList.Items {
		microK8sMasterWorkerNode = append(microK8sMasterWorkerNode, MicroK8sMasterWorkerNode{
			IP:            node.Status.Addresses[0].Address,
			HostName:      node.GetObjectMeta().GetName(),
			UpgradeStatus: "pending",
			Status: func() string {
				for _, condition := range node.Status.Conditions {
					if condition.Status == "True" {
						return string(condition.Type)
					}
				}
				return "Unknown"
			}(),
			IsMaster: func() bool {
				for role := range node.ObjectMeta.Labels {
					if role == "node-role.kubernetes.io/master" || role == "node-role.kubernetes.io/control-plane" || role == "node.kubernetes.io/microk8s-controlplane" {
						return true
					}
				}
				return false
			}(),
			Unschedulable: node.Spec.Unschedulable,
		})
	}
	return microK8sMasterWorkerNode, nil
}

func ParseAndCheckIfNodeUnschedulable(s []byte, hostName string) (bool, error) {
	var nodeList v1.NodeList

	err := json.Unmarshal(s, &nodeList)
	if err != nil {
		return false, err
	}
	for _, node := range nodeList.Items {
		if hostName == node.GetObjectMeta().GetName() {
			return node.Spec.Unschedulable, nil
		}
	}
	return false, nil
}

func EnableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, nodeIp string, addon string) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	command := "microk8s enable " + addon
	return sshClient.RunCommand(command, os.Stdout)
}

func DisableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, nodeIp string, addon string) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	command := "microk8s disable " + addon
	return sshClient.RunCommand(command, os.Stdout)
}

func InstallMicrok8sOnNode(user, password, passphrase, privateKey, nodeIp, kubernetesVersion string) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	for i := 0; i < 3; i++ {
		// Try to install microk8s up to 3 times before we give up.
		cmd := "snap install microk8s --classic --channel=" + kubernetesVersion
		log.Info().Msg("MicroK8s install command on " + nodeIp + ": " + cmd)
		err := sshClient.RunCommand(cmd, os.Stdout)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	err = sshClient.RunCommand("microk8s status --wait-ready", os.Stdout)
	if err != nil {
		return err
	}

	// Default set of addons.
	// ha-cluster is automatically enabled when adding more master nodes to the cluster.
	// Newer versions of microk8s enable dns, rbac and helm by default, but we'll just
	// enable them all here to be sure.
	// TODO: From 1.27 onward we can probably remove this code. For now, keep it while we still
	// support older versions.
	addons := []string{"dns", "rbac", "helm", "community"}
	for _, addon := range addons {
		if addon != "community" {
			err = EnableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, nodeIp, addon)
			if err != nil {
				log.Debug().Err(err).Msgf("Failed to enable addon %s on node %s", addon, nodeIp)
			}
		} else {
			// community addon should be enabled on all the master nodes.
			var respNodes bytes.Buffer
			if err = sshClient.RunCommand(
				"microk8s kubectl get nodes -o json",
				&respNodes,
			); err != nil {
				log.Error().Err(err).Msgf("Failed to run ssh command on node %s", nodeIp)
				continue
			}
			nodeIps, err := ParseKubernetesNodes(respNodes.Bytes())
			if err != nil {
				log.Error().Err(err).Msg("Failed to get the kubernetes node addresses")
				continue
			}
			for _, node := range nodeIps {
				err = EnableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, node.IP, addon)
				if err != nil {
					log.Debug().Err(err).Msgf("Failed to enable addon %s on node %s", addon, node.IP)
					continue
				}
			}
		}
	}

	return nil
}

func ExecuteJoinClusterCommandOnNode(user, password, passphrase, privateKey, nodeIp string, joinInfo *microk8sClusterJoinInfo, asWorkerNode bool) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	workerParam := ""
	if asWorkerNode {
		workerParam = "--worker"
	}

	joinClusterCommand := fmt.Sprintf("microk8s join %s %s", workerParam, joinInfo.URLS[0])
	log.Debug().Msgf("Node with ip %s is joining the cluster with command: %s", nodeIp, joinClusterCommand)
	return sshClient.RunCommand(joinClusterCommand, os.Stdout)
}

func ExecuteLeaveClusterCommandOnNode(user, password, passphrase, privateKey, nodeIp string) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	joinCmd := "microk8s leave"
	log.Debug().Msgf("Node with ip %s is leaving the cluster with command: %s", nodeIp, joinCmd)

	err = sshClient.RunCommand(joinCmd, os.Stdout)
	if err != nil {
		return err
	}

	uninstallCmd := "snap remove microk8s --purge"
	log.Debug().Msgf("Node with ip %s is removing the microk8s snap with command: %s", nodeIp, uninstallCmd)

	// uninstall microk8s
	return sshClient.RunCommand("snap remove microk8s --purge", os.Stdout)
}

func ExecuteAnnotateNodeCommandOnNode(user, password, passphrase, privateKey, masterNodeIp, nodeToAnnotate string) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, masterNodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	// Annotate the node with portainer.io/removing-node=true
	annotateCommand := fmt.Sprintf("microk8s kubectl annotate --overwrite node %s portainer.ip/removing-node=true", nodeToAnnotate)

	log.Debug().Msgf("Annotating node %s with command: %s", nodeToAnnotate, annotateCommand)
	return sshClient.RunCommand(annotateCommand, os.Stdout)
}

func ExecuteRemoveNodeCommandOnNode(user, password, passphrase, privateKey, masterNodeIp, nodeToRemove string, force bool) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, masterNodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	removeNodeCmd := fmt.Sprintf("microk8s remove-node %s", nodeToRemove)
	if force {
		removeNodeCmd = removeNodeCmd + " --force"
	}

	log.Debug().Msgf("Node %s is leaving the cluster with command: %s", nodeToRemove, removeNodeCmd)
	return sshClient.RunCommand(removeNodeCmd, os.Stdout)
}

func ExecuteDrainNodeCommandOnNode(user, password, passphrase, privateKey, masterNodeIp, nodeToDrain string) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, masterNodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	removeNodeCmd := fmt.Sprintf("microk8s kubectl drain %s --ignore-daemonsets --delete-emptydir-data", nodeToDrain)

	log.Debug().Msgf("Node %s is draining with command: %s", nodeToDrain, removeNodeCmd)
	return sshClient.RunCommand(removeNodeCmd, os.Stdout)
}

func RetrieveClusterJoinInformation(user, password, passphrase, privateKey, nodeIp string) (*microk8sClusterJoinInfo, error) {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return nil, err
	}
	defer sshClient.Close()

	addNodeCommand := "microk8s add-node --format json"
	var resp bytes.Buffer
	err = sshClient.RunCommand(addNodeCommand, &resp)
	if err != nil {
		return nil, err
	}

	joinInfo := &microk8sClusterJoinInfo{}
	err = json.Unmarshal(resp.Bytes(), joinInfo)
	if err != nil {
		return nil, err
	}

	return joinInfo, nil
}

func RetrieveHostname(user, password, passphrase, privateKey, nodeIp string) (string, error) {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return "", err
	}
	defer sshClient.Close()

	hostnameCommand := "hostname"
	var resp bytes.Buffer
	err = sshClient.RunCommand(hostnameCommand, &resp)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(resp.String(), "\n"), nil
}

func UpdateHostFile(
	user, password, passphrase, privateKey, nodeIp string,
	hostEntries map[string]string,
) error {
	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, nodeIp)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	for ip, hostname := range hostEntries {
		if ip == nodeIp {
			continue
		}
		s := fmt.Sprintf("%s %s", ip, hostname)
		command := fmt.Sprintf("sh -c 'echo \"%s\" >> /etc/hosts'", s)
		err = sshClient.RunCommand(command, os.Stdout)
		if err != nil {
			return err
		}

		// cloud-init workaround
		// On machines created with cloud-init, which is many cloud providers,
		// you need to edit a file in /etc/cloud/templates instead of the main
		// /etc/hosts file in order for your changes to persist on a restart.
		// NOTE: This is a best effort attempt. If the file doesn't exist we
		// skip it and only edit the main hosts file.
		var resp bytes.Buffer
		err = sshClient.RunCommand(
			`grep -o "\/etc\/cloud\/templates\/[^[:space:]]*\.tmpl" /etc/hosts | head -n 1`, // get the first file that matches
			&resp,
		)
		path := resp.String()
		if err != nil || path == "" || strings.ContainsAny(path, " ") {
			continue
		}
		command = fmt.Sprintf("sh -c 'echo \"%s\" >> %s'", s, path)
		// Not worrying about errors since it will just be the file missing.
		sshClient.RunCommand(command, os.Stdout)
	}

	return nil
}

func SetupHostEntries(user, password, passphrase, privateKey string, nodeIps []string) error {
	// hostEntries is a mapping of nodeIP to hostname.
	hostEntries := make(map[string]string)

	// Build the list of all host entries.
	for _, nodeIp := range nodeIps {
		hostname, err := RetrieveHostname(user, password, passphrase, privateKey, nodeIp)
		if err != nil {
			return err
		}

		hostEntries[nodeIp] = hostname
	}

	// Update each of the nodes with the list of host entries.
	for _, nodeIp := range nodeIps {
		err := UpdateHostFile(user, password, passphrase, privateKey, nodeIp, hostEntries)
		if err != nil {
			return err
		}
	}
	return nil
}

// parseAddonResponse reads the command line response of `microk8s status` and
// returns a list of installed addons.
func ParseAddonResponse(s string) (*Microk8sStatusResponse, error) {
	status := &Microk8sStatusResponse{}
	err := yaml.Unmarshal([]byte(s), status)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func GetEnabledAddons(s string) ([]string, error) {
	status, err := ParseAddonResponse(s)
	if err != nil {
		return nil, err
	}

	addons := make([]string, 0)
	for _, addon := range status.Addons {
		if addon.Status == "enabled" {
			addons = append(addons, addon.Name)
		}
	}
	return addons, nil
}

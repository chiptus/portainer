package microk8s

import (
	"bytes"
	"fmt"
	"os"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func DeleteCluster(endpoint *portaineree.Endpoint, credentials *models.CloudCredential) error {

	// Get a list of all nodes addresses in the cluster
	// for each node:
	// uninstall microk8s with "snap remove microk8s --purge"
	// there is no need to gracefully remove the node from the cluster since we're destroying it anyway

	// The uninstall process should happen on all nodes in parallel
	masterNode := UrlToMasterNode(endpoint.URL)

	nodes, err := GetNodeList(masterNode, credentials)
	if err != nil {
		return fmt.Errorf("unable to get a list of nodes for the microk8s cluster: %w", err)
	}

	var g errgroup.Group

	for _, node := range nodes {
		node := node // local copy of node for passing into goroutine
		g.Go(func() error {
			err := UninstallMicrok8s(node, credentials)
			if err != nil {
				log.Error().Err(err).Msgf("unable to uninstall microk8s from node %s", node)
			}
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("unable to uninstall microk8s from some nodes, see portainer log for details")
	}

	return nil
}

func GetNodeList(masterNode string, credentials *models.CloudCredential) ([]string, error) {
	sshClient, err := ssh.NewConnectionWithCredentials(masterNode, credentials)
	if err != nil {
		return nil, err
	}
	defer sshClient.Close()

	log.Debug().Msgf("Getting a list of nodes to uninstall microk8s from")

	var resp bytes.Buffer
	err = sshClient.RunCommand("microk8s kubectl get nodes -o json", &resp)
	if err != nil {
		return nil, err
	}

	nodes, err := ParseKubernetesNodes(resp.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable to parse kubernetes nodes: %w", err)
	}

	var nodeAddresses []string
	for _, node := range nodes {
		nodeAddresses = append(nodeAddresses, node.IP)
	}

	return nodeAddresses, nil
}

func UninstallMicrok8s(node string, credentials *models.CloudCredential) error {
	conn, err := ssh.NewConnectionWithCredentials(node, credentials)
	if err != nil {
		return fmt.Errorf("failed to uninstall MicroK8s: cannot create client connection: %w", err)
	}
	defer conn.Close()

	log.Debug().Msgf("Uninstalling microk8s from node %s", node)
	return conn.RunCommand("snap remove microk8s --purge", os.Stdout)
}

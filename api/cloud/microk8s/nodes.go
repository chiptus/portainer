package microk8s

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/util"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type (
	Microk8sCluster struct {
		DataStore         dataservices.DataStore
		KubeClientFactory *kubecli.ClientFactory
		SetMessage        func(title, message string, status portaineree.EndpointOperationStatus) error
	}
)

func (service *Microk8sCluster) GetAddons(credential *models.CloudCredential, environmentID int) (*Microk8sStatusResponse, error) {
	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("processing get info request")

	// Gather nodeIP from environmentID
	endpoint, err := service.DataStore.Endpoint().Endpoint(portainer.EndpointID(environmentID))
	if err != nil {
		log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("failed looking up environment nodeIP")
		return nil, err
	}
	nodeIP, _, _ := strings.Cut(endpoint.URL, ":")

	// Gather current addon list.
	sshClient, err := sshUtil.NewConnectionWithCredentials(
		nodeIP,
		credential,
	)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating ssh client")
		return nil, err
	}
	defer sshClient.Close()

	var respSnapList bytes.Buffer
	if err = sshClient.RunCommand(
		"snap list",
		&respSnapList,
	); err != nil {
		return nil, fmt.Errorf("failed to run ssh command: %w", err)
	}

	currentVersion, err := ParseSnapInstalledVersion(respSnapList.String())
	if err != nil {
		return nil, fmt.Errorf("failed to run ssh command: %w", err)
	}

	var resp bytes.Buffer
	err = sshClient.RunCommand("microk8s status --format yaml", &resp)
	if err != nil {
		return nil, err
	}

	mk8sStatus, err := ParseAddonResponse(resp.String())
	if err != nil {
		return nil, err
	}
	// Fill with arguments from endpoint.CloudProvider.AddonsWithArgs
	for i, addon := range mk8sStatus.Addons {
		for _, endAddon := range endpoint.CloudProvider.AddonsWithArgs {
			if addon.Name == endAddon.Name {
				mk8sStatus.Addons[i].Arguments = endAddon.Args
				break
			}
		}
	}
	mk8sStatus.CurrentVersion = currentVersion
	mk8sStatus.KubernetesVersions = MicroK8sVersions
	mk8sStatus.RequiredAddons = GetDefaultAddons()

	return mk8sStatus, nil
}

func (service *Microk8sCluster) AddNodes(endpoint *portaineree.Endpoint, credentials *models.CloudCredential, req *Microk8sScalingRequest) error {
	log.Info().Msgf("Adding %d master nodes and %d worker nodes to microk8s cluster", len(req.MasterNodesToAdd), len(req.WorkerNodesToAdd))

	// Get a list of all the existing nodes in the cluster
	kubectl, err := service.KubeClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return fmt.Errorf("failed to get kube client: %w", err)
	}

	existingNodes, err := kubectl.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	masterNode := util.UrlToMasterNode(endpoint.URL)
	log.Debug().Msgf("Current masterNode: %s", masterNode)

	_, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().
			Str("provider", "microk8s").
			Msg("credentials are missing ssh username")
		return fmt.Errorf("missing ssh username")
	}
	_, passphraseOK := credentials.Credentials["passphrase"]
	_, privateKeyOK := credentials.Credentials["privateKey"]

	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
	}

	// GetKubernetesVersion from the master node
	version, err := GetCurrentVersion(service.DataStore, credentials, endpoint.ID)
	if err != nil {
		return err
	}

	var g errgroup.Group

	// The first step is to install microk8s on all nodes concurrently.
	service.SetMessage("Scaling cluster", "Installing MicroK8s on each node", portaineree.EndpointOperationStatusProcessing)
	nodes := append(req.MasterNodesToAdd, req.WorkerNodesToAdd...)

	for _, node := range nodes {
		func(credentials *models.CloudCredential, ip string) {
			g.Go(func() error {
				return InstallMicrok8sOnNode(credentials, ip, version)
			})
		}(credentials, node)
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	log.Debug().Msgf("Creating host entries on nodes")
	service.SetMessage("Scaling cluster", "Adding host entries to all nodes", portaineree.EndpointOperationStatusProcessing)

	allNodes := append(util.NodeListToIpList(existingNodes), nodes...)
	err = SetupHostEntries(credentials, allNodes)
	if err != nil {
		return fmt.Errorf("error setting up host entries: %w", err)
	}

	sshClient, err := sshUtil.NewConnectionWithCredentials(masterNode, credentials)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	for i := 0; i < len(nodes); i++ {
		log.Info().Msgf("Joining nodes to cluster")

		service.SetMessage("Scaling cluster", "Adding nodes to the cluster", portaineree.EndpointOperationStatusProcessing)
		token, err := RetrieveClusterJoinInformation(sshClient)
		if err != nil {
			return fmt.Errorf("failed to get cluster join information %w", err)
		}

		// two lists if ip addresses concatenated. If the index is greater than the length of the
		// master node list-1, then this node is part of the worker node list
		isWorkerNode := i > len(req.MasterNodesToAdd)-1
		err = ExecuteJoinClusterCommandOnNode(credentials, nodes[i], token, isWorkerNode)
		if err != nil {
			return fmt.Errorf("failed to join node to cluster. %w", err)
		}
	}

	service.SetMessage("Scaling cluster", "Enabling addons", portaineree.EndpointOperationStatusProcessing)
	addons := make(AddonsWithArgs, 0, len(endpoint.CloudProvider.AddonsWithArgs))
	for _, addon := range endpoint.CloudProvider.AddonsWithArgs {
		addons = append(addons, addon)
	}
	// Activate addons on relevant nodes.
	if ok {
		addons.EnableAddons(
			req.MasterNodesToAdd,
			req.WorkerNodesToAdd,
			credentials,
			service.SetMessage,
		)
	}

	return nil
}

func (service *Microk8sCluster) RemoveNodes(endpoint *portaineree.Endpoint, credentials *models.CloudCredential, req *Microk8sScalingRequest) error {
	log.Info().Msgf("Removing %d nodes from the microk8s cluster", len(req.WorkerNodesToAdd))

	masterNode := util.UrlToMasterNode(endpoint.URL)
	log.Debug().Msgf("Current masterNode: %s", masterNode)

	// Prevent removing the master node, which is the node we use to talk to the cluster
	// TODO: should we just continue to remove the other nodes and ignore this one?
	if slices.Contains(req.WorkerNodesToAdd, masterNode) {
		return fmt.Errorf("cannot remove master node %s", masterNode)
	}

	_, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().
			Str("provider", "microk8s").
			Msg("credentials are missing ssh username")
		return fmt.Errorf("missing ssh username")
	}
	_, passphraseOK := credentials.Credentials["passphrase"]
	_, privateKeyOK := credentials.Credentials["privateKey"]

	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
	}

	var nodesNotRemoved []string

	for i := 0; i < len(req.NodesToRemove); i++ {
		log.Debug().Msgf("Removing nodes from the cluster")

		// first get the hostname
		hostname, err := RetrieveHostname(credentials, req.NodesToRemove[i])
		if err != nil {
			msg := fmt.Sprintf("failed to retrieve hostname from node %s. Remove node skipped: %v", req.NodesToRemove[i], err)
			log.Error().Err(err).Msg(msg)
			nodesNotRemoved = append(nodesNotRemoved, req.NodesToRemove[i])
			continue
		}

		err = ExecuteAnnotateNodeCommandOnNode(credentials, masterNode, hostname)
		if err != nil {
			log.Error().Err(err).Msgf("failed to annotate node %s, %v. Continuing to remove node", req.NodesToRemove[i], err)
		}

		err = ExecuteDrainNodeCommandOnNode(credentials, masterNode, hostname)
		if err != nil {
			log.Error().Err(err).Msgf("failed to drain node %s, %v. Continuing to remove node", req.NodesToRemove[i], err)
		}

		force := false
		err = ExecuteLeaveClusterCommandOnNode(credentials, req.NodesToRemove[i])
		if err != nil {
			force = true
		}

		// Sometimes we fail to remove the node.
		// If force is false, try again with force set to true.
		// If force is already true return the error
		for {
			err = ExecuteRemoveNodeCommandOnNode(credentials, masterNode, hostname, force)
			if err == nil {
				break
			}

			if force {
				log.Error().Err(err).Msgf("failed to remove node %s from cluster", req.NodesToRemove[i])
				break
			}
			force = true
		}

		if err != nil {
			msg := fmt.Sprintf("failed to remove node %s from cluster", req.NodesToRemove[i])
			log.Error().Err(err).Msg(msg)
			nodesNotRemoved = append(nodesNotRemoved, req.NodesToRemove[i])
			continue
		}
	}

	if len(nodesNotRemoved) > 0 {
		return fmt.Errorf("failed to remove these nodes from the cluster (%s) See log for details", strings.Join(nodesNotRemoved, ","))
	}

	return nil
}

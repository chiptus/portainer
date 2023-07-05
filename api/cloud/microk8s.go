package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	portaineree "github.com/portainer/portainer-ee/api"
	mk8s "github.com/portainer/portainer-ee/api/cloud/microk8s"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type (
	Microk8sScalingRequest struct {
		EndpointID portaineree.EndpointID `json:"EndpointID"`

		// scaling up
		MasterNodesToAdd []string `json:"MasterNodesToAdd,omitempty"`
		WorkerNodesToAdd []string `json:"WorkerNodesToAdd,omitempty"`

		// scaling down or removing nodes
		NodesToRemove []string `json:"NodesToRemove,omitempty"`
	}

	Microk8sUpgradeRequest struct {
		EndpointID portaineree.EndpointID `json:"EndpointID"`
	}

	Microk8sUpdateAddonsRequest struct {
		EndpointID portaineree.EndpointID `json:"EndpointID"`

		Addons []string `json:"Addons,omitempty"`
	}
)

func (r *Microk8sUpgradeRequest) Provider() string {
	return portaineree.CloudProviderMicrok8s
}

func (r *Microk8sScalingRequest) Provider() string {
	return portaineree.CloudProviderMicrok8s
}

func (r *Microk8sScalingRequest) String() string {
	// convert to json
	b, err := json.Marshal(*r)
	if err != nil {
		return ""
	}
	return string(b)
}

func (service *CloudInfoService) MicroK8sGetInfo() mk8s.MicroK8sInfo {
	addonOptions := make([]mk8s.AddonPair, 0)
	for _, s := range mk8s.GetAllAvailableAddons() {
		addonOptions = append(
			addonOptions, mk8s.AddonPair{
				Pair:                 portaineree.Pair{Name: s.Name, Value: s.Name},
				VersionAvailableFrom: s.MicroK8sVersionAvailableFrom,
				Type:                 s.Type,
			},
		)
	}
	return mk8s.MicroK8sInfo{
		KubernetesVersions: mk8s.MicroK8sVersions,
		AvailableAddons:    addonOptions,
		RequiredAddons:     mk8s.GetAllDefaultAddons(),
	}
}

func (service *CloudInfoService) Microk8sGetAddons(credential *models.CloudCredential, environmentID int) (*mk8s.Microk8sStatusResponse, error) {
	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("processing get info request")

	// Gather nodeIP from environmentID
	endpoint, err := service.dataStore.Endpoint().Endpoint(portaineree.EndpointID(environmentID))
	if err != nil {
		log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("failed looking up environment nodeIP")
		return nil, err
	}
	nodeIP, _, _ := strings.Cut(endpoint.URL, ":")

	// Gather current addon list.
	sshClient, err := sshUtil.NewConnection(
		credential.Credentials["username"],
		credential.Credentials["password"],
		credential.Credentials["passphrase"],
		credential.Credentials["privateKey"],
		nodeIP,
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

	currentVersion, err := mk8s.ParseSnapInstalledVersion(respSnapList.String())
	if err != nil {
		return nil, fmt.Errorf("failed to run ssh command: %w", err)
	}

	var resp bytes.Buffer
	err = sshClient.RunCommand("microk8s status --format yaml", &resp)
	if err != nil {
		return nil, err
	}

	mk8sStatus, err := mk8s.ParseAddonResponse(resp.String())
	if err != nil {
		return nil, err
	}
	mk8sStatus.CurrentVersion = currentVersion
	mk8sStatus.KubernetesVersions = mk8s.MicroK8sVersions
	mk8sStatus.RequiredAddons = mk8s.GetAllDefaultAddons()

	return mk8sStatus, nil
}

func (service *CloudManagementService) Microk8sProvisionCluster(req mk8s.Microk8sProvisioningClusterRequest) (string, error) {

	log.Debug().
		Str("provider", "microk8s").
		Int("node_count", len(req.MasterNodes)+len(req.WorkerNodes)).
		Str("kubernetes_version", req.KubernetesVersion).
		Msg("sending KaaS cluster provisioning request")

	// TODO: REVIEW-POC-MICROK8S
	// Technically using a context here would allow a fail fast approach
	// Right now if an error occurs on one node, the other nodes will still be provisioned
	// See: https://cs.opensource.google/go/x/sync/+/7f9b1623:errgroup/errgroup.go;l=66
	var g errgroup.Group

	user, ok := req.Credentials.Credentials["username"]
	if !ok {
		log.Debug().
			Str("provider", "microk8s").
			Msg("credentials are missing ssh username")
		return "", fmt.Errorf("missing ssh username")
	}
	password := req.Credentials.Credentials["password"]

	passphrase, passphraseOK := req.Credentials.Credentials["passphrase"]
	privateKey, privateKeyOK := req.Credentials.Credentials["privateKey"]
	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
		return "", fmt.Errorf("missing private key, but given passphrase")
	}

	setMessage := service.setMessageHandler(req.EnvironmentID, "")

	// The first step is to install microk8s on all nodes concurrently.
	setMessage("Creating MicroK8s cluster", "Installing MicroK8s on each node", "processing")
	nodes := append(req.MasterNodes, req.WorkerNodes...)
	for _, nodeIp := range nodes {
		func(user, password, passphrase, privateKey, ip string) {
			g.Go(func() error {
				return mk8s.InstallMicrok8sOnNode(user, password, passphrase, privateKey, ip, req.KubernetesVersion)
			})
		}(user, password, passphrase, privateKey, nodeIp)
	}

	err := g.Wait()
	if err != nil {
		return "", err
	}

	if len(nodes) > 1 {
		// If we have more than one node, we need them to form a cluster
		// Note that only 3 node topology is supported at the moment (hardcoded)

		// In order for a microk8s "master" node to join/reach out to other nodes (other managers/workers)
		// it needs to be able to resolve the hostnames of the other nodes
		// See: https://github.com/canonical/microk8s/issues/2967
		// Right now, we extract the hostname/IP from all the nodes after the first
		// and we setup the /etc/hosts file on the first node (where the microk8s add-node command will be run)
		// To be determined whether that is an infrastructure requirement and not something that Portainer should orchestrate.
		setMessage("Creating MicroK8s cluster", "Adding host entries to all nodes", "processing")
		err = mk8s.SetupHostEntries(user, password, passphrase, privateKey, nodes)
		if err != nil {
			return "", err
		}

		for i := 1; i < len(nodes); i++ {
			setMessage("Creating MicroK8s cluster", "Joining nodes to the cluster", "processing")
			token, err := mk8s.RetrieveClusterJoinInformation(user, password, passphrase, privateKey, req.MasterNodes[0])
			if err != nil {
				return "", err
			}

			// worker nodes begin at len(req.MasterNodes)
			asWorkerNode := i >= len(req.MasterNodes)
			err = mk8s.ExecuteJoinClusterCommandOnNode(user, password, passphrase, privateKey, nodes[i], token, asWorkerNode)
			if err != nil {
				return "", err
			}
		}
	}

	// We activate addons on the master node
	if len(req.Addons) > 0 {
		setMessage("Creating MicroK8s cluster", "Enabling MicroK8s addons", "processing")
		errCount := 0
		for _, addon := range req.Addons {
			err = mk8s.EnableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, req.MasterNodes[0], addon)
			if err != nil {
				// Rather than fail the whole thing.  Warn the user and allow them to manually try to enable the addon
				log.Warn().AnErr("error", err).Msgf("failed to enable microk8s addon %s on node. error: ", addon)
				errCount++
			}
		}

		if errCount > 0 {
			log.Error().Msgf("failed to enable %d microk8s addons on node.  Please enable these manually", errCount)
		}
	}

	// Microk8s clusters do not have a cloud provider cluster identifier
	// We currently generate a random identifier for these clusters using UUIDv4
	uid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

func urlToMasterNode(url string) string {
	u := strings.Split(url, ":")
	return u[0]
}

func (service *CloudManagementService) processMicrok8sScalingRequest(req *Microk8sScalingRequest) error {
	log.Debug().Msgf("Processing microk8s scaling request for environment %d", req.EndpointID)
	setMessage := service.setMessageHandler(req.EndpointID, "scale")

	setMessage("Scaling cluster", "Scaling in progress", "processing")

	endpoint, err := service.dataStore.Endpoint().Endpoint(req.EndpointID)
	if err != nil {
		return fmt.Errorf("failed to retrieve environment %d. %w", req.EndpointID, err)
	}

	if endpoint.CloudProvider == nil {
		return fmt.Errorf("environment %d was not provisioned from Portainer", req.EndpointID)
	}

	credentials, err := service.dataStore.CloudCredential().Read(endpoint.CloudProvider.CredentialID)
	if err != nil {
		return fmt.Errorf("failed to retrieve credentials for endpoint %d. %w", req.EndpointID, err)
	}

	if len(req.MasterNodesToAdd) > 0 || len(req.WorkerNodesToAdd) > 0 {
		setMessage("Scaling cluster", "Scaling up in progress", "processing")
		err = service.microk8sAddNodes(endpoint, credentials, req)
	} else if len(req.NodesToRemove) > 0 {
		setMessage("Scaling cluster", "Scaling sown in progress", "processing")
		err = service.microk8sRemoveNodes(endpoint, credentials, req)
	}

	if err != nil {
		details := fmt.Sprintf("Scaling error: %v", err)
		setMessage("Scaling cluster", details, "error")
		return err
	}

	setMessage("Scaling up cluster", "Scaling finished", "")
	return nil
}

func (service *CloudManagementService) processMicrok8sUpdateAddonsRequest(req *Microk8sUpdateAddonsRequest) error {
	log.Debug().Msgf("Processing microk8s addons request for environment %d", req.EndpointID)

	endpoint, err := service.dataStore.Endpoint().Endpoint(req.EndpointID)
	if err != nil {
		return fmt.Errorf("failed to retrieve environment %d. %w", req.EndpointID, err)
	}

	if endpoint.CloudProvider == nil {
		return fmt.Errorf("environment %d was not provisioned from Portainer", req.EndpointID)
	}

	credentials, err := service.dataStore.CloudCredential().Read(endpoint.CloudProvider.CredentialID)
	if err != nil {
		return fmt.Errorf("failed to retrieve credentials for endpoint %d. %w", req.EndpointID, err)
	}

	service.Microk8sUpdateAddons(endpoint, credentials, req)

	return nil
}

func (service *CloudManagementService) processMicrok8sUpgradeRequest(req *Microk8sUpgradeRequest) error {
	log.Debug().Msgf("Processing microk8s scaling request for environment %d", req.EndpointID)

	endpoint, err := service.dataStore.Endpoint().Endpoint(req.EndpointID)
	if err != nil {
		return fmt.Errorf("failed to retrieve environment %d. %w", req.EndpointID, err)
	}

	if endpoint.CloudProvider == nil {
		return fmt.Errorf("environment %d was not provisioned from Portainer", req.EndpointID)
	}

	mk8sUpgrade := mk8s.NewMicrok8sUpgrade(endpoint, service.dataStore)
	_, err = mk8sUpgrade.Upgrade()

	return err
}

func (service *CloudManagementService) microk8sAddNodes(endpoint *portaineree.Endpoint, credentials *models.CloudCredential, req *Microk8sScalingRequest) error {
	log.Info().Msgf("Adding %d master nodes and %d worker nodes to microk8s cluster", len(req.MasterNodesToAdd), len(req.WorkerNodesToAdd))

	masterNode := urlToMasterNode(endpoint.URL)
	log.Debug().Msgf("Current masterNode: %s", masterNode)

	user, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().
			Str("provider", "microk8s").
			Msg("credentials are missing ssh username")
		return fmt.Errorf("missing ssh username")
	}
	password := credentials.Credentials["password"]
	passphrase, passphraseOK := credentials.Credentials["passphrase"]
	privateKey, privateKeyOK := credentials.Credentials["privateKey"]

	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
	}

	// GetKubernetesVersion from the master node
	s := CloudInfoService{
		dataStore:   service.dataStore,
		shutdownCtx: service.shutdownCtx,
	}

	version, err := s.Microk8sVersion(credentials, int(endpoint.ID))
	if err != nil {
		return err
	}

	var g errgroup.Group

	setMessage := service.setMessageHandler(req.EndpointID, "scale")

	// The first step is to install microk8s on all nodes concurrently.
	setMessage("Scaling cluster", "Installing MicroK8s on each node", "processing")
	nodes := append(req.MasterNodesToAdd, req.WorkerNodesToAdd...)

	for _, node := range nodes {
		func(user, password, passphrase, privateKey, ip string) {
			g.Go(func() error {
				return mk8s.InstallMicrok8sOnNode(user, password, passphrase, privateKey, ip, version)
			})
		}(user, password, passphrase, privateKey, node)
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	log.Debug().Msgf("Creating host entries on nodes")

	setMessage("Scaling cluster", "Adding host entries to all nodes", "processing")
	err = mk8s.SetupHostEntries(user, password, passphrase, privateKey, nodes)
	if err != nil {
		return fmt.Errorf("failed to add host entries to node. %w", err)
	}

	for i := 0; i < len(nodes); i++ {
		log.Info().Msgf("Joining nodes to cluster")

		setMessage("Scaling cluster", "Adding nodes to the cluster", "processing")
		token, err := mk8s.RetrieveClusterJoinInformation(user, password, passphrase, privateKey, masterNode)
		if err != nil {
			return fmt.Errorf("failed to get cluster join information %w", err)
		}

		// two lists if ip addresses concatenated. If the index is greater than the length of the
		// master node list-1, then this node is part of the worker node list
		isWorkerNode := i > len(req.MasterNodesToAdd)-1
		err = mk8s.ExecuteJoinClusterCommandOnNode(user, password, passphrase, privateKey, nodes[i], token, isWorkerNode)
		if err != nil {
			return fmt.Errorf("failed to join node to cluster. %w", err)
		}
	}

	return nil
}

func (service *CloudManagementService) Microk8sUpdateAddons(endpoint *portaineree.Endpoint, credentials *models.CloudCredential, req *Microk8sUpdateAddonsRequest) error {
	log.Debug().Str("provider", "microk8s").Msg("Updating microk8s addons")

	user, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().Str("provider", "microk8s").Msg("credentials are missing ssh username")
		return fmt.Errorf("missing ssh username")
	}
	password := credentials.Credentials["password"]
	passphrase, passphraseOK := credentials.Credentials["passphrase"]
	privateKey, privateKeyOK := credentials.Credentials["privateKey"]
	if passphraseOK && !privateKeyOK {
		log.Debug().Str("provider", "microk8s").Msg("passphrase provided, but we are missing a private key")
		return fmt.Errorf("missing private key, but given passphrase")
	}

	nodeIPs, err := service.Microk8sGetNodeIPs(credentials, int(endpoint.ID))
	if err != nil {
		return fmt.Errorf("failed to get existing cluster ips: %w", err)
	}

	log.Debug().Msgf("Microk8s NodeIPs: %v", nodeIPs)

	masterNode := urlToMasterNode(endpoint.URL)

	log.Debug().Msgf("Master node: %s", masterNode)

	payload := types.Microk8sAddonsPayload{
		Addons: req.Addons,
	}

	setMessage := service.setMessageHandler(req.EndpointID, "addons")

	if len(req.Addons) > 0 {
		nodeIP, _, _ := strings.Cut(endpoint.URL, ":")
		if err != nil {
			return fmt.Errorf("unable to get portainer agent address: %w", err)
		}

		setMessage("Updating addons", "Enabling/Disabling MicroK8s addons", "processing")
		microK8sInfo, err := service.Microk8sGetAddons(endpoint.ID, credentials)
		if err != nil {
			return err
		}

		allInstallableAddons := types.Microk8sAddonsPayload{
			Addons: mk8s.GetAllAvailableAddons().GetNames(),
		}

		deletedAddons := []string{}
		newAddons := []string{}
		for _, addon := range microK8sInfo.Addons {
			if allInstallableAddons.IndexOf(addon.Name) != -1 {
				log.Info().Msgf("Addon %s Status %s", addon.Name, addon.Status)
				if addon.Status == "enabled" && payload.IndexOf(addon.Name) == -1 {
					deletedAddons = append(deletedAddons, addon.Name)
				}
				if addon.Status == "disabled" && payload.IndexOf(addon.Name) != -1 {
					newAddons = append(newAddons, addon.Name)
				}
			}
		}

		log.Info().Msgf("New addons: %v", newAddons)
		log.Info().Msgf("Delete addons: %v", deletedAddons)

		log.Debug().Msgf("Enabling addons")

		setMessage("Updating addons", "Enabling MicroK8s addons", "processing")
		errCount := 0
		for _, addon := range newAddons {
			err = mk8s.EnableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, nodeIP, addon)
			if err != nil {
				// Rather than fail the whole thing.  Warn the user and allow them to manually try to enable the addon
				log.Warn().AnErr("error", err).Msgf("failed to enable microk8s addon %s on node. error: ", addon)
				errCount++
			}
		}

		if errCount > 0 {
			log.Error().Msgf("failed to enable %d microk8s addons on node.  Please enable these manually", errCount)
		}

		log.Debug().Msgf("Disabling addons")
		setMessage("Updating addons", "Disabling MicroK8s addons", "processing")
		errCount = 0
		for _, addon := range deletedAddons {
			err = mk8s.DisableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, nodeIP, addon)
			if err != nil {
				// Rather than fail the whole thing.  Warn the user and allow them to manually try to disable the addon
				log.Warn().AnErr("error", err).Msgf("failed to disable microk8s addon %s on node. error: ", addon)
				errCount++
			}
		}

		if errCount > 0 {
			log.Error().Msgf("failed to disable %d microk8s addons on node.  Please disable these manually", errCount)
		}
	}

	setMessage("Updating addons", "Addons updated", "")

	return nil
}

// Microk8sGetCluster simply connects to the first node IP and retrieves the cluster information (kubeconfig)
func (service *CloudManagementService) Microk8sGetCluster(user, password, passphrase, privateKey, clusterID string, clusterip string) (*KaasCluster, error) {
	log.Debug().
		Str("provider", "microk8s").
		Str("cluster_id", clusterID).
		Str("cluster ip", fmt.Sprintf("%v", clusterip)).
		Msg("sending KaaS cluster details request")

	sshClient, err := sshUtil.NewConnection(user, password, passphrase, privateKey, clusterip)
	if err != nil {
		return nil, err
	}
	defer sshClient.Close()

	var kubeconfig bytes.Buffer
	err = sshClient.RunCommand("microk8s config", &kubeconfig)
	if err != nil {
		return nil, err
	}

	return &KaasCluster{
		Id:         clusterID,
		Name:       "",
		Ready:      true,
		KubeConfig: kubeconfig.String(),
	}, nil
}

func (service *CloudInfoService) Microk8sVersion(credential *models.CloudCredential, environmentID int) (string, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msg("processing version request")

	// Gather nodeIP from environmentID.
	endpoint, err := service.dataStore.Endpoint().Endpoint(portaineree.EndpointID(environmentID))
	if err != nil {
		log.Debug().Str(
			"provider",
			portaineree.CloudProviderMicrok8s,
		).Msg("failed looking up environment nodeIP")
		return "", err
	}
	nodeIP, _, _ := strings.Cut(endpoint.URL, ":")

	// Gather current version.
	// We need to ssh into the server to fetch this live. Even if we stored the
	// version in the database, it could be outdated as the user can always
	// update their cluster manually outside of portainer.
	sshClient, err := sshUtil.NewConnection(
		credential.Credentials["username"],
		credential.Credentials["password"],
		credential.Credentials["passphrase"],
		credential.Credentials["privateKey"],
		nodeIP,
	)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating ssh credentials")
		return "", err
	}
	defer sshClient.Close()

	// We can't use the microk8s version command as it was added in 1.25.
	// Instead we parse the output from snap.
	var resp bytes.Buffer
	err = sshClient.RunCommand(
		"snap list",
		&resp,
	)
	if err != nil {
		return "", fmt.Errorf("failed to run ssh command: %w", err)
	}
	return mk8s.ParseSnapInstalledVersion(resp.String())
}

func (service *CloudManagementService) GetSSHConnection(environmentID portaineree.EndpointID, credential *models.CloudCredential) (*sshUtil.SSHConnection, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msg("processing upgrade request")

	// Gather nodeIP from environmentID.
	endpoint, err := service.dataStore.Endpoint().Endpoint(environmentID)
	if err != nil {
		log.Debug().Str(
			"provider",
			portaineree.CloudProviderMicrok8s,
		).Msg("failed looking up environment nodeIP")
		return nil, err
	}
	nodeIP, _, _ := strings.Cut(endpoint.URL, ":")

	// Gather current version.
	// We need to ssh into the server to fetch this live. Even if we stored the
	// version in the database, it could be outdated as the user can always
	// update their cluster manually outside of portainer.
	sshClient, err := sshUtil.NewConnection(
		credential.Credentials["username"],
		credential.Credentials["password"],
		credential.Credentials["passphrase"],
		credential.Credentials["privateKey"],
		nodeIP,
	)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating ssh client")
	}

	return sshClient, nil
}

func (service *CloudManagementService) Microk8sGetAddons(environmentID portaineree.EndpointID, credential *models.CloudCredential) (*mk8s.Microk8sStatusResponse, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msg("processing upgrade request")

	conn, err := service.GetSSHConnection(portaineree.EndpointID(environmentID), credential)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var resp bytes.Buffer
	err = conn.RunCommand("microk8s status --format yaml", &resp)
	if err != nil {
		return nil, err
	}

	return mk8s.ParseAddonResponse(resp.String())
}

func (service *CloudManagementService) Microk8sGetNodeIPs(credential *models.CloudCredential, environmentID int) ([]string, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msg("processing upgrade request")

	conn, err := service.GetSSHConnection(portaineree.EndpointID(environmentID), credential)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// We can't use the microk8s version command as it was added in 1.25.
	// Instead we parse the output from snap.
	var resp bytes.Buffer
	err = conn.RunCommand(
		"microk8s kubectl get nodes -o json",
		&resp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run ssh command: %w", err)
	}
	return mk8s.ParseKubernetesNodesResponse(resp.String())
}

func (service *CloudManagementService) microk8sRemoveNodes(endpoint *portaineree.Endpoint, credentials *models.CloudCredential, req *Microk8sScalingRequest) error {
	log.Info().Msgf("Removing %d nodes from the microk8s cluster", len(req.WorkerNodesToAdd))

	masterNode := urlToMasterNode(endpoint.URL)
	log.Debug().Msgf("Current masterNode: %s", masterNode)

	// Prevent removing the master node, which is the node we use to talk to the cluster
	// TODO: should we just continue to remeove the other nodes and ignore this one?
	if slices.Contains(req.WorkerNodesToAdd, masterNode) {
		return fmt.Errorf("cannot remove master node %s", masterNode)
	}

	user, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().
			Str("provider", "microk8s").
			Msg("credentials are missing ssh username")
		return fmt.Errorf("missing ssh username")
	}
	password := credentials.Credentials["password"]
	passphrase, passphraseOK := credentials.Credentials["passphrase"]
	privateKey, privateKeyOK := credentials.Credentials["privateKey"]

	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
	}

	var nodesNotRemoved []string

	for i := 0; i < len(req.NodesToRemove); i++ {
		log.Debug().Msgf("Removing nodes from the cluster")

		// first get the hostname
		hostname, err := mk8s.RetrieveHostname(user, password, passphrase, privateKey, req.NodesToRemove[i])
		if err != nil {
			msg := fmt.Sprintf("failed to retrieve hostname from node %s. Remove node skipped: %v", req.NodesToRemove[i], err)
			log.Error().Err(err).Msg(msg)
			nodesNotRemoved = append(nodesNotRemoved, req.NodesToRemove[i])
			continue
		}

		err = mk8s.ExecuteAnnotateNodeCommandOnNode(user, password, passphrase, privateKey, masterNode, hostname)
		if err != nil {
			log.Error().Err(err).Msgf("failed to annotate node %s, %v. Continuing to remove node", req.NodesToRemove[i], err)
		}

		err = mk8s.ExecuteDrainNodeCommandOnNode(user, password, passphrase, privateKey, masterNode, hostname)
		if err != nil {
			log.Error().Err(err).Msgf("failed to drain node %s, %v. Continuing to remove node", req.NodesToRemove[i], err)
		}

		force := false
		err = mk8s.ExecuteLeaveClusterCommandOnNode(user, password, passphrase, privateKey, req.NodesToRemove[i])
		if err != nil {
			force = true
		}

		// Sometimes we fail to remove the node.
		// If force is false, try again with force set to true.
		// If force is already true return the error
		for {
			err = mk8s.ExecuteRemoveNodeCommandOnNode(user, password, passphrase, privateKey, masterNode, hostname, force)
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

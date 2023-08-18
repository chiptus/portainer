package cloud

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	portaineree "github.com/portainer/portainer-ee/api"
	mk8s "github.com/portainer/portainer-ee/api/cloud/microk8s"
	"github.com/portainer/portainer-ee/api/cloud/util"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type (
	Microk8sUpgradeRequest struct {
		EndpointID portaineree.EndpointID `json:"EndpointID"`
	}

	Microk8sUpdateAddonsRequest struct {
		EndpointID portaineree.EndpointID `json:"EndpointID"`

		Addons []portaineree.MicroK8sAddon `json:"addons,omitempty"`
	}
)

func (r *Microk8sUpgradeRequest) Provider() string {
	return portaineree.CloudProviderMicrok8s
}

func (service *CloudInfoService) MicroK8sGetInfo() mk8s.MicroK8sInfo {
	return mk8s.MicroK8sInfo{
		KubernetesVersions: mk8s.MicroK8sVersions,
		AvailableAddons:    mk8s.GetAllAvailableAddons(),
		RequiredAddons:     mk8s.GetDefaultAddons(),
	}
}

// Microk8sGetNodeStatus returns the status of a microk8s node
func (service *CloudInfoService) Microk8sGetStatus(credential *models.CloudCredential, environmentID int, nodeIP string) (string, error) {
	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("processing get info request")

	// Gather current addon list.
	sshClient, err := sshUtil.NewConnectionWithCredentials(
		nodeIP,
		credential,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed creating ssh client")
		return "", err
	}
	defer sshClient.Close()

	var respSnapList bytes.Buffer
	if err = sshClient.RunCommand(
		"snap list",
		&respSnapList,
	); err != nil {
		return "", fmt.Errorf("failed to run ssh command: %w", err)
	}

	var resp bytes.Buffer
	err = sshClient.RunCommand("microk8s status", &resp)
	return resp.String(), err
}

func (service *CloudInfoService) Microk8sGetAddons(credential *models.CloudCredential, environmentID int) (*mk8s.Microk8sStatusResponse, error) {
	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("processing get info request")

	nodeObj := mk8s.Microk8sCluster{DataStore: service.dataStore}
	return nodeObj.GetAddons(credential, environmentID)
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

	_, ok := req.Credentials.Credentials["username"]
	if !ok {
		log.Debug().
			Str("provider", "microk8s").
			Msg("credentials are missing ssh username")
		return "", fmt.Errorf("missing ssh username")
	}
	_, passphraseOK := req.Credentials.Credentials["passphrase"]
	_, privateKeyOK := req.Credentials.Credentials["privateKey"]
	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
		return "", fmt.Errorf("missing private key, but given passphrase")
	}

	setMessage := service.setMessageHandler(req.EnvironmentID, "")

	// The first step is to install microk8s on all nodes concurrently.
	setMessage("Creating MicroK8s cluster", "Installing MicroK8s on each node", portaineree.EndpointOperationStatusProcessing)
	nodes := append(req.MasterNodes, req.WorkerNodes...)
	for _, nodeIp := range nodes {
		func(credentials *models.CloudCredential, ip string) {
			g.Go(func() error {
				return mk8s.InstallMicrok8sOnNode(credentials, ip, req.KubernetesVersion)
			})
		}(req.Credentials, nodeIp)
	}

	err := g.Wait()
	if err != nil {
		return "", err
	}

	sshClient, err := sshUtil.NewConnectionWithCredentials(req.MasterNodes[0], req.Credentials)
	if err != nil {
		return "", err
	}
	defer sshClient.Close()

	if len(nodes) > 1 {
		// If we have more than one node, we need them to form a cluster
		// Note that only 3 node topology is supported at the moment (hardcoded)

		// In order for a microk8s "master" node to join/reach out to other nodes (other managers/workers)
		// it needs to be able to resolve the hostnames of the other nodes
		// See: https://github.com/canonical/microk8s/issues/2967
		// Right now, we extract the hostname/IP from all the nodes after the first
		// and we setup the /etc/hosts file on the first node (where the microk8s add-node command will be run)
		// To be determined whether that is an infrastructure requirement and not something that Portainer should orchestrate.
		setMessage("Creating MicroK8s cluster", "Adding host entries to all nodes", portaineree.EndpointOperationStatusProcessing)
		err = mk8s.SetupHostEntries(req.Credentials, nodes)
		if err != nil {
			return "", err
		}

		for i := 1; i < len(nodes); i++ {
			setMessage("Creating MicroK8s cluster", "Joining nodes to the cluster", portaineree.EndpointOperationStatusProcessing)
			token, err := mk8s.RetrieveClusterJoinInformation(sshClient)
			if err != nil {
				return "", err
			}

			// worker nodes begin at len(req.MasterNodes)
			asWorkerNode := i >= len(req.MasterNodes)
			err = mk8s.ExecuteJoinClusterCommandOnNode(req.Credentials, nodes[i], token, asWorkerNode)
			if err != nil {
				return "", err
			}
		}
	}

	setMessageAddons := service.setMessageHandler(req.EnvironmentID, "addons")
	setMessageAddons("Creating MicroK8s cluster", "Enabling addons", portaineree.EndpointOperationStatusProcessing)
	// Activate addons on relevant nodes.
	req.Addons.EnableAddons(
		req.MasterNodes,
		req.WorkerNodes,
		req.Credentials,
		setMessageAddons,
	)

	// Microk8s clusters do not have a cloud provider cluster identifier
	// We currently generate a random identifier for these clusters using UUIDv4
	uid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

func (service *CloudManagementService) processMicrok8sUpdateAddonsRequest(req *Microk8sUpdateAddonsRequest) error {
	log.Info().Msgf("Processing microk8s addons request for environment %d", req.EndpointID)

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

	err, warningSummary := service.Microk8sUpdateAddons(endpoint, credentials, req)

	setMessage := service.setMessageHandler(req.EndpointID, "addons")
	if err != nil {
		setMessage("Failed to update addons", err.Error(), "warning")
		return nil
	}

	if warningSummary != "" {
		setMessage("Addons updated with errors", warningSummary, "warning")
	}

	return nil
}

func (service *CloudManagementService) processMicrok8sUpgradeRequest(req *Microk8sUpgradeRequest) error {
	log.Debug().Msgf("Processing microk8s upgrade request for environment %d", req.EndpointID)

	endpoint, err := service.dataStore.Endpoint().Endpoint(req.EndpointID)
	if err != nil {
		return fmt.Errorf("failed to retrieve environment %d. %w", req.EndpointID, err)
	}

	if endpoint.CloudProvider == nil {
		return fmt.Errorf("environment %d was not provisioned by Portainer", req.EndpointID)
	}

	setMessage := service.setMessageHandler(req.EndpointID, "upgrade")
	mk8sUpgrade := mk8s.NewMicrok8sUpgrade(endpoint, service.dataStore)
	_, err = mk8sUpgrade.Upgrade()
	if err != nil {
		log.Error().Int("endpoint_id", int(endpoint.ID)).Err(err).Msg("failed to upgrade microk8s cluster")
		setMessage("Upgrading cluster", err.Error(), "warning")
	}

	return err
}

func (service *CloudManagementService) Microk8sUpdateAddons(endpoint *portaineree.Endpoint, credentials *models.CloudCredential, req *Microk8sUpdateAddonsRequest) (error, string) {
	log.Debug().Str("provider", "microk8s").Msg("Updating microk8s addons")

	_, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().Str("provider", "microk8s").Msg("credentials are missing ssh username")
		return fmt.Errorf("Missing SSH username"), ""
	}
	_, passphraseOK := credentials.Credentials["passphrase"]
	_, privateKeyOK := credentials.Credentials["privateKey"]
	if passphraseOK && !privateKeyOK {
		log.Debug().Str("provider", "microk8s").Msg("passphrase provided, but we are missing a private key")
		return fmt.Errorf("Missing private key, but given passphrase"), ""
	}

	nodeIPs, err := service.Microk8sGetNodeIPs(credentials, int(endpoint.ID))
	if err != nil {
		return fmt.Errorf("Failed to get existing cluster IPs: %w", err), ""
	}
	log.Debug().Msgf("Microk8s NodeIPs: %v", nodeIPs)

	masterNode := util.UrlToMasterNode(endpoint.URL)
	log.Debug().Msgf("Master node: %s", masterNode)

	payload := types.Microk8sAddonsPayload{
		Addons: req.Addons,
	}

	setMessage := service.setMessageHandler(req.EndpointID, "addons")
	// defer just in case status message is not updated correctly
	defer setMessage("Updating addons", "Addons updated", "")

	setMessage("Updating addons", "Enabling/Disabling MicroK8s addons", portaineree.EndpointOperationStatusProcessing)
	microK8sInfo, err := service.Microk8sGetAddons(endpoint.ID, credentials)
	if err != nil {
		log.Error().Msgf("Failed to get MicroK8s addons: %v", err)
		return fmt.Errorf("Failed to get MicroK8s addons: %w", err), ""
	}

	allInstallableAddons := mk8s.GetAllAvailableAddons()
	endpointAddons := make(mk8s.AddonsWithArgs, 0, len(endpoint.CloudProvider.AddonsWithArgs))
	for _, addon := range endpoint.CloudProvider.AddonsWithArgs {
		endpointAddons = append(endpointAddons, addon)
	}

	deletedAddons := mk8s.AddonsWithArgs{}
	newAddons := mk8s.AddonsWithArgs{}
	for _, addon := range microK8sInfo.Addons {
		if allInstallableAddons.IndexOf(addon.Name) != -1 {
			log.Info().Msgf("Addon %s Status %s", addon.Name, addon.Status)
			index := payload.IndexOf(addon.Name)

			if index == -1 {
				if addon.Status == "enabled" {
					deletedAddons = append(deletedAddons, portaineree.MicroK8sAddon{Name: addon.Name})
				}
			} else {
				if addon.Status == "disabled" {
					newAddons = append(newAddons, payload.Addons[index])
				} else {
					// If existing arguments mismatch add addon to the deletedAddons and newAddons both
					exitingArgument := func() string {
						for _, addonWithArgs := range endpointAddons {
							if addonWithArgs.Name == addon.Name {
								return addonWithArgs.Args
							}
						}
						return ""
					}()
					if exitingArgument != payload.Addons[index].Args {
						deletedAddons = append(deletedAddons, payload.Addons[index])
						newAddons = append(newAddons, payload.Addons[index])
					}
				}
			}
		}
	}

	log.Info().Msgf("New addons requested: %v", newAddons)
	log.Info().Msgf("Delete addons requested: %v", deletedAddons)

	sshClient, err := sshUtil.NewConnectionWithCredentials(masterNode, credentials)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create SSH connection for node %s", masterNode)
		return fmt.Errorf("Failed to create SSH connection for node %s", masterNode), ""
	}
	defer sshClient.Close()

	nodes, err := mk8s.GetAllNodes(sshClient)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get all the nodes from node %s", masterNode)
		return fmt.Errorf("Failed to get all the nodes from node %s", masterNode), ""
	}

	masterNodes := make([]string, 0)
	workerNodes := make([]string, 0)
	for _, n := range nodes {
		if n.IsMaster {
			masterNodes = append(masterNodes, n.IP)
		} else {
			workerNodes = append(workerNodes, n.IP)
		}
	}

	setMessage("Updating addons", "Disabling addons", portaineree.EndpointOperationStatusProcessing)
	disableWarningAddons := deletedAddons.DisableAddons(masterNodes, workerNodes, credentials, setMessage)
	setMessage("Updating addons", "Enabling addons", portaineree.EndpointOperationStatusProcessing)
	enableWarningAddons := newAddons.EnableAddons(masterNodes, workerNodes, credentials, setMessage)

	// Read endpoint again for fresh-copy
	endpoint, err = service.dataStore.Endpoint().Endpoint(req.EndpointID)
	if err != nil {
		log.Error().Msgf("failed to read endpoint (%s - %d)", endpoint.Name, endpoint.ID)
	}

	// Update the endpoint with the new addons
	endpoint.CloudProvider.AddonsWithArgs = req.Addons
	endpoint.StatusMessage.OperationStatus = ""

	err = service.dataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		log.Error().Msgf("failed to update endpoint (%s - %d) addons", endpoint.Name, endpoint.ID)
	}

	var disableWarningSummary, enableWarningSummary string
	if len(disableWarningAddons) > 0 {
		disableWarningAddonNames := disableWarningAddons.GetNames()
		log.Error().Msgf("error while disabling microk8s addons.  Please disable the following addons manually: %s", strings.Join(disableWarningAddonNames[:], ", "))
		disableWarningSummary = fmt.Sprintf("Errors found while disabling MicroK8s addon%s: %s", func() string {
			if len(disableWarningAddonNames) > 1 {
				return "s"
			}
			return ""
		}(), strings.Join(disableWarningAddonNames[:], ", "))
	}

	if len(enableWarningAddons) > 0 {
		enableWarningAddonNames := enableWarningAddons.GetNames()
		log.Error().Msgf("errors found while enabling microk8s addons.  Please enable the following addons manually: %s", strings.Join(enableWarningAddonNames[:], ", "))
		enableWarningSummary = fmt.Sprintf("Errors found while enabling MicroK8s addon%s: %s", func() string {
			if len(enableWarningAddonNames) > 1 {
				return "s"
			}
			return ""
		}(), strings.Join(enableWarningAddonNames[:], ", "))
	}

	// build a warningSummary with enableWarningSummary, disableWarningSummary separated by '<br/>'
	var warningSummary string
	separator := func() string {
		if enableWarningSummary != "" && disableWarningSummary != "" {
			return "<br/>"
		}
		return ""
	}()
	warningSummary = strings.Join([]string{enableWarningSummary, disableWarningSummary}, separator)
	if warningSummary != "" {
		return nil, warningSummary
	}

	return nil, ""
}

// Microk8sGetCluster simply connects to the first node IP and retrieves the cluster information (kubeconfig)
func (service *CloudManagementService) Microk8sGetCluster(credentials *models.CloudCredential, clusterID, clusterip string) (*KaasCluster, error) {
	log.Debug().
		Str("provider", "microk8s").
		Str("cluster_id", clusterID).
		Str("cluster ip", fmt.Sprintf("%v", clusterip)).
		Msg("sending KaaS cluster details request")

	sshClient, err := sshUtil.NewConnectionWithCredentials(clusterip, credentials)
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

	return mk8s.GetCurrentVersion(service.dataStore, credential, portaineree.EndpointID(environmentID))
}

func (service *CloudManagementService) GetSSHConnection(environmentID portaineree.EndpointID, credential *models.CloudCredential) (*sshUtil.SSHConnection, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msgf("Creating ssh connection for environment %d", environmentID)

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
	sshClient, err := sshUtil.NewConnectionWithCredentials(nodeIP, credential)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating SSH client")
		return nil, fmt.Errorf("Failed creating SSH client: %w", err)
	}

	return sshClient, nil
}

func (service *CloudManagementService) Microk8sGetAddons(environmentID portaineree.EndpointID, credential *models.CloudCredential) (*mk8s.Microk8sStatusResponse, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msgf("Getting addons for environment %d", environmentID)

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
	).Msgf("Getting nodes for environment %d", environmentID)

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

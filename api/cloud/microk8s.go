package cloud

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/sftp"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

type (
	MicroK8sInfo struct {
		Addons []string `json:"addons"`
	}

	microk8sClusterJoinInfo struct {
		Token string   `json:"token"`
		URLS  []string `json:"urls"`
	}

	Microk8sProvisioningClusterRequest struct {
		EnvironmentID     portaineree.EndpointID `json:"environmentID"`
		Credentials       *models.CloudCredential
		NodeIps, Addons   []string
		KubernetesVersion string `json:"kubernetesVersion"`
	}
)

func (service *CloudClusterInfoService) Microk8sGetAddons(credential *models.CloudCredential, environmentID int) (interface{}, error) {
	log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("processing get info request")

	// Gather nodeIP from environmentID
	endpoint, err := service.dataStore.Endpoint().Endpoint(portaineree.EndpointID(environmentID))
	if err != nil {
		log.Debug().Str("provider", portaineree.CloudProviderMicrok8s).Msg("failed looking up environment nodeIP")
		return nil, err
	}
	nodeIP, _, _ := strings.Cut(endpoint.URL, ":")

	// Gather current addon list.
	config, err := NewSSHConfig(
		credential.Credentials["username"],
		credential.Credentials["password"],
		credential.Credentials["passphrase"],
		credential.Credentials["privateKey"],
	)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating ssh credentials")
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIP), config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := runSSHCommandAndGetOutput(conn, credential.Credentials["password"], "microk8s status")
	if err != nil {
		return nil, err
	}
	addons := parseAddonResponse(resp)
	return &MicroK8sInfo{
		Addons: addons,
	}, nil
}

func (service *CloudClusterSetupService) Microk8sProvisionCluster(req Microk8sProvisioningClusterRequest) (string, error) {
	log.Debug().
		Str("provider", "microk8s").
		Int("node_count", len(req.NodeIps)).
		Str("kubernetes_version", req.KubernetesVersion).
		Msg("sending KaaS cluster provisioning request")

	// Microk8s clusters do not have a cloud provider cluster identifier
	// We currently generate a random identifier for these clusters using UUIDv4
	uid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

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
	password, _ := req.Credentials.Credentials["password"]

	passphrase, passphraseOK := req.Credentials.Credentials["passphrase"]
	privateKey, privateKeyOK := req.Credentials.Credentials["privateKey"]
	if passphraseOK && !privateKeyOK {
		log.Debug().
			Str("provider", "microk8s").
			Msg("passphrase provided, but we are missing a private key")
		return "", fmt.Errorf("missing private key, but given passphrase")
	}

	// The first step is to install microk8s on all nodes concurrently.
	service.setMessage(req.EnvironmentID, "Creating MicroK8s cluster", "Installing MicroK8s on each node")
	for _, nodeIp := range req.NodeIps {
		func(user, password, passphrase, privateKey, ip string) {
			g.Go(func() error {
				return installMicrok8sOnNode(user, password, passphrase, privateKey, ip, req.KubernetesVersion)
			})
		}(user, password, passphrase, privateKey, nodeIp)
	}

	err = g.Wait()
	if err != nil {
		return "", err
	}

	if len(req.NodeIps) > 1 {
		// If we have more than one node, we need them to form a cluster
		// Note that only 3 node topology is supported at the moment (hardcoded)

		// In order for a microk8s "master" node to join/reach out to other nodes (other managers/workers)
		// it needs to be able to resolve the hostnames of the other nodes
		// See: https://github.com/canonical/microk8s/issues/2967
		// Right now, we extract the hostname/IP from all the nodes after the first
		// and we setup the /etc/hosts file on the first node (where the microk8s add-node command will be run)
		// To be determined whether that is an infrastructure requirement and not something that Portainer should orchestrate.
		service.setMessage(req.EnvironmentID, "Creating MicroK8s cluster", "Adding host entries to all nodes")
		err = setupHostEntries(user, password, passphrase, privateKey, req.NodeIps)
		if err != nil {
			return "", err
		}

		for i := 1; i < len(req.NodeIps); i++ {
			service.setMessage(req.EnvironmentID, "Creating MicroK8s cluster", "Joining nodes to the cluster")
			token, err := retrieveClusterJoinInformation(user, password, passphrase, privateKey, req.NodeIps[0])
			if err != nil {
				return "", err
			}

			// Join nodes to the cluster. The first 1-3 nodes are managers, the rest are workers. For a 2 node cluster, the first node is a manager and the second is a worker.
			asWorkerNode := i >= 3 || (len(req.NodeIps) == 2 && i == 1) // index of 1 is the second node

			err = executeJoinClusterCommandOnNode(user, password, passphrase, privateKey, req.NodeIps[i], token, asWorkerNode)
			if err != nil {
				return "", err
			}
		}
	}

	// We activate addons on the master node
	if len(req.Addons) > 0 {
		service.setMessage(req.EnvironmentID, "Creating MicroK8s cluster", "Enabling MicroK8s addons")
		errCount := 0
		for _, addon := range req.Addons {
			err = enableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, req.NodeIps[0], addon)
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

	return uid.String(), nil
}

// Microk8sGetCluster simply connects to the first node IP and retrieves the cluster information (kubeconfig)
func (service *CloudClusterSetupService) Microk8sGetCluster(user, password, passphrase, privateKey, clusterID string, nodeIps []string) (*KaasCluster, error) {
	log.Debug().
		Str("provider", "microk8s").
		Str("cluster_id", clusterID).
		Msg("sending KaaS cluster details request")

	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return nil, err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIps[0]), config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	kubeconfig, err := runSSHCommandAndGetOutput(conn, password, "microk8s config")
	if err != nil {
		return nil, err
	}

	return &KaasCluster{
		Id:         clusterID,
		Name:       "",
		Ready:      true,
		KubeConfig: kubeconfig,
	}, nil
}

func enableMicrok8sAddonsOnNode(user, password, passphrase, privateKey, nodeIp string, addon string) error {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	command := "microk8s enable " + addon
	return runSSHCommand(conn, password, command)
}

func installMicrok8sOnNode(user, password, passphrase, privateKey, nodeIp, kubernetesVersion string) error {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := 0; i < 3; i++ {
		// Try to install microk8s up to 3 times before we give up.
		cmd := "snap install microk8s --classic --channel=" + kubernetesVersion
		log.Info().Msg("MicroK8s install command on " + nodeIp + ": " + cmd)
		err = runSSHCommand(conn, password, cmd)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	err = runSSHCommand(conn, password, "microk8s status --wait-ready")
	if err != nil {
		return err
	}

	// Default set of addons.
	// Note: helm3 is an alias for helm and recently seems to be enabled by default. Leaving helm in here anyway.
	// ha-cluster is automatically enabled when adding more master nodes to the cluster
	return runSSHCommand(conn, password, "microk8s enable dns rbac helm")
}

func executeJoinClusterCommandOnNode(user, password, passphrase, privateKey, nodeIp string, joinInfo *microk8sClusterJoinInfo, asWorkerNode bool) error {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	workerParam := ""
	if asWorkerNode {
		workerParam = "--worker"
	}

	joinClusterCommand := fmt.Sprintf("microk8s join %s %s", workerParam, joinInfo.URLS[0])
	log.Debug().Msgf("Node with ip %s is joining cluster with command: %s", nodeIp, joinClusterCommand)
	return runSSHCommand(conn, password, joinClusterCommand)
}

func retrieveClusterJoinInformation(user, password, passphrase, privateKey, nodeIp string) (*microk8sClusterJoinInfo, error) {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return nil, err
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	addNodeCommand := "microk8s add-node --format json"
	commandOutput, err := runSSHCommandAndGetOutput(conn, password, addNodeCommand)
	if err != nil {
		return nil, err
	}

	joinInfo := &microk8sClusterJoinInfo{}
	err = json.Unmarshal([]byte(commandOutput), joinInfo)
	if err != nil {
		return nil, err
	}

	return joinInfo, nil
}

func retrieveHostname(user, password, passphrase, privateKey, nodeIp string) (string, error) {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return "", err
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	hostnameCommand := "hostname"
	commandOutput, err := runSSHCommandAndGetOutput(conn, password, hostnameCommand)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(commandOutput, "\n"), nil
}

func updateHostFile(
	user, password, passphrase, privateKey, nodeIp string,
	hostEntries map[string]string,
) error {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	for ip, hostname := range hostEntries {
		if ip == nodeIp {
			continue
		}
		s := fmt.Sprintf("%s %s", ip, hostname)
		command := fmt.Sprintf("sh -c 'echo \"%s\" >> /etc/hosts'", s)
		err = runSSHCommand(conn, password, command)
		if err != nil {
			return err
		}

		// cloud-init workaround
		// On machines created with cloud-init, which is many cloud providers,
		// you need to edit a file in /etc/cloud/templates instead of the main
		// /etc/hosts file in order for your changes to persist on a restart.
		// NOTE: This is a best effort attempt. If the file doesn't exist we
		// skip it and only edit the main hosts file.
		path, err := runSSHCommandAndGetOutput(
			conn,
			password,
			`grep -o "\/etc\/cloud\/templates.*\.tmpl$" /etc/hosts`,
		)
		if err != nil || path == "" || strings.ContainsAny(path, " ") {
			continue
		}
		command = fmt.Sprintf("sh -c 'echo \"%s\" >> %s'", s, path)
		// Not worrying about errors since it will just be the file missing.
		_ = runSSHCommand(conn, password, command)
	}

	return nil
}

func setupHostEntries(user, password, passphrase, privateKey string, nodeIps []string) error {
	// hostEntries is a mapping of nodeIP to hostname.
	hostEntries := make(map[string]string)

	// Build the list of all host entries.
	for _, nodeIp := range nodeIps {
		hostname, err := retrieveHostname(user, password, passphrase, privateKey, nodeIp)
		if err != nil {
			return err
		}

		hostEntries[nodeIp] = hostname
	}

	// Update each of the nodes with the list of host entries.
	for _, nodeIp := range nodeIps {
		err := updateHostFile(user, password, passphrase, privateKey, nodeIp, hostEntries)
		if err != nil {
			return err
		}
	}
	return nil
}

func runSSHCommand(conn *ssh.Client, password, command string) error {
	// Connect to the server.
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	passSFTP, err := sftpClient.Create(".password")
	err = sftpClient.Chmod(".password", 0600)
	if err != nil {
		return err
	}
	_, err = passSFTP.Write([]byte(password))
	if err != nil {
		return err
	}
	passSFTP.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run(fmt.Sprintf("cat '.password' | sudo -S %s", command))
	if err != nil {
		return err
	}
	return sftpClient.Remove(".password")
}

func runSSHCommandAndGetOutput(conn *ssh.Client, password, command string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return "", err
	}
	defer sftpClient.Close()

	passSFTP, err := sftpClient.Create(".password")
	err = sftpClient.Chmod(".password", 0600)
	if err != nil {
		return "", err
	}
	_, err = passSFTP.Write([]byte(password))
	if err != nil {
		return "", err
	}
	passSFTP.Close()

	var buff bytes.Buffer
	session.Stdout = &buff

	session.Stderr = os.Stderr

	err = session.Run(fmt.Sprintf("cat '.password' | sudo -S %s", command))
	if err != nil {
		return "", err
	}
	err = sftpClient.Remove(".password")
	return buff.String(), err
}

func NewSSHConfig(user, password, passphrase, privateKey string) (*ssh.ClientConfig, error) {
	auth := ssh.Password(password)
	if privateKey != "" {
		// Create signer with the private key.
		key, err := base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			log.Err(err).Msg("failed to decode private key")
			return nil, err
		}
		var signer ssh.Signer
		if passphrase == "" {
			signer, err = ssh.ParsePrivateKey(key)
			if err != nil {
				log.Err(err).Msg("failed to parse private key")
				return nil, err
			}
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
			if err != nil {
				log.Err(err).Msg("failed to parse private key")
				return nil, err
			}
		}
		auth = ssh.PublicKeys(signer)
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(5) * time.Second,
	}, nil
}

// parseAddonResponse reads the command line response of `microk8s status` and
// returns a list of installed addons.
func parseAddonResponse(s string) []string {
	var addons []string

	// A regular expressiong to match everything between "enabled:" and
	// "disabled:" which is a list of the enabled addons.
	enabledRegex := regexp.MustCompile(`(?s)enabled:\n(.*).*disabled:`)
	match := enabledRegex.FindStringSubmatch(s)
	if match == nil {
		return addons
	}

	var buf bytes.Buffer
	var comment bool
	// Loop over each line to build a list of enabled addons.
	for _, c := range match[1] {
		switch c {
		case '#':
			// We skip comments by enabling "comment mode".
			comment = true
		case ' ':
			continue
		case '\n':
			addons = append(addons, strings.TrimSpace(buf.String()))
			buf.Reset()
			comment = false
		default:
			if !comment {
				buf.WriteRune(c)
			}
		}
	}
	return addons
}
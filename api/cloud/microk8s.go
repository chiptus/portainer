package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

type microk8sClusterJoinInfo struct {
	Token string   `json:"token"`
	URLS  []string `json:"urls"`
}

func Microk8sProvisionCluster(sshUser, sshPassword string, nodeIps, addons []string) (string, error) {
	log.Debug().
		Str("provider", "microk8s").
		Int("node_count", len(nodeIps)).
		Msg("sending KaaS cluster provisioning request")

	// TODO: REVIEW-POC-MICROK8S
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

	// The first step is to install microk8s on all nodes
	// This is done concurrently
	for _, nodeIp := range nodeIps {
		func(user, password, ip string) {
			g.Go(func() error {
				return installMicrok8sOnNode(user, password, ip)
			})
		}(sshUser, sshPassword, nodeIp)
	}

	err = g.Wait()
	if err != nil {
		return "", err
	}

	if len(nodeIps) > 1 {
		// If we have more than one node, we need them to form a cluster
		// Note that only 3 node topology is supported at the moment (hardcoded)

		// TODO: REVIEW-POC-MICROK8S
		// In order for a microk8s "master" node to join/reach out to other nodes (other managers/workers)
		// it needs to be able to resolve the hostnames of the other nodes
		// See: https://github.com/canonical/microk8s/issues/2967
		// Right now, we extract the hostname/IP from all the nodes after the first
		// and we setup the /etc/hosts file on the first node (where the microk8s add-node command will be run)
		// To be determined whether that is an infrastructure requirement and not something that Portainer should orchestrate.
		err = setupHostEntries(sshUser, sshPassword, nodeIps)
		if err != nil {
			return "", err
		}

		// TODO: REVIEW-POC-MICROK8S
		// The process below can probably be done concurrently
		// It should also support different kind of cluster topology in the future (mix of managers/workers, more than 3 nodes...)

		// Once all nodes are ready, we just pick the first as the "master" where the original microk8s add node command will be executed
		// and we retrieve the first token
		joinInfoNode2, err := retrieveClusterJoinInformation(sshUser, sshPassword, nodeIps[0])
		if err != nil {
			return "", err
		}

		// We join the cluster on node 2
		err = executeJoinClusterCommandOnNode(sshUser, sshPassword, nodeIps[1], joinInfoNode2)
		if err != nil {
			return "", err
		}

		// We retrieve another token
		joinInfoNode3, err := retrieveClusterJoinInformation(sshUser, sshPassword, nodeIps[0])
		if err != nil {
			return "", err
		}

		// We join the cluster on node 3
		err = executeJoinClusterCommandOnNode(sshUser, sshPassword, nodeIps[2], joinInfoNode3)
		if err != nil {
			return "", err
		}

	}

	// We activate addons on the master node
	if len(addons) > 0 {
		err = enableMicrok8sAddonsOnNode(sshUser, sshPassword, nodeIps[0], addons)
		if err != nil {
			return "", err
		}
	}

	return uid.String(), nil
}

// Microk8sGetCluster simply connects to the first node IP and retrieves the cluster information (kubeconfig)
func Microk8sGetCluster(sshUser, sshPassword, clusterID string, nodeIps []string) (*KaasCluster, error) {
	log.Debug().
		Str("provider", "microk8s").
		Str("cluster_id", clusterID).
		Msg("sending KaaS cluster details request")

	// TODO: REVIEW-POC-MICROK8S
	// The use of SSH connection can probably be re-handled in a better way
	// At the moment we basically create a new connection for each process - I believe that the connection can be re-used in some cases
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// Arbitrary connection timeout, might need to be reviewed
		// TODO: REVIEW-POC-MICROK8S
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIps[0]), config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	kubeconfig, err := runSSHCommandAndGetOutput(conn, sshPassword, "microk8s config")
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

func enableMicrok8sAddonsOnNode(sshUser, sshPassword, nodeIp string, addons []string) error {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: REVIEW-POC-MICROK8S
		// Arbitrary connection timeout, might need to be reviewed
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	command := "microk8s enable " + strings.Join(addons, " ")
	return runSSHCommand(conn, sshPassword, command)
}

func installMicrok8sOnNode(sshUser, sshPassword, nodeIp string) error {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: REVIEW-POC-MICROK8S
		// Arbitrary connection timeout, might need to be reviewed
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Here we just SSH on the node and install microk8s
	// We then wait till it is ready to be used

	command1 := "snap install microk8s --classic --channel=1.25"
	err = runSSHCommand(conn, sshPassword, command1)
	if err != nil {
		return err
	}

	command2 := "microk8s status --wait-ready"
	err = runSSHCommand(conn, sshPassword, command2)
	if err != nil {
		return err
	}

	// Temporary - should be selectable in UI with default set to values below (helm helm3 ha-cluster are installed by default with microk8s start)
	command3 := "microk8s enable dns hostpath-storage rbac helm helm3 ha-cluster"
	return runSSHCommand(conn, sshPassword, command3)
}

func executeJoinClusterCommandOnNode(sshUser, sshPassword, nodeIp string, joinInfo *microk8sClusterJoinInfo) error {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: REVIEW-POC-MICROK8S
		// Arbitrary connection timeout, might need to be reviewed
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	joinClusterCommand := fmt.Sprintf("microk8s join %s", joinInfo.URLS[0])
	return runSSHCommand(conn, sshPassword, joinClusterCommand)
}

func retrieveClusterJoinInformation(sshUser, sshPassword, nodeIp string) (*microk8sClusterJoinInfo, error) {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: REVIEW-POC-MICROK8S
		// Arbitrary connection timeout, might need to be reviewed
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	addNodeCommand := "microk8s add-node --format json"
	commandOutput, err := runSSHCommandAndGetOutput(conn, sshPassword, addNodeCommand)
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

func retrieveHostname(sshUser, sshPassword, nodeIp string) (string, error) {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: REVIEW-POC-MICROK8S
		// Arbitrary connection timeout, might need to be reviewed
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	hostnameCommand := "hostname"
	commandOutput, err := runSSHCommandAndGetOutput(conn, sshPassword, hostnameCommand)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(commandOutput, "\n"), nil
}

func updateHostFile(sshUser, sshPassword, nodeIp string, hostEntries []string) error {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		// TODO: REVIEW-POC-MICROK8S
		// This is not recommended for production use
		// Investigate how to use ssh.HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: REVIEW-POC-MICROK8S
		// Arbitrary connection timeout, might need to be reviewed
		Timeout: time.Duration(5) * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", nodeIp), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	// TODO: REVIEW-POC-MICROK8S
	// Right now we append one entry at a time // per SSH command
	// There might be a way to do this in one go
	for _, hostEntry := range hostEntries {
		command := fmt.Sprintf("sh -c 'echo \"%s\" >> /etc/hosts'", hostEntry)
		err = runSSHCommand(conn, sshPassword, command)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupHostEntries(sshUser, sshPassword string, nodeIps []string) error {

	hostEntries := []string{}

	// TODO: REVIEW-POC-MICROK8S
	// Retrieving hostnames on each nodes could be done in parallel

	for idx, nodeIp := range nodeIps {
		if idx == 0 {
			continue
		}

		hostname, err := retrieveHostname(sshUser, sshPassword, nodeIp)
		if err != nil {
			return err
		}

		hostEntry := fmt.Sprintf("%s %s", nodeIp, hostname)
		hostEntries = append(hostEntries, hostEntry)
	}

	return updateHostFile(sshUser, sshPassword, nodeIps[0], hostEntries)
}

func runSSHCommand(conn *ssh.Client, sshPassword, command string) error {
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// TODO: REVIEW-POC-MICROK8S
	// One of the main gotcha of this approach is that the password can be seen in the process list
	// We should investigate if there is a better way to use sudo
	return session.Run(fmt.Sprintf("echo '%s' | sudo -S %s", sshPassword, command))
}

func runSSHCommandAndGetOutput(conn *ssh.Client, sshPassword, command string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var buff bytes.Buffer
	session.Stdout = &buff

	session.Stderr = os.Stderr

	// TODO: REVIEW-POC-MICROK8S
	// One of the main gotcha of this approach is that the password can be seen in the process list
	// We should investigate if there is a better way to use sudo
	err = session.Run(fmt.Sprintf("echo '%s' | sudo -S %s", sshPassword, command))
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

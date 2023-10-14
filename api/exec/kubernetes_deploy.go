package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/proxy/factory"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// KubernetesDeployer represents a service to deploy resources inside a Kubernetes environment(endpoint).
type KubernetesDeployer struct {
	binaryPath                  string
	dataStore                   dataservices.DataStore
	reverseTunnelService        portaineree.ReverseTunnelService
	signatureService            portainer.DigitalSignatureService
	kubernetesClientFactory     *cli.ClientFactory
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
	authService                 *authorization.Service
	proxyManager                *proxy.Manager
}

// NewKubernetesDeployer initializes a new KubernetesDeployer service.
func NewKubernetesDeployer(authService *authorization.Service, kubernetesTokenCacheManager *kubernetes.TokenCacheManager, kubernetesClientFactory *cli.ClientFactory, datastore dataservices.DataStore, reverseTunnelService portaineree.ReverseTunnelService, signatureService portainer.DigitalSignatureService, proxyManager *proxy.Manager, binaryPath string) *KubernetesDeployer {
	return &KubernetesDeployer{
		binaryPath:                  binaryPath,
		dataStore:                   datastore,
		reverseTunnelService:        reverseTunnelService,
		signatureService:            signatureService,
		kubernetesClientFactory:     kubernetesClientFactory,
		kubernetesTokenCacheManager: kubernetesTokenCacheManager,
		authService:                 authService,
		proxyManager:                proxyManager,
	}
}

func (deployer *KubernetesDeployer) getToken(userID portainer.UserID, endpoint *portaineree.Endpoint, setLocalAdminToken bool) (string, error) {
	kubeCLI, err := deployer.kubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return "", err
	}

	tokenCache := deployer.kubernetesTokenCacheManager.GetOrCreateTokenCache(endpoint.ID)

	tokenManager, err := kubernetes.NewTokenManager(kubeCLI, deployer.dataStore, tokenCache, setLocalAdminToken, deployer.authService)
	if err != nil {
		return "", err
	}

	user, err := deployer.dataStore.User().Read(userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch the user")
	}

	if user.Role == portaineree.AdministratorRole {
		return tokenManager.GetAdminServiceAccountToken(), nil
	}

	token, err := tokenManager.GetUserServiceAccountToken(int(user.ID), int(endpoint.ID))
	if err != nil {
		return "", err
	}

	if token == "" {
		return "", fmt.Errorf("can not get a valid user service account token")
	}

	return token, nil
}

func (deployer *KubernetesDeployer) DeployViaKubeConfig(kubeConfig string, clusterID string, manifestFile string) error {
	kubeConfigPath := filepath.Join(os.TempDir(), clusterID)
	err := filesystem.WriteToFile(kubeConfigPath, []byte(kubeConfig))
	if err != nil {
		return err
	}

	command := path.Join(deployer.binaryPath, "kubectl")
	if runtime.GOOS == "windows" {
		command = path.Join(deployer.binaryPath, "kubectl.exe")
	}

	args := []string{"--kubeconfig", kubeConfigPath}
	args = append(args, "apply", "-f", strings.TrimSpace(manifestFile))

	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "POD_NAMESPACE=default")
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "failed to execute kubectl command: %q", stderr.String())
	}

	fmt.Println(string(output))

	return nil
}

// Deploy upserts Kubernetes resources defined in manifest(s)
func (deployer *KubernetesDeployer) Deploy(userID portainer.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return deployer.kubectl("apply", "", manifestFilesToArgs(manifestFiles), namespace, endpoint, userID)
}

// Restart calls restart a kubernetes resource. Valid resource types are: deployment, statefulset, daemonset
// call with
func (deployer *KubernetesDeployer) Restart(userID portainer.UserID, endpoint *portaineree.Endpoint, resourceList []string, namespace string) (string, error) {
	return deployer.kubectl("rollout", "restart", resourceList, namespace, endpoint, userID)
}

// Remove deletes Kubernetes resources defined in manifest(s)
func (deployer *KubernetesDeployer) Remove(userID portainer.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return deployer.kubectl("delete", "", manifestFilesToArgs(manifestFiles), namespace, endpoint, userID)
}

func (deployer *KubernetesDeployer) kubectl(cmd, subcmd string, args []string, namespace string, endpoint *portaineree.Endpoint, userID portainer.UserID) (string, error) {
	kubectlCmd := path.Join(deployer.binaryPath, "kubectl")
	if runtime.GOOS == "windows" {
		kubectlCmd = path.Join(deployer.binaryPath, "kubectl.exe")
	}

	token, err := deployer.getToken(userID, endpoint, endpoint.Type == portaineree.KubernetesLocalEnvironment)
	if err != nil {
		return "", errors.Wrap(err, "failed generating a user token")
	}

	cmdArgs := []string{}
	if token != "" {
		cmdArgs = append(cmdArgs, "--token", token)
	}

	if namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", namespace)
	}

	if endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
		url, proxy, err := deployer.getAgentURL(endpoint)
		if err != nil {
			return "", errors.WithMessage(err, "failed generating endpoint URL")
		}
		defer proxy.Close()

		cmdArgs = append(cmdArgs, "--server", url)
		cmdArgs = append(cmdArgs, "--insecure-skip-tls-verify")
	}

	cmdArgs = append(cmdArgs, cmd)
	if subcmd != "" {
		cmdArgs = append(cmdArgs, subcmd)
	}
	cmdArgs = append(cmdArgs, args...)

	log.Debug().Msgf("kubectl %+v\n", cmdArgs)

	// Add --ignore-not-found=true to delete command to match CE codebase
	for _, arg := range cmdArgs {
		if arg == "delete" {
			cmdArgs = append(cmdArgs, "--ignore-not-found=true")
			break
		}
	}

	var stderr bytes.Buffer
	c := exec.Command(kubectlCmd, cmdArgs...)
	c.Env = os.Environ()
	c.Env = append(c.Env, "POD_NAMESPACE=default")
	c.Stderr = &stderr

	output, err := c.Output()
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute kubectl command: %q", stderr.String())
	}

	return string(output), nil
}

func manifestFilesToArgs(manifestFiles []string) []string {
	args := []string{}
	for _, path := range manifestFiles {
		args = append(args, "-f", strings.TrimSpace(path))
	}
	return args
}

// ConvertCompose leverages the kompose binary to deploy a compose compliant manifest.
func (deployer *KubernetesDeployer) ConvertCompose(data []byte) ([]byte, error) {
	command := path.Join(deployer.binaryPath, "kompose")
	if runtime.GOOS == "windows" {
		command = path.Join(deployer.binaryPath, "kompose.exe")
	}

	args := make([]string, 0)
	args = append(args, "convert", "-f", "-", "--stdout")

	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New(stderr.String())
	}

	return output, nil
}

func (deployer *KubernetesDeployer) getAgentURL(endpoint *portaineree.Endpoint) (string, *factory.ProxyServer, error) {
	proxy, err := deployer.proxyManager.CreateAgentProxyServer(endpoint)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("http://127.0.0.1:%d/kubernetes", proxy.Port), proxy, nil
}

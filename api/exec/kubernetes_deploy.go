package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/proxy/factory"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

// KubernetesDeployer represents a service to deploy resources inside a Kubernetes environment(endpoint).
type KubernetesDeployer struct {
	binaryPath                  string
	dataStore                   dataservices.DataStore
	reverseTunnelService        portaineree.ReverseTunnelService
	signatureService            portaineree.DigitalSignatureService
	kubernetesClientFactory     *cli.ClientFactory
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
	authService                 *authorization.Service
	proxyManager                *proxy.Manager
}

// NewKubernetesDeployer initializes a new KubernetesDeployer service.
func NewKubernetesDeployer(authService *authorization.Service, kubernetesTokenCacheManager *kubernetes.TokenCacheManager, kubernetesClientFactory *cli.ClientFactory, datastore dataservices.DataStore, reverseTunnelService portaineree.ReverseTunnelService, signatureService portaineree.DigitalSignatureService, proxyManager *proxy.Manager, binaryPath string) *KubernetesDeployer {
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

func (deployer *KubernetesDeployer) getToken(userID portaineree.UserID, endpoint *portaineree.Endpoint, setLocalAdminToken bool) (string, error) {
	kubeCLI, err := deployer.kubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return "", err
	}

	tokenCache := deployer.kubernetesTokenCacheManager.GetOrCreateTokenCache(int(endpoint.ID))

	tokenManager, err := kubernetes.NewTokenManager(kubeCLI, deployer.dataStore, tokenCache, setLocalAdminToken, deployer.authService)
	if err != nil {
		return "", err
	}

	user, err := deployer.dataStore.User().User(userID)
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

// Deploy upserts Kubernetes resources defined in manifest(s)
func (deployer *KubernetesDeployer) Deploy(userID portaineree.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return deployer.command("apply", userID, endpoint, manifestFiles, namespace)
}

// Remove deletes Kubernetes resources defined in manifest(s)
func (deployer *KubernetesDeployer) Remove(userID portaineree.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return deployer.command("delete", userID, endpoint, manifestFiles, namespace)
}

func (deployer *KubernetesDeployer) command(operation string, userID portaineree.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	token, err := deployer.getToken(userID, endpoint, endpoint.Type == portaineree.KubernetesLocalEnvironment)
	if err != nil {
		return "", errors.Wrap(err, "failed generating a user token")
	}

	command := path.Join(deployer.binaryPath, "kubectl")
	if runtime.GOOS == "windows" {
		command = path.Join(deployer.binaryPath, "kubectl.exe")
	}

	args := []string{"--token", token}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	if endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
		url, proxy, err := deployer.getAgentURL(endpoint)
		if err != nil {
			return "", errors.WithMessage(err, "failed generating endpoint URL")
		}

		defer proxy.Close()
		args = append(args, "--server", url)
		args = append(args, "--insecure-skip-tls-verify")
	}

	if operation == "delete" {
		args = append(args, "--ignore-not-found=true")
	}

	args = append(args, operation)
	for _, path := range manifestFiles {
		args = append(args, "-f", strings.TrimSpace(path))
	}

	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute kubectl command: %q", stderr.String())
	}

	return string(output), nil
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

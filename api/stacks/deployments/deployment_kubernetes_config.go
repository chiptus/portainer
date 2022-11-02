package deployments

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	"github.com/portainer/portainer/api/filesystem"
)

type KubernetesStackDeploymentConfig struct {
	namespaces              []string
	stack                   *portaineree.Stack
	kuberneteDeployer       portaineree.KubernetesDeployer
	appLabel                k.KubeAppLabels
	tokenData               *portaineree.TokenData
	endpoint                *portaineree.Endpoint
	authorizationService    *authorization.Service
	kubernetesClientFactory *cli.ClientFactory
	output                  string
}

func CreateKubernetesStackDeploymentConfig(stack *portaineree.Stack,
	kubeDeployer portaineree.KubernetesDeployer,
	appLabels k.KubeAppLabels,
	tokenData *portaineree.TokenData,
	endpoint *portaineree.Endpoint,
	authService *authorization.Service,
	k8sClientFactory *cli.ClientFactory) (*KubernetesStackDeploymentConfig, error) {

	return &KubernetesStackDeploymentConfig{
		stack:                   stack,
		kuberneteDeployer:       kubeDeployer,
		appLabel:                appLabels,
		tokenData:               tokenData,
		endpoint:                endpoint,
		authorizationService:    authService,
		kubernetesClientFactory: k8sClientFactory,
	}, nil
}

func (config *KubernetesStackDeploymentConfig) GetUsername() string {
	return config.tokenData.Username
}

func (config *KubernetesStackDeploymentConfig) Deploy() error {
	fileNames := stackutils.GetStackFilePaths(config.stack, false)

	manifestFilePaths := make([]string, len(fileNames))

	namespaces := map[string]bool{}
	if config.stack.Namespace != "" {
		namespaces[config.stack.Namespace] = true
	}

	tmpDir, err := os.MkdirTemp("", "kub_deployment")
	if err != nil {
		return errors.Wrap(err, "failed to create temp kub deployment directory")
	}

	defer os.RemoveAll(tmpDir)

	for _, fileName := range fileNames {
		manifestFilePath := filesystem.JoinPaths(tmpDir, fileName)
		manifestContent, err := os.ReadFile(filesystem.JoinPaths(config.stack.ProjectPath, fileName))
		if err != nil {
			return errors.Wrap(err, "failed to read manifest file")
		}

		if config.stack.IsComposeFormat {
			manifestContent, err = config.kuberneteDeployer.ConvertCompose(manifestContent)
			if err != nil {
				return errors.Wrap(err, "failed to convert docker compose file to a kube manifest")
			}
		}

		manifestContent, err = k.AddAppLabels(manifestContent, config.appLabel.ToMap())
		if err != nil {
			return errors.Wrap(err, "failed to add application labels")
		}

		env := map[string]string{}
		for _, v := range config.stack.Env {
			env[v.Name] = v.Value
		}

		manifestContent, err = k.UpdateContainerEnv(manifestContent, env)
		if err != nil {
			return errors.Wrap(err, "failed to add application vars")
		}

		// get resource namespace, fallback to provided namespace if not explicit on resource
		resourceNamespace, err := kubernetes.GetNamespace(manifestContent)
		if err != nil {
			return errors.Wrap(err, "failed to get resource namespace")
		}

		if resourceNamespace != "" {
			namespaces[resourceNamespace] = true
		}

		err = filesystem.WriteToFile(manifestFilePath, []byte(manifestContent))
		if err != nil {
			return errors.Wrap(err, "failed to create temp manifest file")
		}

		manifestFilePaths = append(manifestFilePaths, manifestFilePath)
	}

	namespacesList := make([]string, 0, len(namespaces))
	for namespace := range namespaces {
		namespacesList = append(namespacesList, namespace)
	}
	config.namespaces = namespacesList

	if config.authorizationService != nil && config.kubernetesClientFactory != nil {
		err := config.checkEndpointPermission()
		if err != nil {
			return fmt.Errorf("user does not have permission to deploy stack: %w", err)
		}
	}

	output, err := config.kuberneteDeployer.Deploy(config.tokenData.ID, config.endpoint, manifestFilePaths, config.stack.Namespace)
	if err != nil {
		return fmt.Errorf("failed to deploy kubernete stack: %w", err)
	}

	config.output = output
	return nil
}

func (config *KubernetesStackDeploymentConfig) GetResponse() string {
	return config.output
}

func (config *KubernetesStackDeploymentConfig) checkEndpointPermission() error {
	permissionDeniedErr := errors.New("Permission denied to access environment")

	if config.tokenData.Role == portaineree.AdministratorRole {
		return nil
	}

	// check if the user has OperationK8sApplicationsAdvancedDeploymentRW access in the environment(endpoint)
	endpointRole, err := config.authorizationService.GetUserEndpointRole(int(config.tokenData.ID), int(config.endpoint.ID))
	if err != nil {
		return errors.Wrap(err, "failed to retrieve user endpoint role")
	}
	if !endpointRole.Authorizations[portaineree.OperationK8sApplicationsAdvancedDeploymentRW] {
		return permissionDeniedErr
	}

	// will skip if user can access all namespaces
	if endpointRole.Authorizations[portaineree.OperationK8sAccessAllNamespaces] {
		return nil
	}

	cli, err := config.kubernetesClientFactory.GetKubeClient(config.endpoint)
	if err != nil {
		return errors.Wrap(err, "unable to create Kubernetes client")
	}

	// check if the user has RW access to the namespace
	namespaceAuthorizations, err := config.authorizationService.GetNamespaceAuthorizations(int(config.tokenData.ID), *config.endpoint, cli)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve user namespace authorizations")
	}

	// if no namespace provided, either by form or by manifest, use the default namespace
	if len(config.namespaces) == 0 {
		config.namespaces = []string{"default"}
	}

	for _, namespace := range config.namespaces {
		if auth, ok := namespaceAuthorizations[namespace]; !ok || !auth[portaineree.OperationK8sAccessNamespaceWrite] {
			return errors.Wrap(permissionDeniedErr, "user does not have permission to access namespace")
		}
	}

	return nil
}

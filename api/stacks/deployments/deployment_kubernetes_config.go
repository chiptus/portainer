package deployments

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/rs/zerolog/log"
)

type KubernetesStackDeploymentConfig struct {
	namespaces              []string
	stack                   *portaineree.Stack
	kubernetesDeployer      portaineree.KubernetesDeployer
	appLabel                k.KubeAppLabels
	tokenData               *portainer.TokenData
	endpoint                *portaineree.Endpoint
	authorizationService    *authorization.Service
	kubernetesClientFactory *cli.ClientFactory
	output                  string
}

func CreateKubernetesStackDeploymentConfig(stack *portaineree.Stack,
	kubeDeployer portaineree.KubernetesDeployer,
	appLabels k.KubeAppLabels,
	tokenData *portainer.TokenData,
	endpoint *portaineree.Endpoint,
	authService *authorization.Service,
	k8sClientFactory *cli.ClientFactory) (*KubernetesStackDeploymentConfig, error) {

	return &KubernetesStackDeploymentConfig{
		stack:                   stack,
		kubernetesDeployer:      kubeDeployer,
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

	manifestFilePaths := make([]string, 0, len(fileNames))

	namespaces := map[string]bool{}
	if config.stack.Namespace != "" {
		namespaces[config.stack.Namespace] = true
	}

	tmpDir, err := os.MkdirTemp("", "kub_deployment")
	if err != nil {
		return errors.Wrap(err, "failed to create temp kub deployment directory")
	}
	defer os.RemoveAll(tmpDir)

	projectPath := stackutils.GetStackProjectPathByVersion(config.stack)
	if strings.Contains(config.stack.ProjectPath, "kub_file_content") {
		// The project path is a temporary path for updating k8s stack,
		// we need to use the temporary path instead of the version path
		projectPath = config.stack.ProjectPath
	}

	for _, fileName := range fileNames {
		manifestFilePath := filesystem.JoinPaths(tmpDir, fileName)
		manifestContent, err := os.ReadFile(filesystem.JoinPaths(projectPath, fileName))
		if err != nil {
			return errors.Wrap(err, "failed to read manifest file")
		}

		if config.stack.IsComposeFormat {
			manifestContent, err = config.kubernetesDeployer.ConvertCompose(manifestContent)
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

	output, err := config.kubernetesDeployer.Deploy(config.tokenData.ID, config.endpoint, manifestFilePaths, config.stack.Namespace)
	if err != nil {
		return fmt.Errorf("failed to deploy kubernete stack: %w", err)
	}

	config.output = output
	return nil
}

// Restart the following resources
func (config *KubernetesStackDeploymentConfig) Restart(resourceList []string) error {

	log.Debug().Msgf("Restarting stack")

	fileNames := stackutils.GetStackFilePaths(config.stack, false)
	resourceMap := make(map[string][]string)
	namespaces := []string{}

	tmpDir, err := os.MkdirTemp("", "kub_restart")
	if err != nil {
		return errors.Wrap(err, "failed to create temp kube directory")
	}
	defer os.RemoveAll(tmpDir)

	projectPath := stackutils.GetStackProjectPathByVersion(config.stack)
	for _, fileName := range fileNames {
		manifestFilePath := filesystem.JoinPaths(tmpDir, fileName)
		manifestContent, err := os.ReadFile(filesystem.JoinPaths(projectPath, fileName))
		if err != nil {
			return errors.Wrap(err, "failed to read manifest file")
		}

		if config.stack.IsComposeFormat {
			manifestContent, err = config.kubernetesDeployer.ConvertCompose(manifestContent)
			if err != nil {
				return errors.Wrap(err, "failed to convert docker compose file to a kube manifest")
			}
		}

		filters := []string{"deployment", "statefulset", "daemonset"}
		manifestResources, err := kubernetes.GetResourcesFromManifest(manifestContent, filters)
		if err != nil {
			return errors.Wrap(err, "failed to get resources")
		}

		err = filesystem.WriteToFile(manifestFilePath, []byte(manifestContent))
		if err != nil {
			return errors.Wrap(err, "failed to create temp manifest file")
		}

		for _, r := range manifestResources {
			namespace := r.Namespace
			if namespace == "" && config.stack.Namespace != "" {
				namespace = config.stack.Namespace
			}

			resourceMap[namespace] = append(resourceMap[namespace], r.Kind+"/"+r.Name)
		}
	}

	for ns := range resourceMap {
		namespaces = append(namespaces, ns)
	}

	config.namespaces = namespaces

	if config.authorizationService != nil && config.kubernetesClientFactory != nil {
		err = config.checkEndpointPermission()
		if err != nil {
			return fmt.Errorf("user does not have permission to restart stack: %w", err)
		}
	}

	log.Debug().Msgf("Restarting resources in namespaces: %v", namespaces)

	// Now restart all the resources
	err = nil
	for namespace, resources := range resourceMap {
		if resourceList != nil {
			resources = filterResources(resources, resourceList)
			if len(resources) == 0 {
				continue
			}
		}

		log.Debug().Msgf("Namespace: %s, Resources: %+v", namespace, resources)

		output, e := config.kubernetesDeployer.Restart(config.tokenData.ID, config.endpoint, resources, namespace)
		if e != nil {
			log.Error().Err(err).Msgf("Failed to restart resources %v in namespace %v", resources, namespace)
			err = fmt.Errorf("Some resources failed to restart, check Portainer log for more details")
		}

		if len(config.output) > 0 {
			config.output += "\n---\n"
		}
		config.output += output
	}

	return err
}

// filterResources removes resources not in resourceList to prevent restarting resources we don't want to restart
func filterResources(resources, resourceList []string) []string {
	log.Debug().Msgf("Filtering resources: %v, resourceList %v", resources, resourceList)
	if len(resourceList) == 0 {
		return resourceList
	}

	filteredResources := []string{}
	for _, r := range resources {
		for _, rl := range resourceList {
			if strings.EqualFold(r, rl) {
				filteredResources = append(filteredResources, r)
			}
		}
	}

	log.Debug().Msgf("Filtered resources: %v", filteredResources)
	return filteredResources
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
	namespaceAuthorizations, err := config.authorizationService.GetNamespaceAuthorizations(nil, int(config.tokenData.ID), *config.endpoint, cli)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve user namespace authorizations")
	}

	// if no namespace provided, either by form or by manifest, use the default namespace
	if len(config.namespaces) == 0 {
		config.namespaces = []string{"default"}
	}

	for _, namespace := range config.namespaces {
		if auth, ok := namespaceAuthorizations[namespace]; !ok || !auth[portaineree.OperationK8sAccessNamespaceWrite] {
			return errors.Wrapf(permissionDeniedErr, "user does not have permission to access namespace: %s", namespace)
		}
	}

	return nil
}

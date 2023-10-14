package stackbuilders

import (
	"fmt"
	"strconv"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
)

type K8sStackFileContentBuilder struct {
	FileContentMethodStackBuilder
	stackCreateMut          *sync.Mutex
	KuberneteDeployer       portaineree.KubernetesDeployer
	TokenData               *portainer.TokenData
	AuthorizationService    *authorization.Service
	KubernetesClientFactory *cli.ClientFactory
}

// CreateK8sStackFileContentBuilder creates a builder for the Kubernetes stack that will be deployed by file content method
func CreateK8sStackFileContentBuilder(dataStore dataservices.DataStore,
	fileService portainer.FileService,
	stackDeployer deployments.StackDeployer,
	kuberneteDeployer portaineree.KubernetesDeployer,
	tokenData *portainer.TokenData,
	AuthorizationService *authorization.Service,
	KubernetesClientFactory *cli.ClientFactory) *K8sStackFileContentBuilder {

	return &K8sStackFileContentBuilder{
		FileContentMethodStackBuilder: FileContentMethodStackBuilder{
			StackBuilder: CreateStackBuilder(dataStore, fileService, stackDeployer),
		},
		stackCreateMut:          &sync.Mutex{},
		KuberneteDeployer:       kuberneteDeployer,
		TokenData:               tokenData,
		AuthorizationService:    AuthorizationService,
		KubernetesClientFactory: KubernetesClientFactory,
	}
}

func (b *K8sStackFileContentBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) FileContentMethodStackBuildProcess {
	b.FileContentMethodStackBuilder.SetGeneralInfo(payload, endpoint)
	return b
}

func (b *K8sStackFileContentBuilder) SetUniqueInfo(payload *StackPayload) FileContentMethodStackBuildProcess {
	if b.hasError() {
		return b
	}
	b.stack.Name = payload.StackName
	b.stack.Type = portaineree.KubernetesStack
	b.stack.EntryPoint = filesystem.ManifestFileDefaultName
	b.stack.Namespace = payload.Namespace
	b.stack.CreatedBy = b.TokenData.Username
	b.stack.IsComposeFormat = payload.ComposeFormat
	b.stack.FromAppTemplate = payload.FromAppTemplate
	return b
}

func (b *K8sStackFileContentBuilder) SetFileContent(payload *StackPayload) FileContentMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	stackFolder := strconv.Itoa(int(b.stack.ID))
	projectPath, err := b.fileService.StoreStackFileFromBytesByVersion(stackFolder, b.stack.EntryPoint, b.stack.StackFileVersion, []byte(payload.StackFileContent))
	if err != nil {
		fileType := "Manifest"
		if b.stack.IsComposeFormat {
			fileType = "Compose"
		}
		errMsg := fmt.Sprintf("Unable to persist Kubernetes %s file on disk", fileType)
		b.err = httperror.InternalServerError(errMsg, err)
		return b
	}
	b.stack.ProjectPath = projectPath

	return b
}

func (b *K8sStackFileContentBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) FileContentMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	b.stackCreateMut.Lock()
	defer b.stackCreateMut.Unlock()

	k8sAppLabel := k.KubeAppLabels{
		StackID:   int(b.stack.ID),
		StackName: b.stack.Name,
		Owner:     stackutils.SanitizeLabel(b.stack.CreatedBy),
		Kind:      "content",
	}

	k8sDeploymentConfig, err := deployments.CreateKubernetesStackDeploymentConfig(b.stack, b.KuberneteDeployer, k8sAppLabel, b.TokenData, endpoint, b.AuthorizationService, b.KubernetesClientFactory)
	if err != nil {
		b.err = httperror.InternalServerError("failed to create temp kub deployment files", err)
		return b
	}

	b.deploymentConfiger = k8sDeploymentConfig

	return b.FileContentMethodStackBuilder.Deploy(payload, endpoint)
}

func (b *K8sStackFileContentBuilder) GetResponse() string {
	return b.FileContentMethodStackBuilder.deploymentConfiger.GetResponse()
}

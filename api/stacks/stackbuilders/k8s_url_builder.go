package stackbuilders

import (
	"strconv"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
)

type KubernetesStackUrlBuilder struct {
	UrlMethodStackBuilder
	stackCreateMut          *sync.Mutex
	KuberneteDeployer       portaineree.KubernetesDeployer
	TokenData               *portainer.TokenData
	AuthorizationService    *authorization.Service
	KubernetesClientFactory *cli.ClientFactory
}

// CreateKuberntesStackGitBuilder creates a builder for the Kubernetes stack that will be deployed by git repository method
func CreateKubernetesStackUrlBuilder(dataStore dataservices.DataStore,
	fileService portainer.FileService,
	stackDeployer deployments.StackDeployer,
	kuberneteDeployer portaineree.KubernetesDeployer,
	tokenData *portainer.TokenData,
	AuthorizationService *authorization.Service,
	KubernetesClientFactory *cli.ClientFactory) *KubernetesStackUrlBuilder {

	return &KubernetesStackUrlBuilder{
		UrlMethodStackBuilder: UrlMethodStackBuilder{
			StackBuilder: CreateStackBuilder(dataStore, fileService, stackDeployer),
		},
		stackCreateMut:          &sync.Mutex{},
		KuberneteDeployer:       kuberneteDeployer,
		TokenData:               tokenData,
		AuthorizationService:    AuthorizationService,
		KubernetesClientFactory: KubernetesClientFactory,
	}
}

func (b *KubernetesStackUrlBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) UrlMethodStackBuildProcess {
	b.UrlMethodStackBuilder.SetGeneralInfo(payload, endpoint)
	return b
}

func (b *KubernetesStackUrlBuilder) SetUniqueInfo(payload *StackPayload) UrlMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	b.stack.Type = portaineree.KubernetesStack
	b.stack.Namespace = payload.Namespace
	b.stack.Name = payload.StackName
	b.stack.EntryPoint = filesystem.ManifestFileDefaultName
	b.stack.CreatedBy = b.TokenData.Username
	b.stack.IsComposeFormat = payload.ComposeFormat
	return b
}

func (b *KubernetesStackUrlBuilder) SetURL(payload *StackPayload) UrlMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	var manifestContent []byte
	manifestContent, err := client.Get(payload.ManifestURL, 30)
	if err != nil {
		b.err = httperror.InternalServerError("Unable to retrieve manifest from URL", err)
		return b
	}

	stackFolder := strconv.Itoa(int(b.stack.ID))
	projectPath, err := b.fileService.StoreStackFileFromBytesByVersion(stackFolder, b.stack.EntryPoint, b.stack.StackFileVersion, manifestContent)
	if err != nil {
		b.err = httperror.InternalServerError("Unable to persist Kubernetes manifest file on disk", err)
		return b
	}
	b.stack.ProjectPath = projectPath

	return b
}

func (b *KubernetesStackUrlBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) UrlMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	b.stackCreateMut.Lock()
	defer b.stackCreateMut.Unlock()

	k8sAppLabel := k.KubeAppLabels{
		StackID:   int(b.stack.ID),
		StackName: b.stack.Name,
		Owner:     stackutils.SanitizeLabel(b.stack.CreatedBy),
		Kind:      "url",
	}

	k8sDeploymentConfig, err := deployments.CreateKubernetesStackDeploymentConfig(b.stack, b.KuberneteDeployer, k8sAppLabel, b.TokenData, endpoint, b.AuthorizationService, b.KubernetesClientFactory)
	if err != nil {
		b.err = httperror.InternalServerError("failed to create temp kub deployment files", err)
		return b
	}

	b.deploymentConfiger = k8sDeploymentConfig

	return b.UrlMethodStackBuilder.Deploy(payload, endpoint)
}

func (b *KubernetesStackUrlBuilder) GetResponse() string {
	return b.UrlMethodStackBuilder.deploymentConfiger.GetResponse()
}

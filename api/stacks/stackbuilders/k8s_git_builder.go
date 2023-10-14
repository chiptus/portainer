package stackbuilders

import (
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
)

type KubernetesStackGitBuilder struct {
	GitMethodStackBuilder
	stackCreateMut          *sync.Mutex
	KuberneteDeployer       portaineree.KubernetesDeployer
	TokenData               *portainer.TokenData
	AuthorizationService    *authorization.Service
	KubernetesClientFactory *cli.ClientFactory
}

// CreateKuberntesStackGitBuilder creates a builder for the Kubernetes stack that will be deployed by git repository method
func CreateKubernetesStackGitBuilder(userActivityService portaineree.UserActivityService,
	dataStore dataservices.DataStore,
	fileService portainer.FileService,
	gitService portainer.GitService,
	scheduler *scheduler.Scheduler,
	stackDeployer deployments.StackDeployer,
	kuberneteDeployer portaineree.KubernetesDeployer,
	tokenData *portainer.TokenData,
	AuthorizationService *authorization.Service,
	KubernetesClientFactory *cli.ClientFactory) *KubernetesStackGitBuilder {

	return &KubernetesStackGitBuilder{
		GitMethodStackBuilder: GitMethodStackBuilder{
			StackBuilder:        CreateStackBuilder(dataStore, fileService, stackDeployer),
			userActivityService: userActivityService,
			gitService:          gitService,
			scheduler:           scheduler,
		},
		stackCreateMut:          &sync.Mutex{},
		KuberneteDeployer:       kuberneteDeployer,
		TokenData:               tokenData,
		AuthorizationService:    AuthorizationService,
		KubernetesClientFactory: KubernetesClientFactory,
	}
}

func (b *KubernetesStackGitBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess {
	b.GitMethodStackBuilder.SetGeneralInfo(payload, endpoint)
	return b
}

func (b *KubernetesStackGitBuilder) SetUniqueInfo(payload *StackPayload) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	b.stack.Type = portaineree.KubernetesStack
	b.stack.Namespace = payload.Namespace
	b.stack.Name = payload.StackName
	b.stack.EntryPoint = payload.ManifestFile
	b.stack.CreatedBy = b.TokenData.Username
	b.stack.IsComposeFormat = payload.ComposeFormat
	return b
}

func (b *KubernetesStackGitBuilder) SetGitRepository(payload *StackPayload, userID portainer.UserID) GitMethodStackBuildProcess {
	b.GitMethodStackBuilder.SetGitRepository(payload, userID)
	return b
}

func (b *KubernetesStackGitBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	b.stackCreateMut.Lock()
	defer b.stackCreateMut.Unlock()

	k8sAppLabel := k.KubeAppLabels{
		StackID:   int(b.stack.ID),
		StackName: b.stack.Name,
		Owner:     stackutils.SanitizeLabel(b.stack.CreatedBy),
		Kind:      "git",
	}

	k8sDeploymentConfig, err := deployments.CreateKubernetesStackDeploymentConfig(b.stack, b.KuberneteDeployer, k8sAppLabel, b.TokenData, endpoint, b.AuthorizationService, b.KubernetesClientFactory)
	if err != nil {
		b.err = httperror.InternalServerError("failed to create temp kub deployment files", err)
		return b
	}

	b.deploymentConfiger = k8sDeploymentConfig

	return b.GitMethodStackBuilder.Deploy(payload, endpoint)
}

func (b *KubernetesStackGitBuilder) SetAutoUpdate(payload *StackPayload) GitMethodStackBuildProcess {
	b.GitMethodStackBuilder.SetAutoUpdate(payload)
	return b
}

func (b *KubernetesStackGitBuilder) GetResponse() string {
	return b.GitMethodStackBuilder.deploymentConfiger.GetResponse()
}

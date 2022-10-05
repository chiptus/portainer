package stackbuilders

import (
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	portainer "github.com/portainer/portainer/api"
)

type ComposeStackGitBuilder struct {
	GitMethodStackBuilder
	SecurityContext *security.RestrictedRequestContext
}

// CreateComposeStackGitBuilder creates a builder for the compose stack (docker standalone) that will be deployed by git repository method
func CreateComposeStackGitBuilder(securityContext *security.RestrictedRequestContext,
	userActivityService portaineree.UserActivityService,
	dataStore dataservices.DataStore,
	fileService portainer.FileService,
	gitService portaineree.GitService,
	scheduler *scheduler.Scheduler,
	stackDeployer deployments.StackDeployer) *ComposeStackGitBuilder {

	return &ComposeStackGitBuilder{
		GitMethodStackBuilder: GitMethodStackBuilder{
			StackBuilder:        CreateStackBuilder(dataStore, fileService, stackDeployer),
			userActivityService: userActivityService,
			gitService:          gitService,
			scheduler:           scheduler,
		},
		SecurityContext: securityContext,
	}
}

func (b *ComposeStackGitBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess {
	b.GitMethodStackBuilder.SetGeneralInfo(payload, endpoint)
	return b
}

func (b *ComposeStackGitBuilder) SetUniqueInfo(payload *StackPayload) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}
	b.stack.Name = payload.Name
	b.stack.Type = portaineree.DockerComposeStack
	b.stack.EntryPoint = payload.ComposeFile
	b.stack.FromAppTemplate = payload.FromAppTemplate
	b.stack.Env = payload.Env
	return b
}

func (b *ComposeStackGitBuilder) SetGitRepository(payload *StackPayload, userID portaineree.UserID) GitMethodStackBuildProcess {
	b.GitMethodStackBuilder.SetGitRepository(payload, userID)
	return b
}

func (b *ComposeStackGitBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	composeDeploymentConfig, err := deployments.CreateComposeStackDeploymentConfig(b.SecurityContext, b.stack, endpoint, b.dataStore, b.fileService, b.stackDeployer, false, false)
	if err != nil {
		b.err = httperror.InternalServerError(err.Error(), err)
		return b
	}

	b.deploymentConfiger = composeDeploymentConfig
	b.stack.CreatedBy = b.deploymentConfiger.GetUsername()

	return b.GitMethodStackBuilder.Deploy(payload, endpoint)
}

func (b *ComposeStackGitBuilder) SetAutoUpdate(payload *StackPayload) GitMethodStackBuildProcess {
	b.GitMethodStackBuilder.SetAutoUpdate(payload)
	return b
}

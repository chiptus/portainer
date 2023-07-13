package stackbuilders

import (
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
)

type SwarmStackFileContentBuilder struct {
	FileContentMethodStackBuilder
	SecurityContext *security.RestrictedRequestContext
}

// CreateSwarmStackFileContentBuilder creates a builder for the swarm stack that will be deployed by file content method
func CreateSwarmStackFileContentBuilder(securityContext *security.RestrictedRequestContext,
	dataStore dataservices.DataStore,
	fileService portainer.FileService,
	stackDeployer deployments.StackDeployer) *SwarmStackFileContentBuilder {

	return &SwarmStackFileContentBuilder{
		FileContentMethodStackBuilder: FileContentMethodStackBuilder{
			StackBuilder: CreateStackBuilder(dataStore, fileService, stackDeployer),
		},
		SecurityContext: securityContext,
	}
}

func (b *SwarmStackFileContentBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) FileContentMethodStackBuildProcess {
	b.FileContentMethodStackBuilder.SetGeneralInfo(payload, endpoint)
	return b
}

func (b *SwarmStackFileContentBuilder) SetUniqueInfo(payload *StackPayload) FileContentMethodStackBuildProcess {
	if b.hasError() {
		return b
	}
	b.stack.Name = payload.Name
	b.stack.Type = portaineree.DockerSwarmStack
	b.stack.SwarmID = payload.SwarmID
	b.stack.EntryPoint = filesystem.ComposeFileDefaultName
	b.stack.Env = payload.Env
	b.stack.FromAppTemplate = payload.FromAppTemplate
	b.stack.Webhook = payload.Webhook
	return b
}

func (b *SwarmStackFileContentBuilder) SetFileContent(payload *StackPayload) FileContentMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	stackFolder := strconv.Itoa(int(b.stack.ID))
	projectPath, err := b.fileService.StoreStackFileFromBytesByVersion(stackFolder, b.stack.EntryPoint, b.stack.StackFileVersion, []byte(payload.StackFileContent))
	if err != nil {
		b.err = httperror.InternalServerError("Unable to persist Compose file on disk", err)
		return b
	}
	b.stack.ProjectPath = projectPath

	return b
}

func (b *SwarmStackFileContentBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) FileContentMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	swarmDeploymentConfig, err := deployments.CreateSwarmStackDeploymentConfig(b.SecurityContext, b.stack, endpoint, b.dataStore, b.fileService, b.stackDeployer, false, true)
	if err != nil {
		b.err = httperror.InternalServerError(err.Error(), err)
		return b
	}

	b.deploymentConfiger = swarmDeploymentConfig
	b.stack.CreatedBy = b.deploymentConfiger.GetUsername()

	return b.FileContentMethodStackBuilder.Deploy(payload, endpoint)
}

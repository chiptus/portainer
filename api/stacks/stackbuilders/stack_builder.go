package stackbuilders

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/rs/zerolog/log"
)

type StackBuilder struct {
	stack              *portaineree.Stack
	dataStore          dataservices.DataStore
	fileService        portainer.FileService
	stackDeployer      deployments.StackDeployer
	deploymentConfiger deployments.StackDeploymentConfiger
	err                *httperror.HandlerError
	doCleanUp          bool
}

func CreateStackBuilder(dataStore dataservices.DataStore, fileService portainer.FileService, deployer deployments.StackDeployer) StackBuilder {
	return StackBuilder{
		stack: &portaineree.Stack{
			StackFileVersion: 1, // when creating a stack, we always set the version to 1
		},
		dataStore:     dataStore,
		fileService:   fileService,
		stackDeployer: deployer,
		doCleanUp:     true,
	}
}

func (b *StackBuilder) SaveStack() (*portaineree.Stack, *httperror.HandlerError) {
	defer b.cleanUp()
	if b.hasError() {
		return nil, b.err
	}

	if b.stack.GitConfig != nil && b.stack.GitConfig.Authentication != nil &&
		b.stack.GitConfig.Authentication.GitCredentialID != 0 {
		// prevent the username and password from saving into db if the git
		// credential is used
		b.stack.GitConfig.Authentication.Username = ""
		b.stack.GitConfig.Authentication.Password = ""
	}

	err := b.dataStore.Stack().Create(b.stack)
	if err != nil {
		b.err = httperror.InternalServerError("Unable to persist the stack inside the database", err)
		return nil, b.err
	}

	b.doCleanUp = false

	return b.stack, nil
}

func (b *StackBuilder) cleanUp() error {
	if !b.doCleanUp {
		return nil
	}

	err := b.fileService.RemoveDirectory(b.stack.ProjectPath)
	if err != nil {
		log.Error().Err(err).Msg("unable to cleanup stack creation")
	}

	return nil
}

func (b *StackBuilder) hasError() bool {
	return b.err != nil
}

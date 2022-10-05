package stackbuilders

import (
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	portainer "github.com/portainer/portainer/api"
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
		stack:         &portaineree.Stack{},
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

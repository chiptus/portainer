package stackbuilders

import (
	"time"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
)

type UrlMethodStackBuildProcess interface {
	// Set general stack information
	SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) UrlMethodStackBuildProcess
	// Set unique stack information, e.g. swarm stack has swarmID, kubernetes stack has namespace
	SetUniqueInfo(payload *StackPayload) UrlMethodStackBuildProcess
	// Deploy stack based on the configuration
	Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) UrlMethodStackBuildProcess
	// Save the stack information to database
	SaveStack() (*portaineree.Stack, *httperror.HandlerError)
	// Get reponse from http request. Use if it is needed
	GetResponse() string
	// Set manifest url
	SetURL(payload *StackPayload) UrlMethodStackBuildProcess
}

type UrlMethodStackBuilder struct {
	StackBuilder
}

func (b *UrlMethodStackBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) UrlMethodStackBuildProcess {
	stackID := b.dataStore.Stack().GetNextIdentifier()
	b.stack.ID = portaineree.StackID(stackID)
	b.stack.EndpointID = endpoint.ID
	b.stack.Status = portaineree.StackStatusActive
	b.stack.CreationDate = time.Now().Unix()
	return b
}

func (b *UrlMethodStackBuilder) SetUniqueInfo(payload *StackPayload) UrlMethodStackBuildProcess {

	return b
}

func (b *UrlMethodStackBuilder) SetURL(payload *StackPayload) UrlMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	return b
}

func (b *UrlMethodStackBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) UrlMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	// Deploy the stack
	err := b.deploymentConfiger.Deploy()
	if err != nil {
		b.err = httperror.InternalServerError(err.Error(), err)
		return b
	}

	return b
}

func (b *UrlMethodStackBuilder) GetResponse() string {
	return ""
}
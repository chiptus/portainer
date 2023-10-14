package edgeasynccommand

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

func (service ServiceTx) generateKey(cmd *portaineree.EdgeAsyncCommand) []byte {
	return append(service.service.connection.ConvertToKey(int(cmd.EndpointID)), service.service.connection.ConvertToKey(cmd.ID)...)
}

// Create assigns an ID to a new EdgeAsyncCommand and saves it.
func (service ServiceTx) Create(cmd *portaineree.EdgeAsyncCommand) error {
	cmd.ID = service.tx.GetNextIdentifier(BucketName)
	log.Debug().Str("command", fmt.Sprintf("%v", cmd)).Msg("create EdgeAsyncCommand")

	return service.tx.CreateObjectWithStringId(BucketName, service.generateKey(cmd), cmd)
}

// Update updates an EdgeAsyncCommand.
func (service ServiceTx) Update(id int, cmd *portaineree.EdgeAsyncCommand) error {
	log.Debug().Str("command", fmt.Sprintf("%v", cmd)).Int("id", cmd.ID).Msg("update EdgeAsyncCommand")

	return service.tx.UpdateObject(BucketName, service.generateKey(cmd), cmd)
}

// EndpointCommands return all EdgeAsyncCommand associated to a given endpoint.
func (service ServiceTx) EndpointCommands(endpointID portainer.EndpointID) ([]portaineree.EdgeAsyncCommand, error) {
	var cmds = make([]portaineree.EdgeAsyncCommand, 0)

	return cmds, service.tx.GetAllWithKeyPrefix(
		BucketName,
		service.service.connection.ConvertToKey(int(endpointID)),
		&portaineree.EdgeAsyncCommand{},
		dataservices.AppendFn(&cmds),
	)
}

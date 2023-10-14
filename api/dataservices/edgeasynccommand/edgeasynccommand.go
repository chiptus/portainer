package edgeasynccommand

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "edge_async_command"

// Service represents a service for managing Edge Async Commands data.
type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

func (service *Service) generateKey(cmd *portaineree.EdgeAsyncCommand) []byte {
	return append(service.connection.ConvertToKey(int(cmd.EndpointID)), service.connection.ConvertToKey(cmd.ID)...)
}

// Create assigns an ID to a new EdgeAsyncCommand and saves it.
func (service *Service) Create(cmd *portaineree.EdgeAsyncCommand) error {
	cmd.ID = service.connection.GetNextIdentifier(BucketName)
	log.Debug().Str("command", fmt.Sprintf("%v", cmd)).Msg("create EdgeAsyncCommand")

	return service.connection.CreateObjectWithStringId(BucketName, service.generateKey(cmd), cmd)
}

// Update updates an EdgeAsyncCommand.
func (service *Service) Update(id int, cmd *portaineree.EdgeAsyncCommand) error {
	log.Debug().Str("command", fmt.Sprintf("%v", cmd)).Int("id", cmd.ID).Msg("update EdgeAsyncCommand")

	return service.connection.UpdateObject(BucketName, service.generateKey(cmd), cmd)
}

// EndpointCommands return all EdgeAsyncCommand associated to a given endpoint.
func (service *Service) EndpointCommands(endpointID portainer.EndpointID) ([]portaineree.EdgeAsyncCommand, error) {
	var cmds = make([]portaineree.EdgeAsyncCommand, 0)

	return cmds, service.connection.GetAllWithKeyPrefix(
		BucketName,
		service.connection.ConvertToKey(int(endpointID)),
		&portaineree.EdgeAsyncCommand{},
		dataservices.AppendFn(&cmds),
	)
}

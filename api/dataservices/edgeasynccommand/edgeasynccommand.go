package edgeasynccommand

import (
	"fmt"

	"github.com/sirupsen/logrus"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/boltdb"
	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edge_async_command"
)

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

func (service *Service) generateKey(cmd *portaineree.EdgeAsyncCommand) []byte {
	return append(service.connection.ConvertToKey(int(cmd.EndpointID)), service.connection.ConvertToKey(cmd.ID)...)
}

// Create assigns an ID to a new EdgeAsyncCommand and saves it.
func (service *Service) Create(cmd *portaineree.EdgeAsyncCommand) error {
	cmd.ID = service.connection.GetNextIdentifier(BucketName)
	logrus.WithField("command", cmd).Info("Create EdgeAsyncCommand")

	return service.connection.CreateObjectWithStringId(BucketName, service.generateKey(cmd), cmd)
}

// Update updates an EdgeAsyncCommand.
func (service *Service) Update(id int, cmd *portaineree.EdgeAsyncCommand) error {
	logrus.WithField("command", cmd).WithField("id", cmd.ID).Info("Update EdgeAsyncCommand")
	return service.connection.UpdateObject(BucketName, service.generateKey(cmd), cmd)
}

// EndpointCommands return all EdgeAsyncCommand associated to a given endpoint.
func (service *Service) EndpointCommands(endpointID portaineree.EndpointID) ([]portaineree.EdgeAsyncCommand, error) {
	var cmds = make([]portaineree.EdgeAsyncCommand, 0)

	err := service.connection.(*boltdb.DbConnection).GetAllWithKeyPrefix(
		BucketName,
		service.connection.ConvertToKey(int(endpointID)),
		&portaineree.EdgeAsyncCommand{},
		func(obj interface{}) (interface{}, error) {
			cmd, ok := obj.(*portaineree.EdgeAsyncCommand)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to EdgeAsyncCommand object")
				return nil, fmt.Errorf("failed to convert to EdgeAsyncCommand object: %s", obj)
			}

			cmds = append(cmds, *cmd)

			return &portaineree.EdgeAsyncCommand{}, nil
		})

	return cmds, err
}

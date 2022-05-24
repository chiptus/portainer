package edgeasynccommand

import (
	"fmt"

	"github.com/sirupsen/logrus"

	portaineree "github.com/portainer/portainer-ee/api"
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

// Create assign an ID to a new EdgeAsyncCommand and saves it.
func (service *Service) Create(command *portaineree.EdgeAsyncCommand) error {
	command.ID = service.connection.GetNextIdentifier(BucketName)
	logrus.WithField("command", command).Info("Create EdgeAsyncCommand")
	return service.connection.CreateObjectWithSetSequence(BucketName, command.ID, command)
}

// Update updates an EdgeAsyncCommand.
func (service *Service) Update(id int, command *portaineree.EdgeAsyncCommand) error {
	identifier := service.connection.ConvertToKey(id)
	logrus.WithField("command", command).WithField("id", id).Info("Update EdgeAsyncCommand")
	return service.connection.UpdateObject(BucketName, identifier, command)
}

// Delete deletes an EdgeAsyncCommand.
func (service *Service) Delete(id int) error {
	identifier := service.connection.ConvertToKey(id)
	logrus.WithField("id", id).Info("Update EdgeAsyncCommand")
	return service.connection.DeleteObject(BucketName, identifier)
}

// EndpointCommands return all EdgeAsyncCommand associated to a given endpoint.
func (service *Service) EndpointCommands(endpointID portaineree.EndpointID) ([]portaineree.EdgeAsyncCommand, error) {
	var commands = make([]portaineree.EdgeAsyncCommand, 0)

	err := service.connection.GetAllWithJsoniter(
		BucketName,
		&portaineree.EdgeAsyncCommand{},
		func(obj interface{}) (interface{}, error) {
			command, ok := obj.(*portaineree.EdgeAsyncCommand)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to EdgeAsyncCommand object")
				return nil, fmt.Errorf("failed to convert to EdgeAsyncCommand object: %s", obj)
			}
			if command.EndpointID == endpointID {
				commands = append(commands, *command)
			}
			return &portaineree.EdgeAsyncCommand{}, nil
		})

	return commands, err
}

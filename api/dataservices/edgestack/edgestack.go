package edgestack

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edge_stack"
)

// Service represents a service for managing Edge stack data.
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

// EdgeStacks returns an array containing all edge stacks
func (service *Service) EdgeStacks() ([]portaineree.EdgeStack, error) {
	var stacks = make([]portaineree.EdgeStack, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.EdgeStack{},
		func(obj interface{}) (interface{}, error) {
			//var tag portaineree.Tag
			stack, ok := obj.(*portaineree.EdgeStack)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EdgeStack object")
				return nil, fmt.Errorf("Failed to convert to EdgeStack object: %s", obj)
			}

			stacks = append(stacks, *stack)

			return &portaineree.EdgeStack{}, nil
		})

	return stacks, err
}

// EdgeStack returns an Edge stack by ID.
func (service *Service) EdgeStack(ID portaineree.EdgeStackID) (*portaineree.EdgeStack, error) {
	var stack portaineree.EdgeStack
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

// CreateEdgeStack saves an Edge stack object to db.
func (service *Service) Create(id portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error {

	edgeStack.ID = id

	return service.connection.CreateObjectWithId(
		BucketName,
		int(edgeStack.ID),
		edgeStack,
	)
}

// UpdateEdgeStack updates an Edge stack.
func (service *Service) UpdateEdgeStack(ID portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, edgeStack)
}

// DeleteEdgeStack deletes an Edge stack.
func (service *Service) DeleteEdgeStack(ID portaineree.EdgeStackID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}

package edgestacklog

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edge_stack_log"
)

// Service represents a service for managing Edge Stack logs.
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

func (service *Service) generateKey(edgeStackID portaineree.EdgeStackID, endpointID portaineree.EndpointID) []byte {
	return append(service.connection.ConvertToKey(int(edgeStackID)),
		service.connection.ConvertToKey(int(endpointID))...)
}

// Create creates an EdgeStackLog and saves it.
func (service *Service) Create(edgeStackLog *portaineree.EdgeStackLog) error {
	key := service.generateKey(edgeStackLog.EdgeStackID, edgeStackLog.EndpointID)
	return service.connection.CreateObjectWithStringId(BucketName, key, edgeStackLog)
}

// Update updates an EdgeStackLog.
func (service *Service) Update(edgeStackLog *portaineree.EdgeStackLog) error {
	key := service.generateKey(edgeStackLog.EdgeStackID, edgeStackLog.EndpointID)
	return service.connection.UpdateObject(BucketName, key, edgeStackLog)
}

// Delete deletes an EdgeStackLog.
func (service *Service) Delete(edgeStackID portaineree.EdgeStackID, endpointID portaineree.EndpointID) error {
	key := service.generateKey(edgeStackID, endpointID)
	return service.connection.DeleteObject(BucketName, key)
}

func (service *Service) EdgeStackLog(edgeStackID portaineree.EdgeStackID, endpointID portaineree.EndpointID) (*portaineree.EdgeStackLog, error) {
	key := service.generateKey(edgeStackID, endpointID)
	o := &portaineree.EdgeStackLog{}

	return o, service.connection.GetObject(BucketName, key, o)
}

package edgegroup

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "edgegroups"

// Service represents a service for managing Edge group data.
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

// EdgeGroups return a slice containing all the Edge groups.
func (service *Service) EdgeGroups() ([]portaineree.EdgeGroup, error) {
	var groups []portaineree.EdgeGroup
	var err error

	err = service.connection.ViewTx(func(tx portainer.Transaction) error {
		groups, err = service.Tx(tx).EdgeGroups()
		return err
	})

	return groups, err
}

// EdgeGroup returns an Edge group by ID.
func (service *Service) EdgeGroup(ID portaineree.EdgeGroupID) (*portaineree.EdgeGroup, error) {
	var group *portaineree.EdgeGroup
	var err error

	err = service.connection.ViewTx(func(tx portainer.Transaction) error {
		group, err = service.Tx(tx).EdgeGroup(ID)
		return err
	})

	return group, err
}

// UpdateEdgeGroup updates an edge group.
func (service *Service) UpdateEdgeGroup(ID portaineree.EdgeGroupID, group *portaineree.EdgeGroup) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, group)
}

// UpdateEdgeGroupFunc updates an edge group inside a transaction avoiding data races.
func (service *Service) UpdateEdgeGroupFunc(ID portaineree.EdgeGroupID, updateFunc func(edgeGroup *portaineree.EdgeGroup)) error {
	id := service.connection.ConvertToKey(int(ID))
	edgeGroup := &portaineree.EdgeGroup{}

	return service.connection.UpdateObjectFunc(BucketName, id, edgeGroup, func() {
		updateFunc(edgeGroup)
	})
}

// DeleteEdgeGroup deletes an Edge group.
func (service *Service) DeleteEdgeGroup(ID portaineree.EdgeGroupID) error {
	return service.connection.UpdateTx(func(tx portainer.Transaction) error {
		return service.Tx(tx).DeleteEdgeGroup(ID)
	})
}

// CreateEdgeGroup assign an ID to a new Edge group and saves it.
func (service *Service) Create(group *portaineree.EdgeGroup) error {
	return service.connection.UpdateTx(func(tx portainer.Transaction) error {
		return service.Tx(tx).Create(group)
	})
}

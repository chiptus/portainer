package edgeupdateschedule

import (
	"github.com/portainer/portainer-ee/api/dataservices"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "edge_update_schedule"

// Service represents a service for managing Edge Update Schedule data.
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

// List return an array containing all the items in the bucket.
func (service *Service) List() ([]edgetypes.UpdateSchedule, error) {
	var list = make([]edgetypes.UpdateSchedule, 0)

	return list, service.connection.GetAll(
		BucketName,
		&edgetypes.UpdateSchedule{},
		dataservices.AppendFn(&list),
	)
}

// Item returns an item by ID.
func (service *Service) Item(ID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error) {
	var item edgetypes.UpdateSchedule
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// Create assign an ID to a new object and saves it.
func (service *Service) Create(item *edgetypes.UpdateSchedule) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			item.ID = edgetypes.UpdateScheduleID(id)
			return int(item.ID), item
		},
	)
}

// Update updates an item.
func (service *Service) Update(id edgetypes.UpdateScheduleID, item *edgetypes.UpdateSchedule) error {
	identifier := service.connection.ConvertToKey(int(id))

	return service.connection.UpdateObject(BucketName, identifier, item)
}

// Delete deletes an item.
func (service *Service) Delete(id edgetypes.UpdateScheduleID) error {
	identifier := service.connection.ConvertToKey(int(id))

	return service.connection.DeleteObject(BucketName, identifier)
}

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
	dataservices.BaseDataService[edgetypes.UpdateSchedule, edgetypes.UpdateScheduleID]
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
		BaseDataService: dataservices.BaseDataService[edgetypes.UpdateSchedule, edgetypes.UpdateScheduleID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: service.BaseDataService.Tx(tx),
	}
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

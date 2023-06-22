package cloudprovisioning

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "cloud_provisioning"

// Service represents a service for managing edge jobs data.
type Service struct {
	dataservices.BaseDataService[portaineree.CloudProvisioningTask, portaineree.CloudProvisioningTaskID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.CloudProvisioningTask, portaineree.CloudProvisioningTaskID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

// Create assign an ID to a new Cloud Provisioning Task and saves it.
func (service *Service) Create(task *portaineree.CloudProvisioningTask) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			task.ID = portaineree.CloudProvisioningTaskID(id)
			return int(task.ID), task
		},
	)
}

// GetNextIdentifier returns the next identifier for a cloud provisionng task
func (service *Service) GetNextIdentifier() int {
	return service.Connection.GetNextIdentifier(BucketName)
}

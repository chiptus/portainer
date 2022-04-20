package cloudprovisioning

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "cloud_provisioning"
)

// Service represents a service for managing edge jobs data.
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

// Tasks returns a list of Cloud Provisioning Tasks
func (service *Service) Tasks() ([]portaineree.CloudProvisioningTask, error) {
	var cloudTasks = make([]portaineree.CloudProvisioningTask, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.CloudProvisioningTask{},
		func(obj interface{}) (interface{}, error) {
			task, ok := obj.(*portaineree.CloudProvisioningTask)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to CloudProvisioningTask object")
				return nil, fmt.Errorf("Failed to convert to CloudProvisioningTask object: %s", obj)
			}
			cloudTasks = append(cloudTasks, *task)
			return &portaineree.CloudProvisioningTask{}, nil
		})

	return cloudTasks, err
}

// Task returns an Cloud Provisioning Task by ID
func (service *Service) Task(ID portaineree.CloudProvisioningTaskID) (*portaineree.CloudProvisioningTask, error) {
	var cloudTask portaineree.CloudProvisioningTask
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &cloudTask)
	if err != nil {
		return nil, err
	}

	return &cloudTask, nil
}

// Create assign an ID to a new Cloud Provisioning Task and saves it.
func (service *Service) Create(task *portaineree.CloudProvisioningTask) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			task.ID = portaineree.CloudProvisioningTaskID(id)
			return int(task.ID), task
		},
	)
}

// Update updates a cloud provisioning task by ID
func (service *Service) Update(ID portaineree.CloudProvisioningTaskID, task *portaineree.CloudProvisioningTask) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, task)
}

// Delete deletes a cloud provisioning task by ID
func (service *Service) Delete(ID portaineree.CloudProvisioningTaskID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for a cloud provisionng task
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}

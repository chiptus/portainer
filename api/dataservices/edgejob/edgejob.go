package edgejob

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edgejobs"
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

// EdgeJobs returns a list of Edge jobs
func (service *Service) EdgeJobs() ([]portaineree.EdgeJob, error) {
	var edgeJobs = make([]portaineree.EdgeJob, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.EdgeJob{},
		func(obj interface{}) (interface{}, error) {
			//var tag portaineree.Tag
			job, ok := obj.(*portaineree.EdgeJob)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to EdgeJob object")
				return nil, fmt.Errorf("Failed to convert to EdgeJob object: %s", obj)
			}
			edgeJobs = append(edgeJobs, *job)
			return &portaineree.EdgeJob{}, nil
		})

	return edgeJobs, err
}

// EdgeJob returns an Edge job by ID
func (service *Service) EdgeJob(ID portaineree.EdgeJobID) (*portaineree.EdgeJob, error) {
	var edgeJob portaineree.EdgeJob
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &edgeJob)
	if err != nil {
		return nil, err
	}

	return &edgeJob, nil
}

// Create creates a new EdgeJob
func (service *Service) Create(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	edgeJob.ID = ID

	return service.connection.CreateObjectWithId(
		BucketName,
		int(edgeJob.ID),
		edgeJob,
	)
}

// UpdateEdgeJob updates an Edge job by ID
func (service *Service) UpdateEdgeJob(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, edgeJob)
}

// DeleteEdgeJob deletes an Edge job
func (service *Service) DeleteEdgeJob(ID portaineree.EdgeJobID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}

package dockerhub

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName   = "dockerhub"
	dockerHubKey = "DOCKERHUB"
)

// Service represents a service for managing Dockerhub data.
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

// DockerHub returns the DockerHub object.
func (service *Service) DockerHub() (*portaineree.DockerHub, error) {
	var dockerhub portaineree.DockerHub

	err := service.connection.GetObject(BucketName, []byte(dockerHubKey), &dockerhub)
	if err != nil {
		return nil, err
	}

	return &dockerhub, nil
}

// UpdateDockerHub updates a DockerHub object.
func (service *Service) UpdateDockerHub(dockerhub *portaineree.DockerHub) error {
	return service.connection.UpdateObject(BucketName, []byte(dockerHubKey), dockerhub)
}

package helmuserrepository

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "helm_user_repository"

// Service represents a service for managing environment(endpoint) data.
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

// HelmUserRepository returns an array of all HelmUserRepository
func (service *Service) HelmUserRepositories() ([]portaineree.HelmUserRepository, error) {
	var repos = make([]portaineree.HelmUserRepository, 0)

	return repos, service.connection.GetAll(
		BucketName,
		&portaineree.HelmUserRepository{},
		dataservices.AppendFn(&repos),
	)
}

// HelmUserRepositoryByUserID return an array containing all the HelmUserRepository objects where the specified userID is present.
func (service *Service) HelmUserRepositoryByUserID(userID portaineree.UserID) ([]portaineree.HelmUserRepository, error) {
	var result = make([]portaineree.HelmUserRepository, 0)

	return result, service.connection.GetAll(
		BucketName,
		&portaineree.HelmUserRepository{},
		dataservices.FilterFn(&result, func(e portaineree.HelmUserRepository) bool {
			return e.UserID == userID
		}),
	)
}

// CreateHelmUserRepository creates a new HelmUserRepository object.
func (service *Service) Create(record *portaineree.HelmUserRepository) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.HelmUserRepositoryID(id)
			return int(record.ID), record
		},
	)
}

// UpdateHelmUserRepostory updates an registry.
func (service *Service) UpdateHelmUserRepository(ID portaineree.HelmUserRepositoryID, registry *portaineree.HelmUserRepository) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, registry)
}

// DeleteHelmUserRepository deletes an registry.
func (service *Service) DeleteHelmUserRepository(ID portaineree.HelmUserRepositoryID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

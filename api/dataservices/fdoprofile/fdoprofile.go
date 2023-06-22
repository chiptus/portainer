package fdoprofile

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "fdo_profiles"

// Service represents a service for managingFDO Profiles data.
type Service struct {
	dataservices.BaseDataService[portaineree.FDOProfile, portaineree.FDOProfileID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.FDOProfile, portaineree.FDOProfileID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

// Create assign an ID to a new FDO Profile and saves it.
func (service *Service) Create(fdoProfile *portaineree.FDOProfile) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			fdoProfile.ID = portaineree.FDOProfileID(id)
			return int(fdoProfile.ID), fdoProfile
		},
	)
}

// GetNextIdentifier returns the next identifier for a FDO Profile.
func (service *Service) GetNextIdentifier() int {
	return service.Connection.GetNextIdentifier(BucketName)
}

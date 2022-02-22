package fdoprofile

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "fdo_profiles"
)

// Service represents a service for managingFDO Profiles data.
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

// FDOProfiles return an array containing all the FDO Profiles.
func (service *Service) FDOProfiles() ([]portaineree.FDOProfile, error) {
	var fdoProfiles = make([]portaineree.FDOProfile, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.FDOProfile{},
		func(obj interface{}) (interface{}, error) {
			fdoProfile, ok := obj.(*portaineree.FDOProfile)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to FDOProfile object")
				return nil, fmt.Errorf("Failed to convert to FDOProfile object: %s", obj)
			}
			fdoProfiles = append(fdoProfiles, *fdoProfile)
			return &portaineree.FDOProfile{}, nil
		})

	return fdoProfiles, err
}

// FDOProfile returns an FDO Profile by ID.
func (service *Service) FDOProfile(ID portaineree.FDOProfileID) (*portaineree.FDOProfile, error) {
	var fdoProfile portaineree.FDOProfile
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &fdoProfile)
	if err != nil {
		return nil, err
	}

	return &fdoProfile, nil
}

// Create assign an ID to a new FDO Profile and saves it.
func (service *Service) Create(fdoProfile *portaineree.FDOProfile) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			fdoProfile.ID = portaineree.FDOProfileID(id)
			return int(fdoProfile.ID), fdoProfile
		},
	)
}

// Update updates an FDO Profile.
func (service *Service) Update(ID portaineree.FDOProfileID, fdoProfile *portaineree.FDOProfile) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, fdoProfile)
}

// Delete deletes an FDO Profile.
func (service *Service) Delete(ID portaineree.FDOProfileID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for a FDO Profile.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}

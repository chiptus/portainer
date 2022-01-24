package fdoprofile

import (
	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "fdo_profiles"
)

// Service represents a service for managingFDO Profiles data.
type Service struct {
	connection *internal.DbConnection
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
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

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var fdoProfile portaineree.FDOProfile
			err := internal.UnmarshalObject(v, &fdoProfile)
			if err != nil {
				return err
			}
			fdoProfiles = append(fdoProfiles, fdoProfile)
		}

		return nil
	})

	return fdoProfiles, err
}

// FDOProfile returns an FDO Profile by ID.
func (service *Service) FDOProfile(ID portaineree.FDOProfileID) (*portaineree.FDOProfile, error) {
	var fdoProfile portaineree.FDOProfile
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &fdoProfile)
	if err != nil {
		return nil, err
	}

	return &fdoProfile, nil
}

// Create assign an ID to a new FDO Profile and saves it.
func (service *Service) Create(fdoProfile *portaineree.FDOProfile) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		data, err := internal.MarshalObject(fdoProfile)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(fdoProfile.ID)), data)
	})
}

// Update updates an FDO Profile.
func (service *Service) Update(ID portaineree.FDOProfileID, fdoProfile *portaineree.FDOProfile) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, fdoProfile)
}

// Delete deletes an FDO Profile.
func (service *Service) Delete(ID portaineree.FDOProfileID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for a FDO Profile.
func (service *Service) GetNextIdentifier() int {
	return internal.GetNextIdentifier(service.connection, BucketName)
}

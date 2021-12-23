package role

import (
	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "roles"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection *internal.DbConnection
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

// Role returns a Role by ID
func (service *Service) Role(ID portaineree.RoleID) (*portaineree.Role, error) {
	var set portaineree.Role
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &set)
	if err != nil {
		return nil, err
	}

	return &set, nil
}

// Roles return an array containing all the sets.
func (service *Service) Roles() ([]portaineree.Role, error) {
	var sets = make([]portaineree.Role, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var set portaineree.Role
			err := internal.UnmarshalObject(v, &set)
			if err != nil {
				return err
			}
			sets = append(sets, set)
		}

		return nil
	})

	return sets, err
}

// CreateRole creates a new Role.
func (service *Service) CreateRole(role *portaineree.Role) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		if role.ID == 0 {
			role.ID = portaineree.RoleID(id)
		}

		data, err := internal.MarshalObject(role)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(role.ID)), data)
	})
}

// UpdateRole updates a role.
func (service *Service) UpdateRole(ID portaineree.RoleID, role *portaineree.Role) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, role)
}

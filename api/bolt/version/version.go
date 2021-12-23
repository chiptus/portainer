package version

import (
	"strconv"

	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/errors"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName         = "version"
	versionKey         = "DB_VERSION"
	previousVersionKey = "PREVIOUS_DB_VERSION"
	instanceKey        = "INSTANCE_ID"
	editionKey         = "EDITION"
	updatingKey        = "DB_UPDATING"
)

// Service represents a service to manage stored versions.
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

// Edition retrieves the stored portainer edition.
func (service *Service) Edition() (portaineree.SoftwareEdition, error) {
	editionData, err := service.getKey(editionKey)
	if err != nil {
		return 0, err
	}

	edition, err := strconv.Atoi(string(editionData))
	if err != nil {
		return 0, err
	}

	return portaineree.SoftwareEdition(edition), nil
}

// StoreEdition store the portainer edition.
func (service *Service) StoreEdition(edition portaineree.SoftwareEdition) error {
	return service.setKey(editionKey, strconv.Itoa(int(edition)))
}

// PreviousDBVersion retrieves the stored database version.
func (service *Service) PreviousDBVersion() (int, error) {
	version, err := service.getKey(previousVersionKey)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(version))
}

// DBVersion retrieves the stored database version.
func (service *Service) DBVersion() (int, error) {
	version, err := service.getKey(versionKey)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(version))
}

// StorePreviousDBVersion store the database version.
func (service *Service) StorePreviousDBVersion(version int) error {
	return service.setKey(previousVersionKey, strconv.Itoa(version))
}

// StoreDBVersion store the database version.
func (service *Service) StoreDBVersion(version int) error {
	return service.setKey(versionKey, strconv.Itoa(version))
}

// IsUpdating retrieves the database updating status.
func (service *Service) IsUpdating() (bool, error) {
	isUpdating, err := service.getKey(updatingKey)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(string(isUpdating))
}

// StoreIsUpdating store the database updating status.
func (service *Service) StoreIsUpdating(isUpdating bool) error {
	return service.setKey(updatingKey, strconv.FormatBool(isUpdating))
}

// InstanceID retrieves the stored instance ID.
func (service *Service) InstanceID() (string, error) {
	instanceID, err := service.getKey(instanceKey)
	if err != nil {
		return "", err
	}

	return string(instanceID), nil
}

// StoreInstanceID store the instance ID.
func (service *Service) StoreInstanceID(ID string) error {
	return service.setKey(instanceKey, ID)
}

func (service *Service) getKey(key string) ([]byte, error) {
	var data []byte

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		value := bucket.Get([]byte(key))
		if value == nil {
			return errors.ErrObjectNotFound
		}

		data = make([]byte, len(value))
		copy(data, value)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (service *Service) setKey(key string, value string) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		data := []byte(value)
		return bucket.Put([]byte(key), data)
	})
}

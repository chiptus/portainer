package apikeyrepository

import (
	"bytes"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices/errors"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "api_key"
)

// Service represents a service for managing api-key data.
type Service struct {
	connection portainer.Connection
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

// GetAPIKeysByUserID returns a slice containing all the APIKeys a user has access to.
func (service *Service) GetAPIKeysByUserID(userID portaineree.UserID) ([]portaineree.APIKey, error) {
	var result = make([]portaineree.APIKey, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.APIKey{},
		func(obj interface{}) (interface{}, error) {
			record, ok := obj.(*portaineree.APIKey)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to APIKey object")
				return nil, fmt.Errorf("Failed to convert to APIKey object: %s", obj)
			}

			if record.UserID == userID {
				result = append(result, *record)
			}

			return &portaineree.APIKey{}, nil
		})

	return result, err
}

// GetAPIKeyByDigest returns the API key for the associated digest.
// Note: there is a 1-to-1 mapping of api-key and digest
func (service *Service) GetAPIKeyByDigest(digest []byte) (*portaineree.APIKey, error) {
	var k *portaineree.APIKey
	stop := fmt.Errorf("ok")
	err := service.connection.GetAll(
		BucketName,
		&portaineree.APIKey{},
		func(obj interface{}) (interface{}, error) {
			key, ok := obj.(*portaineree.APIKey)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to APIKey object")
				return nil, fmt.Errorf("Failed to convert to APIKey object: %s", obj)
			}
			if bytes.Equal(key.Digest, digest) {
				k = key
				return nil, stop
			}

			return &portaineree.APIKey{}, nil
		})

	if err == stop {
		return k, nil
	}

	if err == nil {
		return nil, errors.ErrObjectNotFound
	}

	return nil, err
}

// Create creates a new APIKey object.
func (service *Service) Create(record *portaineree.APIKey) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.APIKeyID(id)

			return int(record.ID), record
		},
	)
}

// GetAPIKey retrieves an existing APIKey object by api key ID.
func (service *Service) GetAPIKey(keyID portaineree.APIKeyID) (*portaineree.APIKey, error) {
	var key portaineree.APIKey
	identifier := service.connection.ConvertToKey(int(keyID))

	err := service.connection.GetObject(BucketName, identifier, &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (service *Service) UpdateAPIKey(key *portaineree.APIKey) error {
	identifier := service.connection.ConvertToKey(int(key.ID))
	return service.connection.UpdateObject(BucketName, identifier, key)
}

func (service *Service) DeleteAPIKey(ID portaineree.APIKeyID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

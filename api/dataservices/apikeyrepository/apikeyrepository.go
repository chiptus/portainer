package apikeyrepository

import (
	"bytes"
	"errors"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/rs/zerolog/log"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "api_key"

// Service represents a service for managing api-key data.
type Service struct {
	dataservices.BaseDataService[portaineree.APIKey, portaineree.APIKeyID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.APIKey, portaineree.APIKeyID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

// GetAPIKeysByUserID returns a slice containing all the APIKeys a user has access to.
func (service *Service) GetAPIKeysByUserID(userID portaineree.UserID) ([]portaineree.APIKey, error) {
	var result = make([]portaineree.APIKey, 0)

	err := service.Connection.GetAll(
		BucketName,
		&portaineree.APIKey{},
		func(obj interface{}) (interface{}, error) {
			record, ok := obj.(*portaineree.APIKey)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to APIKey object")
				return nil, fmt.Errorf("failed to convert to APIKey object: %s", obj)
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
	err := service.Connection.GetAll(
		BucketName,
		&portaineree.APIKey{},
		func(obj interface{}) (interface{}, error) {
			key, ok := obj.(*portaineree.APIKey)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to APIKey object")
				return nil, fmt.Errorf("failed to convert to APIKey object: %s", obj)
			}
			if bytes.Equal(key.Digest, digest) {
				k = key
				return nil, stop
			}

			return &portaineree.APIKey{}, nil
		})

	if errors.Is(err, stop) {
		return k, nil
	}

	if err == nil {
		return nil, dserrors.ErrObjectNotFound
	}

	return nil, err
}

// Create creates a new APIKey object.
func (service *Service) Create(record *portaineree.APIKey) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.APIKeyID(id)

			return int(record.ID), record
		},
	)
}

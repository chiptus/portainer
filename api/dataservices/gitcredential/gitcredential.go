package gitcredential

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices/errors"
	"github.com/sirupsen/logrus"
)

const BucketName = "git_credentials"

type Service struct {
	connection portainer.Connection
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// Create creates a new git credential object.
func (service *Service) Create(record *portaineree.GitCredential) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.GitCredentialID(id)

			return int(record.ID), record
		},
	)
}

// GetGitCredentials returns an array containing all git-credentials. It is necessary for export feature
func (service *Service) GetGitCredentials() ([]portaineree.GitCredential, error) {
	var result = make([]portaineree.GitCredential, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		func(obj interface{}) (interface{}, error) {
			record, ok := obj.(*portaineree.GitCredential)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to GitCredential object")
				return nil, fmt.Errorf("Failed to convert to GitCredential object: %s", obj)
			}

			result = append(result, *record)
			return &portaineree.GitCredential{}, nil
		})

	return result, err
}

// GetGitCredential retrieves a single GitCredential object by credential ID.
func (service *Service) GetGitCredential(credID portaineree.GitCredentialID) (*portaineree.GitCredential, error) {
	var cred portaineree.GitCredential
	identifier := service.connection.ConvertToKey(int(credID))

	err := service.connection.GetObject(BucketName, identifier, &cred)
	if err != nil {
		return nil, err
	}

	return &cred, nil
}

func (service *Service) UpdateGitCredential(ID portaineree.GitCredentialID, cred *portaineree.GitCredential) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, cred)
}

func (service *Service) DeleteGitCredential(ID portaineree.GitCredentialID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetGitCredentialsByUserID returns an array containing all git-credentials owned by a specific user
func (service *Service) GetGitCredentialsByUserID(userID portaineree.UserID) ([]portaineree.GitCredential, error) {
	var result = make([]portaineree.GitCredential, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		func(obj interface{}) (interface{}, error) {
			record, ok := obj.(*portaineree.GitCredential)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to GitCredential object")
				return nil, fmt.Errorf("Failed to convert to GitCredential object: %s", obj)
			}

			if record.UserID == userID {
				result = append(result, *record)
			}
			return &portaineree.GitCredential{}, nil
		})

	return result, err
}

// GetGitCredentialByName retrieves a single GitCredential object owned by a specific user with an unique git credential name
func (service *Service) GetGitCredentialByName(userID portaineree.UserID, name string) (*portaineree.GitCredential, error) {
	var credential *portaineree.GitCredential
	stop := fmt.Errorf("ok")
	err := service.connection.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		func(obj interface{}) (interface{}, error) {
			cred, ok := obj.(*portaineree.GitCredential)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to GitCredential object")
				return nil, fmt.Errorf("Failed to convert to GitCredential object: %s", obj)
			}

			if cred.UserID == userID && cred.Name == name {
				credential = cred
				return nil, stop
			}
			return &portaineree.GitCredential{}, nil
		})
	if err == stop {
		return credential, nil
	}
	if err == nil {
		return nil, errors.ErrObjectNotFound
	}

	return nil, err
}

package gitcredential

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
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

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
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

	return result, service.connection.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		dataservices.AppendFn(&result),
	)
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

	return result, service.connection.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		dataservices.FilterFn(&result, func(e portaineree.GitCredential) bool {
			return e.UserID == userID
		}),
	)
}

// GetGitCredentialByName retrieves a single GitCredential object owned by a specific user with a unique git credential name
func (service *Service) GetGitCredentialByName(userID portaineree.UserID, name string) (*portaineree.GitCredential, error) {
	var credential portaineree.GitCredential

	err := service.connection.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		dataservices.FirstFn(&credential, func(e portaineree.GitCredential) bool {
			return e.UserID == userID && e.Name == name
		}),
	)

	if errors.Is(err, dataservices.ErrStop) {
		return &credential, nil
	}

	if err == nil {
		return nil, dserrors.ErrObjectNotFound
	}

	return nil, err
}

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
	dataservices.BaseDataService[portaineree.GitCredential, portaineree.GitCredentialID]
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.GitCredential, portaineree.GitCredentialID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.GitCredential, portaineree.GitCredentialID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// Create creates a new git credential object.
func (service *Service) Create(record *portaineree.GitCredential) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.GitCredentialID(id)

			return int(record.ID), record
		},
	)
}

// GetGitCredentialsByUserID returns an array containing all git-credentials owned by a specific user
func (service *Service) GetGitCredentialsByUserID(userID portaineree.UserID) ([]portaineree.GitCredential, error) {
	var result = make([]portaineree.GitCredential, 0)

	return result, service.Connection.GetAll(
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

	err := service.Connection.GetAll(
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

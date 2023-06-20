package gitcredential

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

// Create creates a new git credential object.
func (service ServiceTx) Create(record *portaineree.GitCredential) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.GitCredentialID(id)

			return int(record.ID), record
		},
	)
}

// GetGitCredentials returns an array containing all git-credentials. It is necessary for export feature
func (service ServiceTx) GetGitCredentials() ([]portaineree.GitCredential, error) {
	var result = make([]portaineree.GitCredential, 0)

	return result, service.tx.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		dataservices.AppendFn(&result),
	)
}

// GetGitCredential retrieves a single GitCredential object by credential ID.
func (service ServiceTx) GetGitCredential(credID portaineree.GitCredentialID) (*portaineree.GitCredential, error) {
	var cred portaineree.GitCredential
	identifier := service.service.connection.ConvertToKey(int(credID))

	err := service.tx.GetObject(BucketName, identifier, &cred)
	if err != nil {
		return nil, err
	}

	return &cred, nil
}

func (service ServiceTx) UpdateGitCredential(ID portaineree.GitCredentialID, cred *portaineree.GitCredential) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, cred)
}

func (service ServiceTx) DeleteGitCredential(ID portaineree.GitCredentialID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}

// GetGitCredentialsByUserID returns an array containing all git-credentials owned by a specific user
func (service ServiceTx) GetGitCredentialsByUserID(userID portaineree.UserID) ([]portaineree.GitCredential, error) {
	var result = make([]portaineree.GitCredential, 0)

	return result, service.tx.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		dataservices.FilterFn(&result, func(e portaineree.GitCredential) bool {
			return e.UserID == userID
		}),
	)
}

// GetGitCredentialByName retrieves a single GitCredential object owned by a specific user with a unique git credential name
func (service ServiceTx) GetGitCredentialByName(userID portaineree.UserID, name string) (*portaineree.GitCredential, error) {
	var credential portaineree.GitCredential

	err := service.tx.GetAll(
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

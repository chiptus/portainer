package gitcredential

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.GitCredential, portaineree.GitCredentialID]
}

// Create creates a new git credential object.
func (service ServiceTx) Create(record *portaineree.GitCredential) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			record.ID = portaineree.GitCredentialID(id)

			return int(record.ID), record
		},
	)
}

// GetGitCredentialsByUserID returns an array containing all git-credentials owned by a specific user
func (service ServiceTx) GetGitCredentialsByUserID(userID portainer.UserID) ([]portaineree.GitCredential, error) {
	var result = make([]portaineree.GitCredential, 0)

	return result, service.Tx.GetAll(
		BucketName,
		&portaineree.GitCredential{},
		dataservices.FilterFn(&result, func(e portaineree.GitCredential) bool {
			return e.UserID == userID
		}),
	)
}

// GetGitCredentialByName retrieves a single GitCredential object owned by a specific user with a unique git credential name
func (service ServiceTx) GetGitCredentialByName(userID portainer.UserID, name string) (*portaineree.GitCredential, error) {
	var credential portaineree.GitCredential

	err := service.Tx.GetAll(
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

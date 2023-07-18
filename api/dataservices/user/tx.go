package user

import (
	"errors"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.User, portaineree.UserID]
}

// UserByUsername returns a user by username.
func (service ServiceTx) UserByUsername(username string) (*portaineree.User, error) {
	var u portaineree.User

	err := service.Tx.GetAll(
		BucketName,
		&portaineree.User{},
		dataservices.FirstFn(&u, func(e portaineree.User) bool {
			return strings.EqualFold(e.Username, username)
		}),
	)

	if errors.Is(err, dataservices.ErrStop) {
		return &u, nil
	}

	if err == nil {
		return nil, dserrors.ErrObjectNotFound
	}

	return nil, err
}

// UsersByRole return an array containing all the users with the specified role.
func (service ServiceTx) UsersByRole(role portaineree.UserRole) ([]portaineree.User, error) {
	var users = make([]portaineree.User, 0)

	return users, service.Tx.GetAll(
		BucketName,
		&portaineree.User{},
		dataservices.FilterFn(&users, func(e portaineree.User) bool {
			return e.Role == role
		}),
	)
}

// CreateUser creates a new user.
func (service ServiceTx) Create(user *portaineree.User) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			user.ID = portaineree.UserID(id)
			user.Username = strings.ToLower(user.Username)

			return int(user.ID), user
		},
	)
}

package user

import (
	"errors"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "users"

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
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

// User returns a user by ID
func (service *Service) User(ID portaineree.UserID) (*portaineree.User, error) {
	var user portaineree.User
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UserByUsername returns a user by username.
func (service *Service) UserByUsername(username string) (*portaineree.User, error) {
	var u portaineree.User

	err := service.connection.GetAll(
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

// Users return an array containing all the users.
func (service *Service) Users() ([]portaineree.User, error) {
	var users = make([]portaineree.User, 0)

	return users, service.connection.GetAll(
		BucketName,
		&portaineree.User{},
		dataservices.AppendFn(&users),
	)
}

// UsersByRole return an array containing all the users with the specified role.
func (service *Service) UsersByRole(role portaineree.UserRole) ([]portaineree.User, error) {
	var users = make([]portaineree.User, 0)

	return users, service.connection.GetAll(
		BucketName,
		&portaineree.User{},
		dataservices.FilterFn(&users, func(e portaineree.User) bool {
			return e.Role == role
		}),
	)
}

// UpdateUser saves a user.
func (service *Service) UpdateUser(ID portaineree.UserID, user *portaineree.User) error {
	identifier := service.connection.ConvertToKey(int(ID))
	user.Username = strings.ToLower(user.Username)
	return service.connection.UpdateObject(BucketName, identifier, user)
}

// CreateUser creates a new user.
func (service *Service) Create(user *portaineree.User) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			user.ID = portaineree.UserID(id)
			user.Username = strings.ToLower(user.Username)

			return int(user.ID), user
		},
	)
}

// DeleteUser deletes a user.
func (service *Service) DeleteUser(ID portaineree.UserID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

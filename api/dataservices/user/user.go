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
	dataservices.BaseDataService[portaineree.User, portainer.UserID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.User, portainer.UserID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.User, portainer.UserID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// UserByUsername returns a user by username.
func (service *Service) UserByUsername(username string) (*portaineree.User, error) {
	var u portaineree.User

	err := service.Connection.GetAll(
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
func (service *Service) UsersByRole(role portainer.UserRole) ([]portaineree.User, error) {
	var users = make([]portaineree.User, 0)

	return users, service.Connection.GetAll(
		BucketName,
		&portaineree.User{},
		dataservices.FilterFn(&users, func(e portaineree.User) bool {
			return e.Role == role
		}),
	)
}

// CreateUser creates a new user.
func (service *Service) Create(user *portaineree.User) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			user.ID = portainer.UserID(id)
			user.Username = strings.ToLower(user.Username)

			return int(user.ID), user
		},
	)
}

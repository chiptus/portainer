package team

import (
	"errors"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "teams"

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	dataservices.BaseDataService[portaineree.Team, portaineree.TeamID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.Team, portaineree.TeamID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

// TeamByName returns a team by name.
func (service *Service) TeamByName(name string) (*portaineree.Team, error) {
	var t portaineree.Team

	err := service.Connection.GetAll(
		BucketName,
		&portaineree.Team{},
		dataservices.FirstFn(&t, func(e portaineree.Team) bool {
			return strings.EqualFold(e.Name, name)
		}),
	)

	if errors.Is(err, dataservices.ErrStop) {
		return &t, nil
	}

	if err == nil {
		return nil, dserrors.ErrObjectNotFound
	}

	return nil, err
}

// CreateTeam creates a new Team.
func (service *Service) Create(team *portaineree.Team) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			team.ID = portaineree.TeamID(id)
			return int(team.ID), team
		},
	)
}

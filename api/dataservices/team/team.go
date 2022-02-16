package team

import (
	"fmt"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/dataservices/errors"
	"github.com/sirupsen/logrus"

	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "teams"
)

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

// Team returns a Team by ID
func (service *Service) Team(ID portaineree.TeamID) (*portaineree.Team, error) {
	var team portaineree.Team
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &team)
	if err != nil {
		return nil, err
	}

	return &team, nil
}

// TeamByName returns a team by name.
func (service *Service) TeamByName(name string) (*portaineree.Team, error) {
	var t *portaineree.Team

	stop := fmt.Errorf("ok")
	err := service.connection.GetAll(
		BucketName,
		&portaineree.Team{},
		func(obj interface{}) (interface{}, error) {
			team, ok := obj.(*portaineree.Team)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to Team object")
				return nil, fmt.Errorf("Failed to convert to Team object: %s", obj)
			}
			if strings.EqualFold(team.Name, name) {
				t = team
				return nil, stop
			}
			return &portaineree.Team{}, nil
		})
	if err == stop {
		return t, nil
	}
	if err == nil {
		return nil, errors.ErrObjectNotFound
	}

	return nil, err
}

// Teams return an array containing all the teams.
func (service *Service) Teams() ([]portaineree.Team, error) {
	var teams = make([]portaineree.Team, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Team{},
		func(obj interface{}) (interface{}, error) {
			team, ok := obj.(*portaineree.Team)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to Team object")
				return nil, fmt.Errorf("Failed to convert to Team object: %s", obj)
			}
			teams = append(teams, *team)
			return &portaineree.Team{}, nil
		})

	return teams, err
}

// UpdateTeam saves a Team.
func (service *Service) UpdateTeam(ID portaineree.TeamID, team *portaineree.Team) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, team)
}

// CreateTeam creates a new Team.
func (service *Service) Create(team *portaineree.Team) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			team.ID = portaineree.TeamID(id)
			return int(team.ID), team
		},
	)
}

// DeleteTeam deletes a Team.
func (service *Service) DeleteTeam(ID portaineree.TeamID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

package team

import (
	"strings"

	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/errors"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "teams"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection *internal.DbConnection
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
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
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &team)
	if err != nil {
		return nil, err
	}

	return &team, nil
}

// TeamByName returns a team by name.
func (service *Service) TeamByName(name string) (*portaineree.Team, error) {
	var team *portaineree.Team

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var t portaineree.Team
			err := internal.UnmarshalObject(v, &t)
			if err != nil {
				return err
			}

			if strings.EqualFold(t.Name, name) {
				team = &t
				break
			}
		}

		if team == nil {
			return errors.ErrObjectNotFound
		}

		return nil
	})

	return team, err
}

// Teams return an array containing all the teams.
func (service *Service) Teams() ([]portaineree.Team, error) {
	var teams = make([]portaineree.Team, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var team portaineree.Team
			err := internal.UnmarshalObject(v, &team)
			if err != nil {
				return err
			}
			teams = append(teams, team)
		}

		return nil
	})

	return teams, err
}

// UpdateTeam saves a Team.
func (service *Service) UpdateTeam(ID portaineree.TeamID, team *portaineree.Team) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, team)
}

// CreateTeam creates a new Team.
func (service *Service) CreateTeam(team *portaineree.Team) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		team.ID = portaineree.TeamID(id)

		data, err := internal.MarshalObject(team)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(team.ID)), data)
	})
}

// DeleteTeam deletes a Team.
func (service *Service) DeleteTeam(ID portaineree.TeamID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}

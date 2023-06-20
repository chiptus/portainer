package teammembership

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "team_membership"

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

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

// TeamMembership returns a TeamMembership object by ID
func (service *Service) TeamMembership(ID portaineree.TeamMembershipID) (*portaineree.TeamMembership, error) {
	var membership portaineree.TeamMembership
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &membership)
	if err != nil {
		return nil, err
	}

	return &membership, nil
}

// TeamMemberships return an array containing all the TeamMembership objects.
func (service *Service) TeamMemberships() ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	return memberships, service.connection.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		dataservices.AppendFn(&memberships),
	)
}

// TeamMembershipsByUserID return an array containing all the TeamMembership objects where the specified userID is present.
func (service *Service) TeamMembershipsByUserID(userID portaineree.UserID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	return memberships, service.connection.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		dataservices.FilterFn(&memberships, func(e portaineree.TeamMembership) bool {
			return e.UserID == userID
		}),
	)
}

// TeamMembershipsByTeamID return an array containing all the TeamMembership objects where the specified teamID is present.
func (service *Service) TeamMembershipsByTeamID(teamID portaineree.TeamID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	return memberships, service.connection.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		dataservices.FilterFn(&memberships, func(e portaineree.TeamMembership) bool {
			return e.TeamID == teamID
		}),
	)
}

// UpdateTeamMembership saves a TeamMembership object.
func (service *Service) UpdateTeamMembership(ID portaineree.TeamMembershipID, membership *portaineree.TeamMembership) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, membership)
}

// CreateTeamMembership creates a new TeamMembership object.
func (service *Service) Create(membership *portaineree.TeamMembership) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			membership.ID = portaineree.TeamMembershipID(id)
			return int(membership.ID), membership
		},
	)
}

// DeleteTeamMembership deletes a TeamMembership object.
func (service *Service) DeleteTeamMembership(ID portaineree.TeamMembershipID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// DeleteTeamMembershipByUserID deletes all the TeamMembership object associated to a UserID.
func (service *Service) DeleteTeamMembershipByUserID(userID portaineree.UserID) error {
	return service.connection.DeleteAllObjects(
		BucketName,
		&portaineree.TeamMembership{},
		func(obj interface{}) (id int, ok bool) {
			membership, ok := obj.(*portaineree.TeamMembership)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to TeamMembership object")
				//return fmt.Errorf("Failed to convert to TeamMembership object: %s", obj)
				return -1, false
			}

			if membership.UserID == userID {
				return int(membership.ID), true
			}

			return -1, false
		})
}

// DeleteTeamMembershipByTeamID deletes all the TeamMembership object associated to a TeamID.
func (service *Service) DeleteTeamMembershipByTeamID(teamID portaineree.TeamID) error {
	return service.connection.DeleteAllObjects(
		BucketName,
		&portaineree.TeamMembership{},
		func(obj interface{}) (id int, ok bool) {
			membership, ok := obj.(*portaineree.TeamMembership)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to TeamMembership object")
				//return fmt.Errorf("Failed to convert to TeamMembership object: %s", obj)
				return -1, false
			}

			if membership.TeamID == teamID {
				return int(membership.ID), true
			}

			return -1, false
		})
}

func (service *Service) DeleteTeamMembershipByTeamIDAndUserID(teamID portaineree.TeamID, userID portaineree.UserID) error {
	return service.connection.DeleteAllObjects(
		BucketName,
		&portaineree.TeamMembership{},
		func(obj interface{}) (id int, ok bool) {
			membership, ok := obj.(*portaineree.TeamMembership)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to TeamMembership object")
				return -1, false
			}

			if membership.TeamID == teamID && membership.UserID == userID {
				return int(membership.ID), true
			}

			return -1, false
		})
}

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
	dataservices.BaseDataService[portaineree.TeamMembership, portaineree.TeamMembershipID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.TeamMembership, portaineree.TeamMembershipID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.TeamMembership, portaineree.TeamMembershipID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// TeamMembershipsByUserID return an array containing all the TeamMembership objects where the specified userID is present.
func (service *Service) TeamMembershipsByUserID(userID portaineree.UserID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	return memberships, service.Connection.GetAll(
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

	return memberships, service.Connection.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		dataservices.FilterFn(&memberships, func(e portaineree.TeamMembership) bool {
			return e.TeamID == teamID
		}),
	)
}

// CreateTeamMembership creates a new TeamMembership object.
func (service *Service) Create(membership *portaineree.TeamMembership) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			membership.ID = portaineree.TeamMembershipID(id)
			return int(membership.ID), membership
		},
	)
}

// DeleteTeamMembershipByUserID deletes all the TeamMembership object associated to a UserID.
func (service *Service) DeleteTeamMembershipByUserID(userID portaineree.UserID) error {
	return service.Connection.DeleteAllObjects(
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
	return service.Connection.DeleteAllObjects(
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
	return service.Connection.DeleteAllObjects(
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

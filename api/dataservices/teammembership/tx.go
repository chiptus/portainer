package teammembership

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// TeamMembership returns a TeamMembership object by ID
func (service ServiceTx) TeamMembership(ID portaineree.TeamMembershipID) (*portaineree.TeamMembership, error) {
	var membership portaineree.TeamMembership
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &membership)
	if err != nil {
		return nil, err
	}

	return &membership, nil
}

// TeamMemberships return an array containing all the TeamMembership objects.
func (service ServiceTx) TeamMemberships() ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		func(obj interface{}) (interface{}, error) {
			membership, ok := obj.(*portaineree.TeamMembership)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to TeamMembership object")
				return nil, fmt.Errorf("Failed to convert to TeamMembership object: %s", obj)
			}

			memberships = append(memberships, *membership)

			return &portaineree.TeamMembership{}, nil
		})

	return memberships, err
}

// TeamMembershipsByUserID return an array containing all the TeamMembership objects where the specified userID is present.
func (service ServiceTx) TeamMembershipsByUserID(userID portaineree.UserID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		func(obj interface{}) (interface{}, error) {
			membership, ok := obj.(*portaineree.TeamMembership)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to TeamMembership object")
				return nil, fmt.Errorf("Failed to convert to TeamMembership object: %s", obj)
			}

			if membership.UserID == userID {
				memberships = append(memberships, *membership)
			}

			return &portaineree.TeamMembership{}, nil
		})

	return memberships, err
}

// TeamMembershipsByTeamID return an array containing all the TeamMembership objects where the specified teamID is present.
func (service ServiceTx) TeamMembershipsByTeamID(teamID portaineree.TeamID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		func(obj interface{}) (interface{}, error) {
			membership, ok := obj.(*portaineree.TeamMembership)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to TeamMembership object")
				return nil, fmt.Errorf("Failed to convert to TeamMembership object: %s", obj)
			}

			if membership.TeamID == teamID {
				memberships = append(memberships, *membership)
			}

			return &portaineree.TeamMembership{}, nil
		})

	return memberships, err
}

// UpdateTeamMembership saves a TeamMembership object.
func (service ServiceTx) UpdateTeamMembership(ID portaineree.TeamMembershipID, membership *portaineree.TeamMembership) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, membership)
}

// CreateTeamMembership creates a new TeamMembership object.
func (service ServiceTx) Create(membership *portaineree.TeamMembership) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			membership.ID = portaineree.TeamMembershipID(id)
			return int(membership.ID), membership
		},
	)
}

// DeleteTeamMembership deletes a TeamMembership object.
func (service ServiceTx) DeleteTeamMembership(ID portaineree.TeamMembershipID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}

// DeleteTeamMembershipByUserID deletes all the TeamMembership object associated to a UserID.
func (service ServiceTx) DeleteTeamMembershipByUserID(userID portaineree.UserID) error {
	return service.tx.DeleteAllObjects(
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
func (service ServiceTx) DeleteTeamMembershipByTeamID(teamID portaineree.TeamID) error {
	return service.tx.DeleteAllObjects(
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

func (service ServiceTx) DeleteTeamMembershipByTeamIDAndUserID(teamID portaineree.TeamID, userID portaineree.UserID) error {
	return service.tx.DeleteAllObjects(
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

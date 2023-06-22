package teammembership

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"

	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.TeamMembership, portaineree.TeamMembershipID]
}

// TeamMembershipsByUserID return an array containing all the TeamMembership objects where the specified userID is present.
func (service ServiceTx) TeamMembershipsByUserID(userID portaineree.UserID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	return memberships, service.Tx.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		dataservices.FilterFn(&memberships, func(e portaineree.TeamMembership) bool {
			return e.UserID == userID
		}),
	)
}

// TeamMembershipsByTeamID return an array containing all the TeamMembership objects where the specified teamID is present.
func (service ServiceTx) TeamMembershipsByTeamID(teamID portaineree.TeamID) ([]portaineree.TeamMembership, error) {
	var memberships = make([]portaineree.TeamMembership, 0)

	return memberships, service.Tx.GetAll(
		BucketName,
		&portaineree.TeamMembership{},
		dataservices.FilterFn(&memberships, func(e portaineree.TeamMembership) bool {
			return e.TeamID == teamID
		}),
	)
}

// CreateTeamMembership creates a new TeamMembership object.
func (service ServiceTx) Create(membership *portaineree.TeamMembership) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			membership.ID = portaineree.TeamMembershipID(id)
			return int(membership.ID), membership
		},
	)
}

// DeleteTeamMembershipByUserID deletes all the TeamMembership object associated to a UserID.
func (service ServiceTx) DeleteTeamMembershipByUserID(userID portaineree.UserID) error {
	return service.Tx.DeleteAllObjects(
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
	return service.Tx.DeleteAllObjects(
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
	return service.Tx.DeleteAllObjects(
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

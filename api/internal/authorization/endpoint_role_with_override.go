package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// CleanNAPWithOverridePolicies Clean Namespace Access Policies with override policies
func (service *Service) CleanNAPWithOverridePolicies(
	tx dataservices.DataStoreTx,
	endpoint *portaineree.Endpoint,
	endpointGroup *portainer.EndpointGroup,
) error {
	kubecli, err := service.K8sClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return err
	}

	accessPolicies, err := kubecli.GetNamespaceAccessPolicies()
	if err != nil {
		return err
	}

	hasChange := false

	for namespace, policy := range accessPolicies {
		for teamID := range policy.TeamAccessPolicies {
			endpointRole, err := service.GetTeamEndpointRoleWithOverridePolicies(tx, teamID, endpoint, endpointGroup)
			if err != nil {
				return err
			}

			if endpointRole == nil {
				delete(accessPolicies[namespace].TeamAccessPolicies, teamID)
				hasChange = true
			}
		}

		for userID := range policy.UserAccessPolicies {
			_, err := tx.User().Read(userID)
			if tx.IsErrObjectNotFound(err) {
				continue
			}

			endpointRole, err := service.GetUserEndpointRoleWithOverridePolicies(tx, userID, endpoint, endpointGroup)
			if err != nil {
				return err
			}

			if endpointRole == nil {
				delete(accessPolicies[namespace].UserAccessPolicies, userID)
				hasChange = true
			}
		}
	}

	if hasChange {
		return kubecli.UpdateNamespaceAccessPolicies(accessPolicies)
	}

	return nil
}

func (service *Service) GetUserEndpointRoleWithOverridePolicies(
	tx dataservices.DataStoreTx,
	userID portainer.UserID,
	endpoint *portaineree.Endpoint,
	endpointGroup *portainer.EndpointGroup,
) (*portaineree.Role, error) {
	user, err := tx.User().Read(portainer.UserID(userID))
	if err != nil {
		return nil, err
	}

	userMemberships, err := tx.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	endpointGroups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return nil, err
	}

	roles, err := tx.Role().ReadAll()
	if err != nil {
		return nil, err
	}

	groupUserAccessPolicies, groupTeamAccessPolicies := getGroupPolicies(endpointGroups)

	if endpointGroup != nil {
		groupUserAccessPolicies[endpointGroup.ID] = endpointGroup.UserAccessPolicies
		groupTeamAccessPolicies[endpointGroup.ID] = endpointGroup.TeamAccessPolicies
	}

	return getUserEndpointRole(user, *endpoint, groupUserAccessPolicies, groupTeamAccessPolicies, roles, userMemberships), nil
}

func (service *Service) GetTeamEndpointRoleWithOverridePolicies(
	tx dataservices.DataStoreTx,
	teamID portainer.TeamID,
	endpoint *portaineree.Endpoint,
	endpointGroup *portainer.EndpointGroup,
) (*portaineree.Role, error) {
	memberships, err := tx.TeamMembership().TeamMembershipsByTeamID(teamID)
	if err != nil {
		return nil, err
	}

	endpointGroups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return nil, err
	}

	roles, err := tx.Role().ReadAll()
	if err != nil {
		return nil, err
	}

	_, groupTeamAccessPolicies := getGroupPolicies(endpointGroups)

	if endpointGroup != nil {
		groupTeamAccessPolicies[endpointGroup.ID] = endpointGroup.TeamAccessPolicies
	}

	role := getRoleFromTeamAccessPolicies(memberships, endpoint.TeamAccessPolicies, roles)
	if role != nil {
		return role, nil
	}

	return getRoleFromTeamEndpointGroupPolicies(memberships, endpoint, roles, groupTeamAccessPolicies), nil
}

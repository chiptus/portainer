package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// CleanNAPWithOverridePolicies Clean Namespace Access Policies with override policies
func (service *Service) CleanNAPWithOverridePolicies(
	endpoint *portaineree.Endpoint,
	endpointGroup *portaineree.EndpointGroup,
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
			endpointRole, err := service.GetTeamEndpointRoleWithOverridePolicies(teamID, endpoint, endpointGroup)
			if err != nil {
				return err
			}
			if endpointRole == nil {
				delete(accessPolicies[namespace].TeamAccessPolicies, teamID)
				hasChange = true
			}
		}

		for userID := range policy.UserAccessPolicies {
			_, err := service.dataStore.User().User(userID)
			if service.dataStore.IsErrObjectNotFound(err) {
				continue
			}

			endpointRole, err := service.GetUserEndpointRoleWithOverridePolicies(userID, endpoint, endpointGroup)
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
		err = kubecli.UpdateNamespaceAccessPolicies(accessPolicies)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *Service) GetUserEndpointRoleWithOverridePolicies(
	userID portaineree.UserID,
	endpoint *portaineree.Endpoint,
	endpointGroup *portaineree.EndpointGroup,
) (*portaineree.Role, error) {
	user, err := service.dataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		return nil, err
	}

	userMemberships, err := service.dataStore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return nil, err
	}

	roles, err := service.dataStore.Role().Roles()
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
	teamID portaineree.TeamID,
	endpoint *portaineree.Endpoint,
	endpointGroup *portaineree.EndpointGroup,
) (*portaineree.Role, error) {

	memberships, err := service.dataStore.TeamMembership().TeamMembershipsByTeamID(teamID)
	if err != nil {
		return nil, err
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return nil, err
	}

	roles, err := service.dataStore.Role().Roles()
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

	role = getRoleFromTeamEndpointGroupPolicies(memberships, endpoint, roles, groupTeamAccessPolicies)
	return role, nil
}

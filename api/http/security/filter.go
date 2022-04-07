package security

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// FilterUserTeams filters teams based on user role.
// non-administrator users only have access to team they are member of.
func FilterUserTeams(teams []portaineree.Team, context *RestrictedRequestContext) []portaineree.Team {
	filteredTeams := teams

	if !context.IsAdmin {
		filteredTeams = make([]portaineree.Team, 0)
		for _, membership := range context.UserMemberships {
			for _, team := range teams {
				if team.ID == membership.TeamID {
					filteredTeams = append(filteredTeams, team)
					break
				}
			}
		}
	}

	return filteredTeams
}

// FilterLeaderTeams filters teams based on user role.
// Team leaders only have access to team they lead.
func FilterLeaderTeams(teams []portaineree.Team, context *RestrictedRequestContext) []portaineree.Team {
	filteredTeams := teams

	if context.IsTeamLeader {
		filteredTeams = make([]portaineree.Team, 0)
		for _, membership := range context.UserMemberships {
			for _, team := range teams {
				if team.ID == membership.TeamID && membership.Role == portaineree.TeamLeader {
					filteredTeams = append(filteredTeams, team)
					break
				}
			}
		}
	}

	return filteredTeams
}

// FilterUsers filters users based on user role.
// Non-administrator users only have access to non-administrator users.
func FilterUsers(users []portaineree.User, context *RestrictedRequestContext) []portaineree.User {
	filteredUsers := users

	if !context.IsAdmin {
		filteredUsers = make([]portaineree.User, 0)

		for _, user := range users {
			if user.Role != portaineree.AdministratorRole {
				filteredUsers = append(filteredUsers, user)
			}
		}
	}

	return filteredUsers
}

// FilterRegistries filters registries based on user role and team memberships.
// Non administrator users only have access to authorized registries.
func FilterRegistries(registries []portaineree.Registry, user *portaineree.User, teamMemberships []portaineree.TeamMembership, endpointID portaineree.EndpointID) []portaineree.Registry {
	if user.Role == portaineree.AdministratorRole {
		return registries
	}

	filteredRegistries := []portaineree.Registry{}
	for _, registry := range registries {
		if AuthorizedRegistryAccess(&registry, user, teamMemberships, endpointID) {
			filteredRegistries = append(filteredRegistries, registry)
		}
	}
	return filteredRegistries
}

// FilterEndpoints filters environments(endpoints) based on user role and team memberships.
// Non administrator users only have access to authorized environments(endpoints) (can be inherited via endpoint groups).
func FilterEndpoints(endpoints []portaineree.Endpoint, groups []portaineree.EndpointGroup, context *RestrictedRequestContext) []portaineree.Endpoint {
	filteredEndpoints := endpoints

	if !context.IsAdmin {
		filteredEndpoints = make([]portaineree.Endpoint, 0)

		for _, endpoint := range endpoints {
			endpointGroup := getAssociatedGroup(&endpoint, groups)

			if authorizedEndpointAccess(&endpoint, endpointGroup, context.UserID, context.UserMemberships) {
				filteredEndpoints = append(filteredEndpoints, endpoint)
			}
		}
	}

	return filteredEndpoints
}

// FilterEndpointGroups filters environment(endpoint) groups based on user role and team memberships.
// Non administrator users only have access to authorized environment(endpoint) groups.
func FilterEndpointGroups(endpointGroups []portaineree.EndpointGroup, context *RestrictedRequestContext) []portaineree.EndpointGroup {
	filteredEndpointGroups := endpointGroups

	if !context.IsAdmin {
		filteredEndpointGroups = make([]portaineree.EndpointGroup, 0)

		for _, group := range endpointGroups {
			if authorizedEndpointGroupAccess(&group, context.UserID, context.UserMemberships) {
				filteredEndpointGroups = append(filteredEndpointGroups, group)
			}
		}
	}

	return filteredEndpointGroups
}

func getAssociatedGroup(endpoint *portaineree.Endpoint, groups []portaineree.EndpointGroup) *portaineree.EndpointGroup {
	for _, group := range groups {
		if group.ID == endpoint.GroupID {
			return &group
		}
	}
	return nil
}

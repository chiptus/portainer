package auth

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

// removeMemberships removes a user's team membership(s) if user does not belong to it/them anymore
func removeMemberships(tms dataservices.TeamMembershipService, user portaineree.User, teams []portaineree.Team) error {
	log.Println("[DEBUG] [internal,oauth] [message: removing user team memberships which no longer exist]")
	memberships, err := tms.TeamMembershipsByUserID(user.ID)
	if err != nil {
		return err
	}

	for _, membership := range memberships {
		teamsContainsTeamID := false
		for _, team := range teams {
			if team.ID == membership.TeamID {
				teamsContainsTeamID = true
				break
			}
		}

		if !teamsContainsTeamID {
			err := tms.DeleteTeamMembership(membership.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// createOrUpdateMembership creates a membership if it does not exist or updates a memberships role (if already existent)
func createOrUpdateMembership(tms dataservices.TeamMembershipService, user portaineree.User, team portaineree.Team) error {
	memberships, err := tms.TeamMembershipsByTeamID(team.ID)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] [internal,oauth] [message: memberships: %v]", memberships)

	var membership *portaineree.TeamMembership
	for _, m := range memberships {
		if m.UserID == user.ID {
			membership = &m
			break
		}
	}

	if membership == nil {
		membership = &portaineree.TeamMembership{
			UserID: user.ID,
			TeamID: team.ID,
			Role:   portaineree.MembershipRole(user.Role),
		}
		log.Printf("[DEBUG] [internal,oauth] [message: creating oauth user team membership: %v]", membership)
		err = tms.Create(membership)
		if err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] [internal,oauth] [message: membership found %v]", membership)
		if updatedRole := portaineree.MembershipRole(user.Role); membership.Role != updatedRole {
			log.Printf("[DEBUG] [internal,oauth] [message: updating membership role %d]", updatedRole)
			membership.Role = updatedRole
			err = tms.UpdateTeamMembership(membership.ID, membership)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// mapAllClaimValuesToTeams maps claim values to teams if no explicit mapping exists.
// Mapping oauth teams (claim values) to portainer teams by case-insensitive team name
func mapAllClaimValuesToTeams(ts dataservices.TeamService, user portaineree.User, oAuthTeams []string) ([]portaineree.Team, error) {
	teams := make([]portaineree.Team, 0)

	log.Println("[DEBUG] [internal,oauth] [message: mapping oauth claim values automatically to existing portainer teams]")
	dsTeams, err := ts.Teams()
	if err != nil {
		return []portaineree.Team{}, err
	}

	for _, oAuthTeam := range oAuthTeams {
		for _, team := range dsTeams {
			if strings.EqualFold(team.Name, oAuthTeam) {
				teams = append(teams, team)
			}
		}
	}

	return teams, nil
}

// mapClaimValRegexToTeams maps oauth ClaimValRegex values (stored in settings) to oauth provider teams.
// The `ClaimValRegex` is a regexp string that is matched against the oauth team value(s) extracted from oauth user response.
// A successful match entails extraction of the respective portainer team (for the mapping).
func mapClaimValRegexToTeams(ts dataservices.TeamService, claimMappings []portaineree.OAuthClaimMappings, oAuthTeams []string) ([]portaineree.Team, error) {
	teams := make([]portaineree.Team, 0)

	log.Println("[DEBUG] [internal,oauth] [message: using oauth claim mappings to map groups to portainer teams]")
	for _, oAuthTeam := range oAuthTeams {
		for _, mapping := range claimMappings {
			match, err := regexp.MatchString(mapping.ClaimValRegex, oAuthTeam)
			if err != nil {
				return nil, err
			}

			if match {
				log.Printf("[DEBUG] [internal,oauth] [message: oauth mapping claim matched; claim: %s, team: %s]\n", mapping.ClaimValRegex, oAuthTeam)

				team, err := ts.Team(portaineree.TeamID(mapping.Team))
				if err != nil {
					return nil, err
				}

				teams = append(teams, *team)
			}
		}
	}

	return teams, nil
}

// updateOAuthTeamMemberships will create, update and delete an oauth user's team memberships.
// The mappings of oauth groups to portainer teams is based on the length of `OAuthClaimMappings`; use them if they exist (len > 0),
// otherwise map the **values** of the oauth `Claim name` (`OAuthClaimName`) to already existent portainer teams (case-insensitive).
func updateOAuthTeamMemberships(dataStore dataservices.DataStore, oAuthSettings portaineree.OAuthSettings, user portaineree.User, oAuthTeams []string) error {
	var teams []portaineree.Team
	var err error
	oAuthClaimMappings := oAuthSettings.TeamMemberships.OAuthClaimMappings

	if len(oAuthClaimMappings) > 0 {
		teams, err = mapClaimValRegexToTeams(dataStore.Team(), oAuthClaimMappings, oAuthTeams)
		if err != nil {
			return fmt.Errorf("failed to map claim value regex(s) to teams, mappings: %v, err: %w", oAuthClaimMappings, err)
		}
	} else {
		teams, err = mapAllClaimValuesToTeams(dataStore.Team(), user, oAuthTeams)
		if err != nil {
			return fmt.Errorf("failed to map claim value(s) to portainer teams, err: %w", err)
		}
	}

	// if user cannot be assigned to any teams based on claims, then assign user to the default team
	if len(teams) == 0 && oAuthSettings.DefaultTeamID != 0 {
		defaultTeam, err := dataStore.Team().Team(oAuthSettings.DefaultTeamID)
		if err != nil {
			return fmt.Errorf("failed to retrieve default portainer team, err: %w", err)
		}
		teams = append(teams, *defaultTeam)
	}

	for _, team := range teams {
		err := createOrUpdateMembership(dataStore.TeamMembership(), user, team)
		if err != nil {
			return fmt.Errorf("failed to create or update oauth memberships, user: %v, team: %v, err: %w", user, team, err)
		}
	}

	err = removeMemberships(dataStore.TeamMembership(), user, teams)
	if err != nil {
		return fmt.Errorf("failed to remove oauth memberships, user: %v, err: %w", user, err)
	}

	return nil
}

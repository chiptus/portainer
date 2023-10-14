package auth

import (
	"fmt"
	"regexp"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

// removeMemberships removes a user's team membership(s) if user does not belong to it/them anymore
func removeMemberships(tms dataservices.TeamMembershipService, user portaineree.User, teams []portainer.Team) error {
	log.Debug().Msg("removing user team memberships which no longer exist")
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
			err := tms.Delete(membership.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// createOrUpdateMembership creates a membership if it does not exist or updates a memberships role (if already existent)
func createOrUpdateMembership(tms dataservices.TeamMembershipService, user portaineree.User, team portainer.Team) error {
	memberships, err := tms.TeamMembershipsByTeamID(team.ID)
	if err != nil {
		return err
	}

	log.Debug().Str("memberships", fmt.Sprintf("%+v", memberships)).Msg("")

	var membership *portainer.TeamMembership
	for _, m := range memberships {
		m := m
		if m.UserID == user.ID {
			membership = &m
			break
		}
	}

	if membership == nil {
		membership = &portainer.TeamMembership{
			UserID: user.ID,
			TeamID: team.ID,
			Role:   portainer.MembershipRole(user.Role),
		}

		log.Debug().Str("membership", fmt.Sprintf("%+v", membership)).Msg("creating OAuth user team membership")

		err = tms.Create(membership)
		if err != nil {
			return err
		}
	} else {
		log.Debug().Str("membership", fmt.Sprintf("%+v", membership)).Msg("membership found")

		if updatedRole := portainer.MembershipRole(user.Role); membership.Role != updatedRole {
			log.Debug().Int("updated_role", int(updatedRole)).Msg("updating membership role")

			membership.Role = updatedRole
			err = tms.Update(membership.ID, membership)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// mapAllClaimValuesToTeams maps claim values to teams if no explicit mapping exists.
// Mapping oauth teams (claim values) to portainer teams by case-insensitive team name
func mapAllClaimValuesToTeams(ts dataservices.TeamService, user portaineree.User, oAuthTeams []string) ([]portainer.Team, error) {
	teams := make([]portainer.Team, 0)

	log.Debug().Msg("mapping OAuth claim values automatically to existing portainer teams")
	dsTeams, err := ts.ReadAll()
	if err != nil {
		return []portainer.Team{}, err
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
func mapClaimValRegexToTeams(ts dataservices.TeamService, claimMappings []portaineree.OAuthClaimMappings, oAuthTeams []string) ([]portainer.Team, error) {
	teams := make([]portainer.Team, 0)

	log.Debug().Msg("using oauth claim mappings to map groups to portainer teams")

	for _, oAuthTeam := range oAuthTeams {
		for _, mapping := range claimMappings {
			match, err := regexp.MatchString(mapping.ClaimValRegex, oAuthTeam)
			if err != nil {
				return nil, err
			}

			if match {
				log.Debug().
					Str("claim", mapping.ClaimValRegex).
					Str("team", oAuthTeam).
					Msg("OAuth mapping claim matched")

				team, err := ts.Read(portainer.TeamID(mapping.Team))
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
	var teams []portainer.Team
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
		defaultTeam, err := dataStore.Team().Read(oAuthSettings.DefaultTeamID)
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

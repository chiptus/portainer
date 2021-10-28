package auth

import (
	"regexp"

	portainer "github.com/portainer/portainer/api"
)

func validateAdminClaims(oAuthSettings portainer.OAuthSettings, oAuthTeams []string) (bool, error) {
	for _, team := range oAuthTeams {
		for _, regex := range oAuthSettings.TeamMemberships.AdminGroupClaimsRegexList {
			match, err := regexp.MatchString(regex, team)
			if err != nil {
				return false, err
			} else if match {
				return true, nil
			}
		}
	}
	return false, nil
}

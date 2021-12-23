package auth

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
)

func Test_validateClaimWithRegex(t *testing.T) {
	t.Run("returns false if no regex match occurs", func(t *testing.T) {
		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				AdminGroupClaimsRegexList: []string{"@"},
			},
		}
		oAuthTeams := []string{"#portainer"}

		isValid, err := validateAdminClaims(oAuthSettings, oAuthTeams)
		if err != nil {
			t.Errorf("failed to validate, error: %v", err)
		}
		if isValid {
			t.Errorf("should be invalid when matching AdminGroupClaimsRegexList: %v and OAuth team: %v", oAuthSettings.TeamMemberships.AdminGroupClaimsRegexList, oAuthTeams)
		}
	})

	t.Run("returns true if regex match - single element in AdminGroupClaimsRegexList", func(t *testing.T) {
		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				AdminGroupClaimsRegexList: []string{"@"},
			},
		}
		oAuthTeams := []string{"@portainer"}

		isValid, err := validateAdminClaims(oAuthSettings, oAuthTeams)
		if err != nil {
			t.Errorf("failed to validate, error: %v", err)
		}
		if !isValid {
			t.Errorf("should be valid when matching AdminGroupClaimsRegexList: %v and OAuth team: %v", oAuthSettings.TeamMemberships.AdminGroupClaimsRegexList, oAuthTeams)
		}
	})

	t.Run("returns true if regex match - multiple elements in AdminGroupClaimsRegexList and oAuthTeams", func(t *testing.T) {
		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				AdminGroupClaimsRegexList: []string{"@", "#"},
			},
		}
		oAuthTeams := []string{"portainer", "@portainer"}

		isValid, err := validateAdminClaims(oAuthSettings, oAuthTeams)
		if err != nil {
			t.Errorf("failed to validate, error: %v", err)
		}
		if !isValid {
			t.Errorf("should be valid when matching AdminGroupClaimsRegexList: %v and OAuth team: %v", oAuthSettings.TeamMemberships.AdminGroupClaimsRegexList, oAuthTeams)
		}
	})
}

package auth

import (
	"reflect"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/bolttest"
)

func Test_mapClaimValRegexToTeams(t *testing.T) {
	t.Run("returns no portainer teams if no oauth teams are present", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		mappings := []portaineree.OAuthClaimMappings{}
		oAuthTeams := []string{}
		teams, _ := mapClaimValRegexToTeams(store.TeamService, mappings, oAuthTeams)
		if len(teams) > 0 {
			t.Errorf("mapClaimValRegexToTeams return no teams; teams returned %d", len(teams))
		}
	})

	t.Run("returns no portainer teams if no regex match occurs", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.TeamService.CreateTeam(&portaineree.Team{ID: 1, Name: "testing"})

		mappings := []portaineree.OAuthClaimMappings{
			{ClaimValRegex: "@", Team: 1},
		}
		oAuthTeams := []string{"portainer"}

		teams, _ := mapClaimValRegexToTeams(store.TeamService, mappings, oAuthTeams)
		if len(teams) > 0 {
			t.Errorf("mapClaimValRegexToTeams return no teams upon no regex match; teams returned %d", len(teams))
		}
	})

	t.Run("returns team upon regex match", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.TeamService.CreateTeam(&portaineree.Team{ID: 1, Name: "testing"})

		mappings := []portaineree.OAuthClaimMappings{
			{ClaimValRegex: "@", Team: 1},
		}
		oAuthTeams := []string{"@portainer"}

		got, _ := mapClaimValRegexToTeams(store.TeamService, mappings, oAuthTeams)
		want := []portaineree.Team{{ID: 1, Name: "testing"}}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("mapClaimValRegexToTeams failed to return team; got=%v, want=%v", got, want)
		}
	})

	t.Run("succcessfully fails to return non-existent team upon regex match", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		mappings := []portaineree.OAuthClaimMappings{
			{ClaimValRegex: "@", Team: 1337},
		}
		oAuthTeams := []string{"@portainer"}
		_, err = mapClaimValRegexToTeams(store.TeamService, mappings, oAuthTeams)
		if err == nil {
			t.Errorf("mapClaimValRegexToTeams should fail to return non-existent team")
		}
	})

}

func Test_mapAllClaimValuesToTeams(t *testing.T) {
	t.Parallel()
	store, teardown, err := bolttest.NewTestStore(false)
	if err != nil {
		t.Errorf("failed to initialise test store")
	}
	defer teardown()

	store.TeamService.CreateTeam(&portaineree.Team{ID: 1, Name: "team-x"})

	t.Run("returns no portainer teams if no oauth teams are present", func(t *testing.T) {
		oAuthTeams := []string{}
		user := portaineree.User{ID: 1, Role: 1}

		teams, _ := mapAllClaimValuesToTeams(store.TeamService, user, oAuthTeams)
		if len(teams) > 0 {
			t.Errorf("mapAllClaimValuesToTeams return no teams; teams returned %d", len(teams))
		}
	})

	t.Run("returns no portainer teams if no regex match occurs", func(t *testing.T) {
		oAuthTeams := []string{"team-1"}
		user := portaineree.User{ID: 1, Role: 1}

		teams, _ := mapAllClaimValuesToTeams(store.TeamService, user, oAuthTeams)
		if len(teams) > 0 {
			t.Errorf("mapAllClaimValuesToTeams return no teams upon no regex match; teams returned %d", len(teams))
		}
	})

	t.Run("returns team upon regex match", func(t *testing.T) {
		oAuthTeams := []string{"team-x"}
		user := portaineree.User{ID: 1, Role: 1}

		got, _ := mapAllClaimValuesToTeams(store.TeamService, user, oAuthTeams)
		want := []portaineree.Team{{ID: 1, Name: "team-x"}}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("mapAllClaimValuesToTeams failed to return team; got=%v, want=%v", got, want)
		}
	})

}

func Test_createOrUpdateMembership(t *testing.T) {
	t.Run("creates membership for new user-team", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		user := portaineree.User{ID: 1, Role: 1}
		team := portaineree.Team{ID: 1, Name: "team-1"}

		err = createOrUpdateMembership(store.TeamMembershipService, user, team)
		if err != nil {
			t.Errorf("createOrUpdateMembership should not throw error when creating new team membership")
		}

		got, _ := store.TeamMembershipService.TeamMemberships()
		want := []portaineree.TeamMembership{{ID: 1, UserID: 1, TeamID: 1, Role: 1}}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("createOrUpdateMembership should succeed in creating new team membership; got=%v, want=%v", got, want)
		}
	})

	t.Run("updates membership for existing user-team", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		user := portaineree.User{ID: 1, Role: 3}
		team := portaineree.Team{ID: 1, Name: "team-1"}
		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{ID: 1, UserID: user.ID, TeamID: team.ID, Role: 1})

		err = createOrUpdateMembership(store.TeamMembershipService, user, team)
		if err != nil {
			t.Errorf("createOrUpdateMembership should not throw error when updating existing team membership")
		}

		got, _ := store.TeamMembershipService.TeamMembership(1)
		want := &portaineree.TeamMembership{ID: 1, UserID: 1, TeamID: 1, Role: 3}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("createOrUpdateMembership should succeed in creating new team membership; got=%v, want=%v", got, want)
		}
	})

}

func Test_removeMemberships(t *testing.T) {
	t.Run("removes nothing if no user team memberships exist", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		user := portaineree.User{ID: 1, Role: 1}
		teams := []portaineree.Team{{ID: 1, Name: "team-remove"}}

		before, _ := store.TeamMembershipService.TeamMemberships()

		removeMemberships(store.TeamMembershipService, user, teams)

		after, _ := store.TeamMembershipService.TeamMemberships()

		if !reflect.DeepEqual(before, after) {
			t.Errorf("removeMemberships should not have removed any memberships; before=%v, after=%v", before, after)
		}
	})

	t.Run("does not remove user team membership if it does belong to team whitelist", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		user := portaineree.User{ID: 1, Role: 1}
		teams := []portaineree.Team{{ID: 1, Name: "team-remove"}}
		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{ID: 1, UserID: user.ID, TeamID: teams[0].ID, Role: 1})

		before, _ := store.TeamMembershipService.TeamMembership(1)

		removeMemberships(store.TeamMembershipService, user, teams)

		after, _ := store.TeamMembershipService.TeamMembership(1)

		if !reflect.DeepEqual(before, after) {
			t.Errorf("removeMemberships should not have removed any memberships; before=%v, after=%v", before, after)
		}
	})

	t.Run("removes memberships if user team membership does not belong to team whitelist", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		user := portaineree.User{ID: 1, Role: 1}
		teams := []portaineree.Team{{ID: 1, Name: "team-xyz"}}
		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{ID: 2, UserID: user.ID, TeamID: 100, Role: 1})
		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{ID: 3, UserID: user.ID, TeamID: 50, Role: 1})

		removeMemberships(store.TeamMembershipService, user, teams)

		memberships, _ := store.TeamMembershipService.TeamMembershipsByTeamID(100)
		if len(memberships) > 0 {
			t.Errorf("removeMemberships should have successfully removed team membership; team-membership=%v", memberships)
		}

		memberships, _ = store.TeamMembershipService.TeamMembershipsByTeamID(50)
		if len(memberships) > 0 {
			t.Errorf("removeMemberships should have successfully removed team membership; team-membership=%v", memberships)
		}
	})
}

func Test_updateOAuthTeamMemberships(t *testing.T) {
	t.Run("creates new team memberships based on claim val regex", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.Team().CreateTeam(&portaineree.Team{ID: 1, Name: "testing"})

		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{
					{ClaimValRegex: "@portainer", Team: 1},
				},
			},
		}
		user := portaineree.User{ID: 1, Role: 1}
		oAuthTeams := []string{"@portainer"}

		before, _ := store.TeamMembershipService.TeamMembershipsByTeamID(1)
		if len(before) > 0 {
			t.Errorf("updateOAuthTeamMemberships should not have a team membership with team id 1")
		}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		after, _ := store.TeamMembershipService.TeamMembershipsByTeamID(1)

		if reflect.DeepEqual(before, after) {
			t.Errorf("updateOAuthTeamMemberships should have created new team membership based on claim value regex")
		}
	})

	t.Run("fallsback to creating team memberships by mapping oauth teams directly to portainer teams", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.TeamService.CreateTeam(&portaineree.Team{ID: 1, Name: "testing"})
		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{},
			},
		}
		user := portaineree.User{ID: 1, Role: 1}
		oAuthTeams := []string{"testing"}

		before, _ := store.TeamMembershipService.TeamMembershipsByTeamID(1)
		if len(before) > 0 {
			t.Errorf("updateOAuthTeamMemberships should not have a team membership with team id 1")
		}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		after, _ := store.TeamMembershipService.TeamMembershipsByTeamID(1)

		if reflect.DeepEqual(before, after) {
			t.Errorf("updateOAuthTeamMemberships should have created new team membership based on existing portainer teams matching oauth team")
		}
	})

	t.Run("updates existing team membership based on claim val regex", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.Team().CreateTeam(&portaineree.Team{ID: 1, Name: "testing"})
		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{
			ID: 1, UserID: 1, TeamID: 1, Role: 1,
		})

		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{
					{ClaimValRegex: "@portainer", Team: 1},
				},
			},
		}
		user := portaineree.User{ID: 1, Role: 2}
		oAuthTeams := []string{"@portainer"}

		got, _ := store.TeamMembershipService.TeamMembershipsByUserID(1)
		want := []portaineree.TeamMembership{{UserID: 1, TeamID: 1, Role: 1}}
		if !(len(got) == 1 && got[0].Role == want[0].Role) {
			t.Errorf("updateOAuthTeamMemberships should have initial role of 1, got=%v, want=%v", got, want)
		}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		got, _ = store.TeamMembershipService.TeamMembershipsByUserID(1)
		want = []portaineree.TeamMembership{{ID: 1, UserID: 1, TeamID: 1, Role: 2}}
		if !(len(got) == 1 && got[0].Role == want[0].Role) {
			t.Errorf("updateOAuthTeamMemberships should have updated existing team membership role, got=%v, want=%v", got, want)
		}
	})

	t.Run("removes an outdated oauth team membership", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{
			ID: 1, UserID: 1, TeamID: 1, Role: 1,
		})

		oAuthSettings := portaineree.OAuthSettings{
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{},
			},
		}
		user := portaineree.User{ID: 1, Role: 1}
		oAuthTeams := []string{}

		got, _ := store.TeamMembershipService.TeamMembershipsByTeamID(1)
		if len(got) == 0 {
			t.Errorf("updateOAuthTeamMemberships should have initial team membership")
		}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		got, _ = store.TeamMembershipService.TeamMembershipsByTeamID(1)
		if len(got) > 0 {
			t.Errorf("updateOAuthTeamMemberships should have removed existing, non-mapped team membership")
		}
	})

	t.Run("assigns to default team if no mapping found", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.Team().CreateTeam(&portaineree.Team{ID: 1, Name: "testing"})

		oAuthSettings := portaineree.OAuthSettings{
			DefaultTeamID: 1,
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{},
			},
		}
		user := portaineree.User{ID: 1, Role: 1}
		oAuthTeams := []string{}

		got, _ := store.TeamMembershipService.TeamMembershipsByUserID(1)
		if len(got) > 0 {
			t.Errorf("updateOAuthTeamMemberships should not have team membership")
		}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		got, _ = store.TeamMembershipService.TeamMembershipsByUserID(1)
		if len(got) == 0 || got[0].TeamID != 1 {
			t.Errorf("updateOAuthTeamMemberships should have default membership")
		}
	})

	t.Run("removes to default team membership if no default team is set", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{
			ID: 1, UserID: 1, TeamID: 1, Role: 1,
		})

		oAuthSettings := portaineree.OAuthSettings{
			DefaultTeamID: 0,
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{},
			},
		}
		user := portaineree.User{ID: 1, Role: 1}
		oAuthTeams := []string{}

		got, _ := store.TeamMembershipService.TeamMembershipsByUserID(1)
		if len(got) == 0 {
			t.Errorf("updateOAuthTeamMemberships should have initial default membership")
		}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		got, _ = store.TeamMembershipService.TeamMembershipsByUserID(1)
		if len(got) > 0 {
			t.Errorf("updateOAuthTeamMemberships should have removed default team membership")
		}
	})

	t.Run("removes to default team membership if mapping is found", func(t *testing.T) {
		t.Parallel()
		store, teardown, err := bolttest.NewTestStore(false)
		if err != nil {
			t.Errorf("failed to initialise test store")
		}
		defer teardown()

		store.Team().CreateTeam(&portaineree.Team{ID: 1, Name: "test"})
		store.Team().CreateTeam(&portaineree.Team{ID: 2, Name: "@portainer"})
		store.TeamMembershipService.CreateTeamMembership(&portaineree.TeamMembership{
			UserID: 1, TeamID: 1, Role: 1,
		})

		got, _ := store.TeamMembershipService.TeamMembershipsByUserID(1)
		if !(len(got) == 1 && got[0].TeamID == 1) {
			t.Errorf("updateOAuthTeamMemberships should have initial default membership")
		}

		oAuthSettings := portaineree.OAuthSettings{
			DefaultTeamID: 0,
			TeamMemberships: portaineree.TeamMemberships{
				OAuthClaimMappings: []portaineree.OAuthClaimMappings{
					{ClaimValRegex: "@portainer", Team: 2},
				},
			},
		}
		user := portaineree.User{ID: 1, Role: 1}
		oAuthTeams := []string{"@portainer"}

		updateOAuthTeamMemberships(store, oAuthSettings, user, oAuthTeams)

		got, _ = store.TeamMembershipService.TeamMembershipsByUserID(1)
		if !(len(got) == 1 && got[0].TeamID == 2) {
			t.Errorf("updateOAuthTeamMemberships should have removed default team membership and created 'testing' team membership; got=%v", got)
		}
	})
}

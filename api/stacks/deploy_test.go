package stacks

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

type gitService struct {
	cloneErr error
	id       string
}

func (g *gitService) CloneRepository(destination, repositoryURL, referenceName, username, password string) error {
	return g.cloneErr
}

func (g *gitService) LatestCommitID(repositoryURL, referenceName, username, password string) (string, error) {
	return g.id, nil
}

type noopDeployer struct {
	SwarmStackDeployed      bool
	ComposeStackDeployed    bool
	KubernetesStackDeployed bool
}

func (s *noopDeployer) DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	s.SwarmStackDeployed = true
	return nil
}

func (s *noopDeployer) DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error {
	s.ComposeStackDeployed = true
	return nil
}

func (s *noopDeployer) DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error {
	s.KubernetesStackDeployed = true
	return nil
}

func Test_redeployWhenChanged_FailsWhenCannotFindStack(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	err := RedeployWhenChanged(1, nil, store, nil, nil)
	assert.Error(t, err)
	assert.Truef(t, strings.HasPrefix(err.Error(), "failed to get the stack"), "it isn't an error we expected: %v", err.Error())
}

func Test_redeployWhenChanged_DoesNothingWhenNotAGitBasedStack(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")
	stack := portaineree.Stack{ID: 1, CreatedBy: "admin"}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().UpdateStack(stack.ID, &stack)
		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "oldHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
	})
}

func Test_redeployWhenChanged_FailsWhenCannotClone(t *testing.T) {
	cloneErr := errors.New("failed to clone")
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:        1,
		CreatedBy: "admin",
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		}})
	assert.NoError(t, err, "failed to create a test stack")

	err = RedeployWhenChanged(1, nil, store, &gitService{cloneErr, "newHash"}, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cloneErr, "should failed to clone but didn't, check test setup")
}

func Test_redeployWhenChanged_ForceUpdateOn(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	tmpDir, _ := ioutil.TempDir("", "stack")

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1})
	assert.NoError(t, err, "error creating environment")

	username := "user"
	err = store.User().Create(&portaineree.User{Username: username, Role: portaineree.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	stack := portaineree.Stack{
		ID:          1,
		EndpointID:  1,
		ProjectPath: tmpDir,
		UpdatedBy:   username,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portaineree.StackAutoUpdate{
			ForceUpdate: true,
		},
	}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "oldHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
		result, _ := store.Stack().Stack(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "oldHash")
	})

	t.Run("can deploy docker swarm stack", func(t *testing.T) {
		stack.Type = portaineree.DockerSwarmStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "oldHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.SwarmStackDeployed, true)
		result, _ := store.Stack().Stack(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "oldHash")
	})

	t.Run("can deploy kube app", func(t *testing.T) {
		stack.Type = portaineree.KubernetesStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "oldHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.KubernetesStackDeployed, true)
		result, _ := store.Stack().Stack(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "oldHash")
	})
}

func Test_redeployWhenChanged_RepoNotChanged_ForceUpdateOff(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	tmpDir, _ := ioutil.TempDir("", "stack")

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:          1,
		CreatedBy:   "admin",
		ProjectPath: tmpDir,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portaineree.StackAutoUpdate{
			ForceUpdate: false,
		},
	})
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}
	err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "oldHash"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, noopDeployer.ComposeStackDeployed, false)
	assert.Equal(t, noopDeployer.SwarmStackDeployed, false)
	assert.Equal(t, noopDeployer.KubernetesStackDeployed, false)
}

func Test_redeployWhenChanged_RepoNotChanged_ForceUpdateOff_ForePullImageEnable(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	tmpDir, _ := ioutil.TempDir("", "stack")

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
	})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:          1,
		CreatedBy:   "admin",
		ProjectPath: tmpDir,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portaineree.StackAutoUpdate{
			ForceUpdate:    false,
			ForcePullImage: true,
		},
		Type: portaineree.DockerComposeStack,
	})
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}
	err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "oldHash"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
	assert.Equal(t, noopDeployer.SwarmStackDeployed, false)
	assert.Equal(t, noopDeployer.KubernetesStackDeployed, false)
}

func Test_redeployWhenChanged_RepoChanged_ForceUpdateOff(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	tmpDir, _ := ioutil.TempDir("", "stack")

	err := store.Endpoint().Create(&portaineree.Endpoint{ID: 1})
	assert.NoError(t, err, "error creating environment")

	username := "user"
	err = store.User().Create(&portaineree.User{Username: username, Role: portaineree.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	stack := portaineree.Stack{
		ID:          1,
		EndpointID:  1,
		ProjectPath: tmpDir,
		UpdatedBy:   username,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portaineree.StackAutoUpdate{
			ForceUpdate: false,
		},
	}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "newHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
		result, _ := store.Stack().Stack(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "newHash")
	})

	t.Run("can deploy docker swarm stack", func(t *testing.T) {
		stack.Type = portaineree.DockerSwarmStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "newHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.SwarmStackDeployed, true)
		result, _ := store.Stack().Stack(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "newHash")
	})

	t.Run("can deploy kube app", func(t *testing.T) {
		stack.Type = portaineree.KubernetesStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, &gitService{nil, "newHash"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.KubernetesStackDeployed, true)
		result, _ := store.Stack().Stack(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "newHash")
	})
}

func Test_getUserRegistries(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true)
	defer teardown()

	endpointID := 123

	admin := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	user := &portaineree.User{ID: 2, Username: "user", Role: portaineree.StandardUserRole}
	err = store.User().Create(user)
	assert.NoError(t, err, "error creating a user")

	team := portaineree.Team{ID: 1, Name: "team"}

	store.TeamMembership().Create(&portaineree.TeamMembership{
		ID:     1,
		UserID: user.ID,
		TeamID: team.ID,
		Role:   portaineree.TeamMember,
	})

	registryReachableByUser := portaineree.Registry{
		ID: 1,
		RegistryAccesses: portaineree.RegistryAccesses{
			portaineree.EndpointID(endpointID): {
				UserAccessPolicies: map[portaineree.UserID]portaineree.AccessPolicy{
					user.ID: {RoleID: portaineree.RoleIDStandardUser},
				},
			},
		},
	}
	err = store.Registry().Create(&registryReachableByUser)
	assert.NoError(t, err, "couldn't create a registry")

	registryReachableByTeam := portaineree.Registry{
		ID: 2,
		RegistryAccesses: portaineree.RegistryAccesses{
			portaineree.EndpointID(endpointID): {
				TeamAccessPolicies: map[portaineree.TeamID]portaineree.AccessPolicy{
					team.ID: {RoleID: portaineree.RoleIDStandardUser},
				},
			},
		},
	}
	err = store.Registry().Create(&registryReachableByTeam)
	assert.NoError(t, err, "couldn't create a registry")

	registryRestricted := portaineree.Registry{
		ID: 3,
		RegistryAccesses: portaineree.RegistryAccesses{
			portaineree.EndpointID(endpointID): {
				UserAccessPolicies: map[portaineree.UserID]portaineree.AccessPolicy{
					user.ID + 100: {RoleID: portaineree.RoleIDStandardUser},
				},
			},
		},
	}
	err = store.Registry().Create(&registryRestricted)
	assert.NoError(t, err, "couldn't create a registry")

	t.Run("admin should has access to all registries", func(t *testing.T) {
		registries, err := getUserRegistries(store, admin, portaineree.EndpointID(endpointID))
		assert.NoError(t, err)
		assert.ElementsMatch(t, []portaineree.Registry{registryReachableByUser, registryReachableByTeam, registryRestricted}, registries)
	})

	t.Run("regular user has access to registries allowed to him and/or his team", func(t *testing.T) {
		registries, err := getUserRegistries(store, user, portaineree.EndpointID(endpointID))
		assert.NoError(t, err)
		assert.ElementsMatch(t, []portaineree.Registry{registryReachableByUser, registryReachableByTeam}, registries)
	})
}

func newChangeWindow(start, end string) *portaineree.Endpoint {
	return &portaineree.Endpoint{
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled:   true,
			StartTime: start,
			EndTime:   end,
		},
	}
}

type MockClock struct {
	timeNow string
}

func (mc MockClock) Now() time.Time {
	gt, _ := time.Parse("15:04", mc.timeNow)
	return gt
}

func NewMockClock(t string) Clock {
	return MockClock{
		timeNow: t,
	}
}

func Test_updateAllowed(t *testing.T) {
	is := assert.New(t)

	// Ensure change window works to the following rules including edge cases (start <= t < endtime) = true
	t.Run("updateAllowed is true inside the time window", func(t *testing.T) {
		endpoint := newChangeWindow("22:00", "23:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("22:30"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed is true when time equal to start time", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "23:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("10:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed is false when time equal to end time (exclusive)", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "11:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("11:00"))
		is.NoError(err)
		is.Equal(false, allowed, "updateAllowed should be false")
	})

	t.Run("updateAllowed is true when start and end time are equal", func(t *testing.T) {
		// we treat this as 24hrs window which means fully on
		endpoint := newChangeWindow("10:00", "10:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("12:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed is false when time outside the time window", func(t *testing.T) {
		endpoint := newChangeWindow("00:00", "05:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("07:00"))
		is.NoError(err)
		is.Equal(false, allowed, "updateAllowed should be false")
	})

	t.Run("updateAllowed when end time spans over to the next day", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "02:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("11:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day and time within and next day", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "02:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("01:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day and time within current day", func(t *testing.T) {
		endpoint := newChangeWindow("10:35", "02:45")
		allowed, err := updateAllowed(endpoint, NewMockClock("10:47"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day and time within next day same hour", func(t *testing.T) {
		endpoint := newChangeWindow("10:30", "02:27")
		allowed, err := updateAllowed(endpoint, NewMockClock("02:20"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day time outside window next day same hour", func(t *testing.T) {
		endpoint := newChangeWindow("10:35", "02:45")
		allowed, err := updateAllowed(endpoint, NewMockClock("02:47"))
		is.NoError(err)
		is.Equal(false, allowed, "updateAllowed should be false")
	})
}

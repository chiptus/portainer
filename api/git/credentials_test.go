package git

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

func TestGetGitConfig(t *testing.T) {
	// set test data store
	is := assert.New(t)
	_, store := datastore.MustNewTestStore(t, true, true)
	gitCredential := &portaineree.GitCredential{ID: 1, Username: "username", Password: "password"}
	err := store.GitCredentialService.Create(gitCredential)
	is.NoError(err, "error creating git credential")

	// case 1: gitConfig is nil
	t.Run("gitConfig is nil", func(t *testing.T) {
		var stack portaineree.Stack
		result, err := GetGitConfigWithPassword(stack.GitConfig, store)
		is.NoError(err, "error getting git config")
		is.Same(stack.GitConfig, result, "git config should be equal")
	})

	// case 2: gitConfig.Authentication is nil
	t.Run("gitConfig authentication is nil", func(t *testing.T) {
		var stack portaineree.Stack
		stack.GitConfig = &gittypes.RepoConfig{}
		result, err := GetGitConfigWithPassword(stack.GitConfig, store)
		is.NoError(err, "error getting git config")
		is.Same(stack.GitConfig, result, "git config should be equal")
	})

	// case 3: gitConfig.Authentication.GitCredentialID is 0
	t.Run("gitConfig authentication git credential id is 0", func(t *testing.T) {
		var stack portaineree.Stack
		stack.GitConfig = &gittypes.RepoConfig{
			Authentication: &gittypes.GitAuthentication{
				Username: "username",
				Password: "password",
			},
		}
		result, err := GetGitConfigWithPassword(stack.GitConfig, store)
		is.NoError(err, "error getting git config")
		is.Same(stack.GitConfig, result, "git config should be equal")
	})

	// case 4: gitConfig.Authentication.GitCredentialID is not 0

	// case 4.1: git credential is not found
	t.Run("git credential is not found", func(t *testing.T) {
		var stack portaineree.Stack
		stack.GitConfig = &gittypes.RepoConfig{
			Authentication: &gittypes.GitAuthentication{
				GitCredentialID: 2,
			},
		}
		_, err := GetGitConfigWithPassword(stack.GitConfig, store)
		is.ErrorContains(err, "failed to get credentials", "error getting git config")
	})

	// case 4.2: git credential is found
	t.Run("git credential is found", func(t *testing.T) {
		var stack portaineree.Stack
		stack.GitConfig = &gittypes.RepoConfig{
			Authentication: &gittypes.GitAuthentication{
				GitCredentialID: 1,
			},
		}
		result, err := GetGitConfigWithPassword(stack.GitConfig, store)
		is.NoError(err, "error getting git config")
		is.NotSame(stack.GitConfig, result, "git config should not be same address")
		is.NotSame(stack.GitConfig.Authentication, result.Authentication, "git config authentication should not be same address")
		is.NotEqual(stack.GitConfig, result, "git config value should not be equal")
		is.NotEmpty(result.Authentication.Username, "username should not be empty")
		is.NotEmpty(result.Authentication.Password, "password should not be empty")
		is.NotEmpty(result.Authentication.GitCredentialID, "git credential id should not be empty")
	})

	// case 5: use datastore transaction
	t.Run("git credential is found with dataStore tx", func(t *testing.T) {
		var stack portaineree.Stack
		stack.GitConfig = &gittypes.RepoConfig{
			Authentication: &gittypes.GitAuthentication{
				GitCredentialID: 1,
			},
		}

		store.UpdateTx(func(tx dataservices.DataStoreTx) error {
			result, err := GetGitConfigWithPassword(stack.GitConfig, store)
			is.NoError(err, "error getting git config")
			is.NotSame(stack.GitConfig, result, "git config should not be same address")
			is.NotSame(stack.GitConfig.Authentication, result.Authentication, "git config authentication should not be same address")
			is.NotEqual(stack.GitConfig, result, "git config value should not be equal")
			is.NotEmpty(result.Authentication.Username, "username should not be empty")
			is.NotEmpty(result.Authentication.Password, "password should not be empty")
			is.NotEmpty(result.Authentication.GitCredentialID, "git credential id should not be empty")
			return err
		})
	})
}

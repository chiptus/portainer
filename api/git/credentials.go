package git

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	gittypes "github.com/portainer/portainer/api/git/types"
)

func GetCredentials(auth *gittypes.GitAuthentication, dataStore dataservices.DataStoreTx) (string, string, error) {
	if auth == nil {
		return "", "", nil
	}

	if auth.GitCredentialID == 0 {
		return auth.Username, auth.Password, nil
	}

	credential, err := dataStore.GitCredential().Read(portaineree.GitCredentialID(auth.GitCredentialID))
	if err != nil {
		return "", "", errors.WithMessagef(err, "failed to get credentials")
	}
	return credential.Username, credential.Password, nil

}

func GetGitConfigWithPassword(gitConfig *gittypes.RepoConfig, dataStore dataservices.DataStoreTx) (*gittypes.RepoConfig, error) {
	if gitConfig == nil {
		return gitConfig, nil
	}

	if gitConfig.Authentication == nil {
		return gitConfig, nil
	}

	if gitConfig.Authentication.GitCredentialID == 0 {
		return gitConfig, nil
	}

	// Prevents the original git config password from being added back again
	// if the stack is deployed by git credential
	newGitConfig := *gitConfig
	newGitConfig.Authentication = &gittypes.GitAuthentication{
		GitCredentialID: gitConfig.Authentication.GitCredentialID,
	}

	username, password, err := GetCredentials(newGitConfig.Authentication, dataStore)
	if err != nil {
		return nil, err
	}

	newGitConfig.Authentication.Username = username
	newGitConfig.Authentication.Password = password

	return &newGitConfig, nil
}

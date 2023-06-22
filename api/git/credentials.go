package git

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	gittypes "github.com/portainer/portainer/api/git/types"
)

func GetCredentials(auth *gittypes.GitAuthentication, dataStore dataservices.DataStore) (string, string, error) {
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

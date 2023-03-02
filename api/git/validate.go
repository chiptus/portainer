package git

import (
	"github.com/asaskevich/govalidator"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	gittypes "github.com/portainer/portainer/api/git/types"
)

func ValidateRepoConfig(repoConfig *gittypes.RepoConfig) error {
	if govalidator.IsNull(repoConfig.URL) || !govalidator.IsURL(repoConfig.URL) {
		return httperrors.NewInvalidPayloadError("Invalid repository URL. Must correspond to a valid URL format")
	}

	return ValidateRepoAuthentication(repoConfig.Authentication)

}

func ValidateRepoAuthentication(auth *gittypes.GitAuthentication) error {
	if auth != nil && govalidator.IsNull(auth.Password) && auth.GitCredentialID == 0 {
		return httperrors.NewInvalidPayloadError("Invalid repository credentials. Password must be specified when authentication is enabled")
	}

	return nil
}

package registryutils

import (
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/aws/ecr"
	"github.com/portainer/portainer-ee/api/dataservices"

	"github.com/rs/zerolog/log"
)

func isRegTokenValid(registry *portaineree.Registry) (valid bool) {
	return registry.AccessToken != "" && registry.AccessTokenExpiry > time.Now().Unix()
}

func doGetRegToken(dataStore dataservices.DataStore, registry *portaineree.Registry) (err error) {
	ecrClient := ecr.NewService(registry.Username, registry.Password, registry.Ecr.Region)
	accessToken, expiryAt, err := ecrClient.GetAuthorizationToken()
	if err != nil {
		return
	}

	registry.AccessToken = *accessToken
	registry.AccessTokenExpiry = expiryAt.Unix()

	err = dataStore.Registry().UpdateRegistry(registry.ID, registry)

	return
}

func parseRegToken(registry *portaineree.Registry) (username, password string, err error) {
	ecrClient := ecr.NewService(registry.Username, registry.Password, registry.Ecr.Region)
	return ecrClient.ParseAuthorizationToken(registry.AccessToken)
}

func EnsureRegTokenValid(dataStore dataservices.DataStore, registry *portaineree.Registry) (err error) {
	if registry.Type == portaineree.EcrRegistry {
		if isRegTokenValid(registry) {
			log.Debug().Msg("current ECR token is still valid")
		} else {
			err = doGetRegToken(dataStore, registry)
			if err != nil {
				log.Debug().Msg("refresh ECR token")
			}
		}
	}

	return
}

func GetRegEffectiveCredential(registry *portaineree.Registry) (username, password string, err error) {
	if registry.Type == portaineree.EcrRegistry {
		username, password, err = parseRegToken(registry)
	} else {
		username = registry.Username
		password = registry.Password
	}
	return
}

package registryutils

import (
	"time"

	log "github.com/sirupsen/logrus"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/aws/ecr"
)

func isManageTokenValid(config *portaineree.RegistryManagementConfiguration) (valid bool) {
	return config.AccessToken != "" && config.AccessTokenExpiry > time.Now().Unix()
}

func doGetManageToken(config *portaineree.RegistryManagementConfiguration) (err error) {
	ecrClient := ecr.NewService(config.Username, config.Password, config.Ecr.Region)
	accessToken, expiryAt, err := ecrClient.GetAuthorizationToken()
	if err != nil {
		return
	}

	config.AccessToken = *accessToken
	config.AccessTokenExpiry = expiryAt.Unix()

	return
}

func EnsureManageTokenValid(config *portaineree.RegistryManagementConfiguration) (err error) {
	if config.Type == portaineree.EcrRegistry {
		if isManageTokenValid(config) {
			log.Println("[DEBUG] [RegistryManagementConfiguration, GetEcrAccessToken] [message: current ECR token is still valid]")
		} else {
			err = doGetManageToken(config)
			if err != nil {
				log.Println("[DEBUG] [RegistryManagementConfiguration, GetEcrAccessToken] [message: refresh ECR token]")
			}
		}
	}

	return
}

func GetManagementCredential(registry *portaineree.Registry) (username, password, region string) {
	config := registry.ManagementConfiguration
	if config != nil {
		return config.Username, config.Password, config.Ecr.Region
	}

	return registry.Username, registry.Password, registry.Ecr.Region
}
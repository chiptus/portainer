package registryutils

import (
	log "github.com/sirupsen/logrus"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/aws/ecr"
)

func isManegeTokenValid(config *portainer.RegistryManagementConfiguration) (valid bool) {
	return config.AccessToken != "" && config.AccessTokenExpiry > time.Now().Unix()
}

func doGetManegeToken(config *portainer.RegistryManagementConfiguration) (err error) {
	ecrClient := ecr.NewService(config.Username, config.Password, config.Ecr.Region)
	accessToken, expiryAt, err := ecrClient.GetAuthorizationToken()
	if err != nil {
		return
	}

	config.AccessToken = *accessToken
	config.AccessTokenExpiry = expiryAt.Unix()

	return
}

func EnsureManegeTokenValid(config *portainer.RegistryManagementConfiguration) (err error) {
	if config.Type == portainer.EcrRegistry {
		if isManegeTokenValid(config) {
			log.Println("[DEBUG] [RegistryManagementConfiguration, GetEcrAccessToken] [message: current ECR token is still valid]")
		} else {
			err = doGetManegeToken(config)
			if err != nil {
				log.Println("[DEBUG] [RegistryManagementConfiguration, GetEcrAccessToken] [message: refresh ECR token]")
			}
		}
	}

	return
}

func GetManagementCredential(registry *portainer.Registry) (username, password, region string) {
	config := registry.ManagementConfiguration
	if config != nil {
		return config.Username, config.Password, config.Ecr.Region
	}

	return registry.Username, registry.Password, registry.Ecr.Region
}
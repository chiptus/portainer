package migrator

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/sirupsen/logrus"
)

func updateEndpoint(m *Migrator, endpoints []portaineree.Endpoint, credID models.CloudCredentialID, providerName string) error {
	for _, endpoint := range endpoints {

		if endpoint.CloudProvider != nil && endpoint.CloudProvider.Name == providerName {
			endpoint.CloudProvider.CredentialID = credID
			err := m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
			if err != nil {
				logrus.Errorf("Unable to update cloud endpoint %s: %s", endpoint.Name, err)
			}
		}
	}

	return nil
}

func (m *Migrator) migrateCloudAPIKeysToCloudCredentials50() error {
	migrateLog.Info("- migrating cloud api keys to cloud credentials")
	settings, err := m.settingsService.Settings()
	if err != nil {
		return errors.Wrap(err, "while fetching settings from database")
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return errors.Wrap(err, "while fetching endpoints from database")
	}

	if settings.CloudApiKeys.CivoApiKey != "" {
		credentials := models.CloudCredential{
			Provider: "civo",
			Name:     "default",
			Credentials: map[string]string{
				"apiKey": settings.CloudApiKeys.CivoApiKey,
			},
		}
		err = m.cloudCredentialService.Create(&credentials)
		if err != nil {
			logrus.Errorf("Unable to create cloud credential for %s: %s", portaineree.CloudProviderCivo, err)
		} else {
			updateEndpoint(m, endpoints, credentials.ID, portaineree.CloudProviderCivo)
		}
	}

	if settings.CloudApiKeys.DigitalOceanToken != "" {
		credentials := models.CloudCredential{
			Provider: "digitalocean",
			Name:     "default",
			Credentials: map[string]string{
				"apiKey": settings.CloudApiKeys.DigitalOceanToken,
			},
		}
		err = m.cloudCredentialService.Create(&credentials)
		if err != nil {
			logrus.Errorf("Unable to create cloud credential for %s: %s", portaineree.CloudProviderDigitalOcean, err)
		} else {
			updateEndpoint(m, endpoints, credentials.ID, portaineree.CloudProviderDigitalOcean)
		}
	}

	if settings.CloudApiKeys.LinodeToken != "" {
		credentials := models.CloudCredential{
			Provider: "linode",
			Name:     "default",
			Credentials: map[string]string{
				"apiKey": settings.CloudApiKeys.LinodeToken,
			},
		}
		err = m.cloudCredentialService.Create(&credentials)
		if err != nil {
			logrus.Errorf("Unable to create cloud credential for %s: %s", portaineree.CloudProviderLinode, err)
		} else {
			updateEndpoint(m, endpoints, credentials.ID, portaineree.CloudProviderLinode)
		}
	}

	// removing cloud api keys from the settings
	settings.CloudApiKeys.CivoApiKey = ""
	settings.CloudApiKeys.DigitalOceanToken = ""
	settings.CloudApiKeys.LinodeToken = ""
	err = m.settingsService.UpdateSettings(settings)
	if err != nil {
		return errors.Wrap(err, "while updating settings")
	}

	return nil
}

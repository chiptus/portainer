package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (m *Migrator) migrateDBVersionToDB50() error {
	if err := m.migrateCloudAPIKeysToCloudCredentials(); err != nil {
		return err
	}

	return m.migratePasswordLengthSettings()
}

func (m *Migrator) migratePasswordLengthSettings() error {
	log.Info().Msg("updating required password length")

	s, err := m.settingsService.Settings()
	if err != nil {
		return errors.Wrap(err, "while fetching settings from database")
	}

	s.InternalAuthSettings.RequiredPasswordLength = 12
	return m.settingsService.UpdateSettings(s)
}

func (m *Migrator) migrateCloudAPIKeysToCloudCredentials() error {
	log.Info().Msg("migrating cloud api keys to cloud credentials")

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
			log.Error().Str("provider", portaineree.CloudProviderCivo).Err(err).Msg("unable to create cloud credential")
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
			log.Error().
				Str("provider", portaineree.CloudProviderDigitalOcean).
				Err(err).
				Msg("unable to create cloud credential")
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
			log.Error().
				Str("provider", portaineree.CloudProviderLinode).
				Err(err).
				Msg("unable to create cloud credential")
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

func updateEndpoint(m *Migrator, endpoints []portaineree.Endpoint, credID models.CloudCredentialID, providerName string) error {
	for _, endpoint := range endpoints {
		if endpoint.CloudProvider != nil && endpoint.CloudProvider.Name == providerName {
			endpoint.CloudProvider.CredentialID = credID
			err := m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
			if err != nil {
				log.Error().Str("endpoint", endpoint.Name).Err(err).Msg("unable to update cloud endpoint")
			}
		}
	}

	return nil
}

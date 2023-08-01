package license

import (
	"net/http"
	"time"

	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/rs/zerolog/log"
)

const (
	syncInterval = time.Hour * 24

	// a period of time after which license overuse restrictions will be enforced.
	// default value: 10 days
	overuseGracePeriodInSeconds = 10 * 24 * 60 * 60
)

func (service *Service) startSyncLoop() error {
	err := service.SyncLicenses()
	if err != nil {
		log.Err(err).Msg("failed initial license sync")
	}

	ticker := time.NewTicker(syncInterval)

	go func() {
		for {
			select {
			case <-service.shutdownCtx.Done():
				log.Debug().Msg("shutting down License service")
				ticker.Stop()

				return
			case <-ticker.C:
				service.SyncLicenses()
			}
		}
	}()

	return nil
}

// syncLiceses checks all of the licenses with the license server to determine
// if any of them have been revoked or otherwise to not exist on the server.
func (service *Service) SyncLicenses() error {
	licenses, err := service.dataStore.License().Licenses()
	if err != nil {
		return err
	}

	var synced []liblicense.PortainerLicense
	for _, l := range licenses {
		l := ParseLicense(l.LicenseKey, service.expireAbsolute)
		valid, err := liblicense.ValidateLicense(&l)
		if err != nil {
			log.Err(err).Msg("invalid license found")
		}
		if !valid {
			l.Revoked = true
		}
		synced = append(synced, l)
	}
	service.licenses = synced
	return nil
}

func RecalculateLicenseUsage(licenseService portaineree.LicenseService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(rw, r)

		if licenseService != nil {
			licenseService.SyncLicenses()
		}
	})
}

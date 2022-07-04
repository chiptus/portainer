package license

import (
	"log"
	"time"

	"github.com/portainer/liblicense"
	"github.com/portainer/liblicense/master"
)

const (
	syncInterval = 24 * time.Hour
)

func (service *Service) startSyncLoop() error {
	err := service.syncLicenses()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(syncInterval)

	go func() {
		for {
			select {
			case <-service.shutdownCtx.Done():
				log.Println("[DEBUG] [internal,license] [message: shutting down License service]")
				ticker.Stop()
				return
			case <-ticker.C:
				service.syncLicenses()
			}
		}
	}()

	return nil
}

func (service *Service) syncLicenses() error {
	licenses, err := service.Licenses()
	if err != nil {
		return err
	}

	licensesToRevoke := []string{}
	hasFreeSubscription := false

	for _, license := range licenses {
		valid, err := master.ValidateLicense(&license)
		if err != nil || !valid {
			licensesToRevoke = append(licensesToRevoke, license.LicenseKey)
			continue
		}

		if license.Type == liblicense.PortainerLicenseEssentials {
			// allow only one single free subscription license, and revoke the rest
			if !hasFreeSubscription {
				hasFreeSubscription = true
				continue
			}
			licensesToRevoke = append(licensesToRevoke, license.LicenseKey)
		}
	}

	for _, licenseKey := range licensesToRevoke {
		err := service.revokeLicense(licenseKey)
		if err != nil {
			return err
		}
	}

	return nil
}

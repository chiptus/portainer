package license

import (
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/portainer/liblicense"
	"github.com/portainer/liblicense/master"
)

const (
	syncInterval = 24 * time.Hour

	// a period of time after which license overuse restrictions will be enforced.
	// default value: 10 days
	overuseGracePeriodInSeconds = 10 * 24 * 60 * 60
)

func (service *Service) startSyncLoop() error {
	err := service.syncLicenses(master.ValidateLicense)
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
				service.syncLicenses(master.ValidateLicense)
			}
		}
	}()

	return nil
}

// syncLiceses
// - revokes all licenses that are either expired, don't stack with each other, or has an invalid key
// - marks license overuse
func (service *Service) syncLicenses(validateLicenseFunc func(license *liblicense.PortainerLicense) (bool, error)) error {
	err := service.revokeInvalidLicenses(validateLicenseFunc)
	if err != nil {
		return err
	}

	return service.ReaggregareLicenseInfo()
}

// revokeInvalidLicenses revokes all licenses that are either invalid or don't stack with each other
func (service *Service) revokeInvalidLicenses(validateLicenseFunc func(license *liblicense.PortainerLicense) (bool, error)) error {
	licenses, err := service.Licenses()
	if err != nil {
		return errors.Wrap(err, "failed to fetch licenses to check invalid licenses")
	}

	licensesToRevoke := []string{}
	var newestTrialLicense *liblicense.PortainerLicense
	var newestEssentialLicense *liblicense.PortainerLicense

	for _, l := range licenses {
		license := l // recapture loop var to get a new memory address
		valid, err := validateLicenseFunc(&license)
		if err != nil || !valid {
			licensesToRevoke = append(licensesToRevoke, license.LicenseKey)
			continue
		}

		// keep only the most recent license of each type
		switch license.Type {
		case liblicense.PortainerLicenseTrial:
			if newestTrialLicense == nil {
				newestTrialLicense = &license
				continue
			}
			if license.Created > newestTrialLicense.Created {
				licensesToRevoke = append(licensesToRevoke, newestTrialLicense.LicenseKey)
				newestTrialLicense = &license
			} else {
				licensesToRevoke = append(licensesToRevoke, license.LicenseKey)
			}
		case liblicense.PortainerLicenseEssentials:
			if newestEssentialLicense == nil {
				newestEssentialLicense = &license
				continue
			}

			if license.Created > newestEssentialLicense.Created {
				licensesToRevoke = append(licensesToRevoke, newestEssentialLicense.LicenseKey)
				newestEssentialLicense = &license
			} else {
				licensesToRevoke = append(licensesToRevoke, license.LicenseKey)
			}
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

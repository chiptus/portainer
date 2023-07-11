package license

import (
	"time"

	"github.com/rs/zerolog/log"
)

const (
	syncInterval = time.Minute * 5

	// a period of time after which license overuse restrictions will be enforced.
	// default value: 10 days
	overuseGracePeriodInSeconds = 10 * 24 * 60 * 60
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
				log.Debug().Msg("shutting down License service")
				ticker.Stop()

				return
			case <-ticker.C:
				service.syncLicenses()
			}
		}
	}()

	return nil
}

// syncLiceses
// - revokes all licenses that are either expired, don't stack with each other, or has an invalid key
// - marks license overuse
func (service *Service) syncLicenses() error {
	return service.ReaggregateLicenseInfo()
}

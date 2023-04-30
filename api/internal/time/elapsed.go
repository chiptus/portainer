package time

import (
	"time"

	"github.com/rs/zerolog/log"
)

// TrackTime is a helper function that logs the time elapsed for a given operation.
func TrackTime(start time.Time, operation string) {
	elapsed := time.Since(start)
	log.Debug().Str("operation", operation).Float64("elapsed (seconds)", elapsed.Seconds()).Msg("Time measurement")
}

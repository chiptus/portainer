package useractivity

import (
	"context"
	"fmt"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"

	storm "github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/rs/zerolog/log"
)

func (store *Store) startCleanupLoop() error {
	log.Debug().Float64("check_interval_seconds", cleanupInterval.Seconds()).Msg("starting logs cleanup process")

	err := store.cleanLogs()
	if err != nil {
		return fmt.Errorf("failed starting logs cleanup process: %w", err)
	}

	ctx := context.Background()
	ctx, store.cancelFn = context.WithCancel(ctx)

	go store.cleanupLoop(ctx)

	return nil
}

func (store *Store) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)

	for {
		select {
		case <-ticker.C:
			log.Debug().Msg("cleaning logs]")

			err := store.cleanLogs()
			if err != nil {
				log.Error().Err(err).Msg("failed clearing auth logs")
			}
		case <-ctx.Done():
			ticker.Stop()

			return
		}
	}
}

func (store *Store) stopCleanupLoop() {
	store.cancelFn()
}

func (store *Store) cleanLogs() error {
	count, err := store.cleanLogsByType(&portaineree.AuthActivityLog{})
	if err != nil {
		return fmt.Errorf("failed cleaning auth logs: %w", err)
	}

	log.Debug().Int("count", count).Msg("removed old auth logs")

	count, err = store.cleanLogsByType(&portaineree.UserActivityLog{})
	if err != nil {
		return fmt.Errorf("failed cleaning user activity logs: %w", err)
	}

	log.Debug().Int("count", count).Msg("removed old user activity logs")

	return nil
}

func (store *Store) cleanLogsByType(obj interface{}) (int, error) {
	oldLogsDate := time.Now().AddDate(0, 0, -1*maxLogsAge).Unix()
	query := store.db.Select(q.Lte("Timestamp", oldLogsDate))

	count, err := query.Count(obj)
	if err != nil && err != storm.ErrNotFound {
		return 0, fmt.Errorf("failed counting old logs: %w", err)
	}

	if count == 0 {
		return count, nil
	}

	err = query.Delete(obj)
	if err != nil && err != storm.ErrNotFound {
		return 0, fmt.Errorf("failed cleaning logs: %w", err)
	}

	return count, nil
}

package useractivity

import (
	"fmt"

	"github.com/asdine/storm/v3/q"
	portaineree "github.com/portainer/portainer-ee/api"
)

// StoreAuthLog stores a new authLog in the store
func (store *Store) StoreAuthLog(authLog *portaineree.AuthActivityLog) error {
	err := store.db.Save(authLog)
	if err != nil {
		return fmt.Errorf("failed saving activity to db: %w", err)
	}

	return nil
}

// GetAuthLogs queries the db for authentication activity logs
// it returns the logs in this page (offset/limit) and the amount of logs in total for this query
func (store *Store) GetAuthLogs(opts portaineree.AuthLogsQuery) ([]*portaineree.AuthActivityLog, int, error) {
	matchers := []q.Matcher{}

	if len(opts.ContextTypes) > 0 {
		matchers = append(matchers, q.In("Context", opts.ContextTypes))
	}

	if len(opts.ActivityTypes) > 0 {
		matchers = append(matchers, q.In("Type", opts.ActivityTypes))
	}

	if opts.Keyword != "" {
		matchers = append(matchers, q.Or(q.Re("Origin", opts.Keyword), q.Re("Username", opts.Keyword)))
	}

	activities := []*portaineree.AuthActivityLog{}
	count, err := store.getLogs(&activities, &portaineree.AuthActivityLog{}, opts.UserActivityLogBaseQuery, matchers)
	if err != nil {
		return nil, 0, err
	}

	return activities, count, nil
}

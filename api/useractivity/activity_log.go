package useractivity

import (
	"github.com/asdine/storm/v3/q"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
)

func (store *Store) StoreUserActivityLog(userLog *portaineree.UserActivityLog) error {

	err := store.db.Save(userLog)
	if err != nil {
		return errors.Wrap(err, "failed saving activity to db")
	}

	return nil
}

// GetActivityLogs queries the db for activity logs
// it returns the logs in this page (offset/limit) and the amount of logs in total for this query
func (store *Store) GetUserActivityLogs(opts portaineree.UserActivityLogBaseQuery) ([]*portaineree.UserActivityLog, int, error) {
	matchers := []q.Matcher{}

	if opts.Keyword != "" {
		matchers = append(matchers, q.Or(q.Re("Context", opts.Keyword), q.Re("Action", opts.Keyword), q.Re("Payload", opts.Keyword), q.Re("Username", opts.Keyword)))
	}

	activities := []*portaineree.UserActivityLog{}
	count, err := store.getLogs(&activities, &portaineree.UserActivityLog{}, opts, matchers)
	if err != nil {
		return nil, 0, err
	}

	return activities, count, nil
}

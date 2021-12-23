package useractivity

import (
	"io"
	"path"
	"time"

	storm "github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	portaineree "github.com/portainer/portainer-ee/api"
	"go.etcd.io/bbolt"
)

const (
	cleanupInterval = 24 * time.Hour
	maxLogsAge      = 7
)

// Store is a store for user activities
type Store struct {
	db                *dbWrapper
	cleanupStopSignal chan struct{}
}

// dbWrapper wraps the storm db type to make it interchangeable
type dbWrapper struct {
	*storm.DB
}

const databaseFileName = "useractivity.db"

// NewStore Creates a new store
func NewStore(dataPath string) (*Store, error) {
	databasePath := path.Join(dataPath, databaseFileName)

	db, err := storm.Open(databasePath)
	if err != nil {
		return nil, err
	}

	err = db.Init(&portaineree.UserActivityLog{})
	if err != nil {
		return nil, err
	}

	err = db.Init(&portaineree.AuthActivityLog{})
	if err != nil {
		return nil, err
	}

	store := &Store{
		db: &dbWrapper{
			DB: db,
		},
	}

	err = store.startCleanupLoop()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// BackupTo backs up db to a provided writer.
// It does hot backup and doesn't block other database reads and writes
func (store *Store) BackupTo(w io.Writer) error {
	return store.db.Bolt.View(func(tx *bbolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

// Close closes the DB
func (store *Store) Close() error {
	store.stopCleanupLoop()

	return store.db.Close()
}

func (store *Store) getLogs(activities interface{}, activityLogType interface{}, opts portaineree.UserActivityLogBaseQuery, matchers []q.Matcher) (int, error) {
	if opts.Limit == 0 {
		opts.Limit = 50
	}

	if opts.SortBy == "" {
		opts.SortBy = "Timestamp"
	}

	matchers = append(matchers, q.Gte("Timestamp", opts.AfterTimestamp))

	if opts.BeforeTimestamp != 0 {
		matchers = append(matchers, q.Lte("Timestamp", opts.BeforeTimestamp))
	}

	query := store.db.Select(matchers...)

	count, err := query.Count(activityLogType)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	limitedQuery := query.Limit(opts.Limit).Skip(opts.Offset).OrderBy(opts.SortBy)

	if opts.SortDesc {
		limitedQuery = limitedQuery.Reverse()
	}

	err = limitedQuery.Find(activities)
	if err != nil && err != storm.ErrNotFound {
		return 0, err
	}

	return count, nil
}

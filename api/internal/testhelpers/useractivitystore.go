package testhelpers

import (
	"io"

	portaineree "github.com/portainer/portainer-ee/api"
)

type store struct{}

func NewUserActivityStore() portaineree.UserActivityStore {
	return &store{}
}

func (s *store) Close() error               { return nil }
func (s *store) BackupTo(w io.Writer) error { return nil }

func (s *store) GetAuthLogs(opts portaineree.AuthLogsQuery) ([]*portaineree.AuthActivityLog, int, error) {
	return nil, 0, nil
}

func (s *store) GetUserActivityLogs(opts portaineree.UserActivityLogBaseQuery) ([]*portaineree.UserActivityLog, int, error) {
	return nil, 0, nil
}

func (s *store) StoreAuthLog(authLog *portaineree.AuthActivityLog) error {
	return nil
}

func (s *store) StoreUserActivityLog(userLog *portaineree.UserActivityLog) error {
	return nil
}

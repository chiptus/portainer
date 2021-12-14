package testhelpers

import (
	"io"

	portainer "github.com/portainer/portainer/api"
)

type store struct{}

func NewUserActivityStore() portainer.UserActivityStore {
	return &store{}
}

func (s *store) Close() error               { return nil }
func (s *store) BackupTo(w io.Writer) error { return nil }

func (s *store) GetAuthLogs(opts portainer.AuthLogsQuery) ([]*portainer.AuthActivityLog, int, error) {
	return nil, 0, nil
}

func (s *store) GetUserActivityLogs(opts portainer.UserActivityLogBaseQuery) ([]*portainer.UserActivityLog, int, error) {
	return nil, 0, nil
}

func (s *store) StoreAuthLog(authLog *portainer.AuthActivityLog) error {
	return nil
}

func (s *store) StoreUserActivityLog(userLog *portainer.UserActivityLog) error {
	return nil
}

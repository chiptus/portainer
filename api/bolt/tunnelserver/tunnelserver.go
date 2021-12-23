package tunnelserver

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "tunnel_server"
	infoKey    = "INFO"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection *internal.DbConnection
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// Info retrieve the TunnelServerInfo object.
func (service *Service) Info() (*portaineree.TunnelServerInfo, error) {
	var info portaineree.TunnelServerInfo

	err := internal.GetObject(service.connection, BucketName, []byte(infoKey), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// UpdateInfo persists a TunnelServerInfo object.
func (service *Service) UpdateInfo(settings *portaineree.TunnelServerInfo) error {
	return internal.UpdateObject(service.connection, BucketName, []byte(infoKey), settings)
}

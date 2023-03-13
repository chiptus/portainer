package database

import (
	"fmt"

	"github.com/portainer/portainer-ee/api/database/boltdb"
	portainer "github.com/portainer/portainer/api"
)

// NewDatabase should use config options to return a connection to the requested database
func NewDatabase(storeType, storePath string, encryptionKey []byte) (connection portainer.Connection, err error) {
	if storeType == "boltdb" {
		return &boltdb.DbConnection{
			Path:          storePath,
			EncryptionKey: encryptionKey,
		}, nil
	}

	return nil, fmt.Errorf("Unknown storage database: %s", storeType)
}

package filesystem

import (
	"os"

	"github.com/portainer/portainer/api/filesystem"
)

const KaasPath = "kaas"

type Service struct {
	fileStorePath string
	filesystem.Service
}

// NewService initializes a new service. It creates a data directory and a directory to store files
// inside this directory if they don't exist.
func NewService(dataStorePath, fileStorePath string) (*Service, error) {
	s, err := filesystem.NewService(dataStorePath, fileStorePath)
	if err != nil {
		return nil, err
	}

	kaasFolder := filesystem.JoinPaths(dataStorePath, KaasPath)

	if err := os.MkdirAll(kaasFolder, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	service := &Service{fileStorePath: fileStorePath, Service: *s}
	return service, nil
}

func (service *Service) GetKaasFolder() string {
	return filesystem.JoinPaths(service.GetDatastorePath(), KaasPath)
}

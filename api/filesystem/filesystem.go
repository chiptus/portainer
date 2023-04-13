package filesystem

import (
	"bytes"
	"io"
	"os"

	"github.com/portainer/portainer/api/filesystem"
)

const (
	KaasPath = "kaas"

	// NomadJobFileDefaultName represents the default name of a nomad job file.
	NomadJobFileDefaultName = "nomad-job.hcl"

	PublicCACertPath      = "CAs"
	ClientCertificateName = "client.pem"
)

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

	err = setupPublicCACerts(dataStorePath, fileStorePath)
	if err != nil {
		return nil, err
	}

	return &Service{fileStorePath: fileStorePath, Service: *s}, nil
}

func setupPublicCACerts(dataStorePath, fileStorePath string) error {
	caCertsPath := filesystem.JoinPaths(dataStorePath, filesystem.SSLCertPath, PublicCACertPath)
	err := os.MkdirAll(caCertsPath, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (service *Service) GetKaasFolder() string {
	return filesystem.JoinPaths(service.GetDatastorePath(), KaasPath)
}

func (service *Service) certPath() string {
	return filesystem.JoinPaths(service.GetDatastorePath(), filesystem.SSLCertPath, PublicCACertPath)
}

// createFile creates a new file in the file store with the content from r.
func (service *Service) createFileInStore(filePath string, r io.Reader) error {
	return filesystem.CreateFile(filePath, r)
}

// Store the client cert into the certs folder
func (service *Service) StoreSSLClientCert(cert []byte) error {
	certPath := filesystem.JoinPaths(service.certPath(), ClientCertificateName)
	return service.createFileInStore(certPath, bytes.NewReader(cert))
}

func (service *Service) GetSSLClientCertPath() string {
	return filesystem.JoinPaths(service.certPath(), ClientCertificateName)
}

package filesystem

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices/edgeconfig"
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

	fileStorePath = filesystem.JoinPaths(dataStorePath, fileStorePath)

	kaasFolder := filesystem.JoinPaths(dataStorePath, KaasPath)

	if err := os.MkdirAll(kaasFolder, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	err = setupPublicCACerts(dataStorePath)
	if err != nil {
		return nil, err
	}

	return &Service{fileStorePath: fileStorePath, Service: *s}, nil
}

func setupPublicCACerts(dataStorePath string) error {
	caCertsPath := filesystem.JoinPaths(dataStorePath, filesystem.SSLCertPath, PublicCACertPath)

	return os.MkdirAll(caCertsPath, 0755)
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

func (service *Service) getStoreEdgeConfigPath(ID portaineree.EdgeConfigID) string {
	return filesystem.JoinPaths(service.fileStorePath, "edge_configs", strconv.Itoa(int(ID)))
}

func (service *Service) StoreEdgeConfigFile(ID portaineree.EdgeConfigID, path string, r io.Reader) error {
	path = filesystem.JoinPaths(service.getStoreEdgeConfigPath(ID), string(portaineree.EdgeConfigCurrent), path)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return service.createFileInStore(path, r)
}

func (service *Service) GetEdgeConfigFilepaths(ID portaineree.EdgeConfigID, version portaineree.EdgeConfigVersion) (basePath string, filepaths []string, err error) {
	basePath = filesystem.JoinPaths(service.getStoreEdgeConfigPath(ID), string(version))

	err = filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		filepaths = append(filepaths, strings.TrimPrefix(path, basePath))

		return nil
	})

	return
}

func (service *Service) GetEdgeConfigDirEntries(edgeConfig *portaineree.EdgeConfig, edgeID string, version portaineree.EdgeConfigVersion) (dirEntries []filesystem.DirEntry, err error) {
	basePath, filepaths, err := service.GetEdgeConfigFilepaths(edgeConfig.ID, version)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve the files for the edge configuration: %w", err)
	}

	for _, p := range filepaths {
		remotePath := p

		switch edgeConfig.Type {
		case edgeconfig.EdgeConfigTypeSpecificFolder:
			after, found := strings.CutPrefix(p, "/"+edgeID)
			if !found {
				continue
			}

			remotePath = after

		case edgeconfig.EdgeConfigTypeSpecificFile:
			if !strings.HasSuffix(p, "/"+edgeID+filepath.Ext(p)) {
				continue
			}
		}

		content, err := os.ReadFile(filepath.Join(basePath, p))
		if err != nil {
			return nil, fmt.Errorf("unable to read the content of the file: %w", err)
		}

		dirEntries = append(dirEntries, filesystem.DirEntry{
			Name:        remotePath,
			Content:     base64.StdEncoding.EncodeToString([]byte(content)),
			IsFile:      true,
			Permissions: 0444,
		})
	}

	return dirEntries, nil
}

func (service *Service) RotateEdgeConfigs(ID portaineree.EdgeConfigID) error {
	prevPath := filesystem.JoinPaths(service.getStoreEdgeConfigPath(ID), string(portaineree.EdgeConfigPrevious))
	curPath := filesystem.JoinPaths(service.getStoreEdgeConfigPath(ID), string(portaineree.EdgeConfigCurrent))

	if err := os.RemoveAll(prevPath); err != nil {
		return err
	}

	if err := os.Rename(curPath, prevPath); err != nil {
		return err
	}

	return nil
}

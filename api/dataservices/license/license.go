package license

import (
	"github.com/portainer/liblicense"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "license"

// Service represents a service for managing license data.
type Service struct {
	connection portainer.Connection
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// License returns a license by licenseKey
func (service *Service) License(licenseKey string) (*liblicense.PortainerLicense, error) {
	var license liblicense.PortainerLicense
	identifier := []byte(licenseKey)

	err := service.connection.GetObject(BucketName, identifier, &license)
	if err != nil {
		return nil, err
	}

	return &license, nil
}

// Licenses return an array containing all the licenses.
func (service *Service) Licenses() ([]liblicense.PortainerLicense, error) {
	var licenses = make([]liblicense.PortainerLicense, 0)

	return licenses, service.connection.GetAll(
		BucketName,
		&liblicense.PortainerLicense{},
		dataservices.AppendFn(&licenses),
	)
}

// AddLicense saves a licence
func (service *Service) AddLicense(licenseKey string, license *liblicense.PortainerLicense) error {
	return service.connection.CreateObjectWithStringId(BucketName, []byte(licenseKey), license)
}

// UpdateLicense updates a license.
func (service *Service) UpdateLicense(licenseKey string, license *liblicense.PortainerLicense) error {
	identifier := []byte(licenseKey)
	return service.connection.UpdateObject(BucketName, identifier, license)
}

// DeleteLicense deletes a License.
func (service *Service) DeleteLicense(licenseKey string) error {
	identifier := []byte(licenseKey)
	return service.connection.DeleteObject(BucketName, identifier)
}

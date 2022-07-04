package license

import (
	"fmt"

	"github.com/portainer/liblicense"
	"github.com/portainer/portainer-ee/api/license"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "license"
)

// Service represents a service for managing environment(endpoint) data.
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

	err := service.connection.GetAll(
		BucketName,
		&liblicense.PortainerLicense{},
		func(obj interface{}) (interface{}, error) {
			r, ok := obj.(*liblicense.PortainerLicense)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to PortainerLicense object")
				return nil, fmt.Errorf("Failed to convert to PortainerLicense object: %s", obj)
			}

			l := license.ParseLicense(*r)
			licenses = append(licenses, l)

			return &liblicense.PortainerLicense{}, nil
		})
	return licenses, err
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

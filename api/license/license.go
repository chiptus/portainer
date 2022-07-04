package license

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/portainer/liblicense"
	"github.com/portainer/liblicense/master"
	portaineree "github.com/portainer/portainer-ee/api"
)

// Service represents a service for managing portainer licenses
type Service struct {
	repository  portaineree.LicenseRepository
	info        *portaineree.LicenseInfo
	shutdownCtx context.Context
}

// NewService creates a new instance of Service
func NewService(repository portaineree.LicenseRepository, shutdownCtx context.Context) *Service {
	return &Service{
		repository:  repository,
		info:        nil,
		shutdownCtx: shutdownCtx,
	}
}

// Start starts the service
func (service *Service) Start() error {
	return service.startSyncLoop()
}

// Init initializes internal state
func (service *Service) Init() error {
	service.info = &portaineree.LicenseInfo{Valid: false}

	licenses, err := service.Licenses()
	if err != nil {
		return err
	}

	service.aggregate(licenses)
	return nil
}

// Info returns aggregation of the information about the existing licenses
func (service *Service) Info() *portaineree.LicenseInfo {
	return service.info
}

// Licenses returns the list of the existing licenses
func (service *Service) Licenses() ([]liblicense.PortainerLicense, error) {
	return service.repository.Licenses()
}

// AddLicense adds the license to instance
func (service *Service) AddLicense(licenseKey string) (*liblicense.PortainerLicense, error) {
	licenses, err := service.Licenses()
	if err != nil {
		return nil, err
	}

	license, err := master.ParseLicenseKey(licenseKey)
	if err != nil {
		return nil, err
	}

	if license.Type == liblicense.PortainerLicenseTrial && len(licenses) > 0 {
		return nil, errors.New("Trial license can not be applied when there are existing licenses")
	}

	for _, existingLicense := range licenses {
		if existingLicense.Type == liblicense.PortainerLicenseEssentials && license.Type == liblicense.PortainerLicenseEssentials {
			if existingLicense.LicenseKey != license.LicenseKey {
				return nil, errors.New("Multiple free licenses are not allowed")
			}
		}

		if existingLicense.Type == liblicense.PortainerLicenseTrial {
			err = service.repository.DeleteLicense(existingLicense.LicenseKey)
			if err != nil {
				return nil, err
			}

			log.Printf("[DEBUG] [msg: removing the existing trial license]")
		}
	}

	valid, err := master.ValidateLicense(license)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("license is invalid")
	}

	err = service.repository.AddLicense(license.LicenseKey, license)
	if err != nil {
		return nil, err
	}

	licenses, err = service.Licenses()
	if err != nil {
		return nil, err
	}

	service.aggregate(licenses)
	return license, nil
}

// DeleteLicense removes the license from instance
func (service *Service) DeleteLicense(licenseKey string) error {

	// BUG: When the frontend makes a delete request it does so with the license
	// key json field NOT the actual boltDB key. Which means if these two values
	// are ever different (ie. database corruption or manual database editing)
	// the delete will fail. Our current database abstractions do not provide a
	// way to list boltDB keys so we have no way to detect if these values are
	// different. This problem exists throughout portainer. You also cannot
	// delete an endpoint if it's json ID is different than it's boltDB key.
	license, err := service.repository.License(licenseKey)
	if err != nil {
		return err
	}

	valid := isLicenseValid(*license)

	if valid {
		licenses, err := service.Licenses()
		if err != nil {
			return err
		}

		hasMoreValidLicenses := false
		for _, otherLicense := range licenses {
			if licenseKey != otherLicense.LicenseKey && isLicenseValid(otherLicense) {
				hasMoreValidLicenses = true
				break
			}
		}

		if !hasMoreValidLicenses {
			return errors.New("At least one valid license is expected")
		}
	}

	err = service.repository.DeleteLicense(licenseKey)
	if err != nil {
		return err
	}

	licenses, err := service.Licenses()
	if err != nil {
		return err
	}

	service.aggregate(licenses)
	return nil
}

// revokeLicense attempts to mark a license in the database and in the running
// info cache as revoked.
func (service *Service) revokeLicense(licenseKey string) error {
	var licenses []liblicense.PortainerLicense
	defer service.aggregate(licenses)

	license, err := service.repository.License(licenseKey)
	if err != nil {
		return err
	}

	license.Revoked = true

	err = service.repository.UpdateLicense(licenseKey, license)
	if err != nil {
		return err
	}

	licenses, err = service.Licenses()
	return err
}

// aggregate takes a list of licenses and updates service.info with a single
// combined license value. If there are no valid licenses this license will be
// revoked.
func (service *Service) aggregate(licenses []liblicense.PortainerLicense) {
	nodes := 0
	var expiresAt time.Time
	company := ""
	edition := liblicense.PortainerEE
	licenseType := liblicense.PortainerLicenseSubscription

	if len(licenses) == 0 {
		service.info = &portaineree.LicenseInfo{Valid: false}
	}

	hasValidLicenses := false
	for _, license := range licenses {
		l := ParseLicense(license)

		valid := isLicenseValid(l)
		if !valid {
			continue
		}
		hasValidLicenses = true

		if l.Company != "" {
			company = l.Company
		}

		nodes = nodes + l.Nodes

		licenseExpiresAt := licenseExpiresAt(l)
		if licenseExpiresAt.Before(expiresAt) || expiresAt.IsZero() {
			expiresAt = licenseExpiresAt
		}

		if int(l.ProductEdition) < int(edition) {
			edition = l.ProductEdition
		}

		if l.Type == liblicense.PortainerLicenseTrial {
			licenseType = liblicense.PortainerLicenseTrial
		}

	}

	expiresAtUnix := expiresAt.Unix()
	if expiresAt.IsZero() {
		expiresAtUnix = 0
	}

	service.info = &portaineree.LicenseInfo{
		Company:        company,
		ProductEdition: edition,
		Nodes:          nodes,
		ExpiresAt:      expiresAtUnix,
		Type:           licenseType,
		Valid:          hasValidLicenses,
	}
}

func ParseLicense(l liblicense.PortainerLicense) liblicense.PortainerLicense {
	key := l.LicenseKey

	parsedLicense, err := master.ParseLicenseKey(key)
	if err != nil {
		// If we can't even parse the license we simply revoke it.
		l.Revoked = true
		return l
	}

	return *parsedLicense
}

func licenseExpiresAt(license liblicense.PortainerLicense) time.Time {
	return master.ExpiresAt(license.Created, license.ExpiresAfter)
}

func isLicenseValid(license liblicense.PortainerLicense) bool {
	return !master.Expired(license.Created, license.ExpiresAfter) && !license.Revoked
}

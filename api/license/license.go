package license

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/portainer/liblicense"
	"github.com/portainer/liblicense/master"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"

	"github.com/rs/zerolog/log"
)

// Service represents a service for managing portainer licenses
type Service struct {
	info        *portaineree.LicenseInfo
	dataStore   dataservices.DataStore
	shutdownCtx context.Context
}

// NewService creates a new instance of Service
func NewService(dataStore dataservices.DataStore, shutdownCtx context.Context) *Service {

	return &Service{
		info:        nil,
		dataStore:   dataStore,
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

	service.info = service.aggregateLicenses(licenses)

	return nil
}

// Info returns aggregation of the information about the existing licenses
func (service *Service) Info() *portaineree.LicenseInfo {
	return service.info
}

// Licenses returns the list of the existing licenses
func (service *Service) Licenses() ([]liblicense.PortainerLicense, error) {
	return service.dataStore.License().Licenses()
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
			err = service.dataStore.License().DeleteLicense(existingLicense.LicenseKey)
			if err != nil {
				return nil, err
			}

			log.Debug().Msg("removing the existing trial license")
		}
	}

	valid, err := master.ValidateLicense(license)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("license is invalid")
	}

	err = service.dataStore.License().AddLicense(license.LicenseKey, license)
	if err != nil {
		return nil, err
	}

	licenses, err = service.Licenses()
	if err != nil {
		return nil, err
	}

	service.info = service.aggregateLicenses(licenses)
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
	license, err := service.dataStore.License().License(licenseKey)
	if err != nil {
		return err
	}

	valid := isExpiredOrRevoked(*license)

	if valid {
		licenses, err := service.Licenses()
		if err != nil {
			return err
		}

		hasMoreValidLicenses := false
		for _, otherLicense := range licenses {
			if licenseKey != otherLicense.LicenseKey && isExpiredOrRevoked(otherLicense) {
				hasMoreValidLicenses = true
				break
			}
		}

		if !hasMoreValidLicenses {
			return errors.New("At least one valid license is expected")
		}
	}

	err = service.dataStore.License().DeleteLicense(licenseKey)
	if err != nil {
		return err
	}

	licenses, err := service.Licenses()
	if err != nil {
		return err
	}

	service.info = service.aggregateLicenses(licenses)

	return nil
}

// revokeLicense attempts to mark a license in the database and in the running
// info cache as revoked.
func (service *Service) revokeLicense(licenseKey string) error {
	var licenses []liblicense.PortainerLicense

	license, err := service.dataStore.License().License(licenseKey)
	if err != nil {
		return errors.Wrap(err, "failed to fetch licenses to revoke")
	}

	license.Revoked = true

	err = service.dataStore.License().UpdateLicense(licenseKey, license)
	if err != nil {
		return errors.Wrap(err, "failed to revoke a license")
	}

	licenses, err = service.Licenses()
	if err != nil {
		return errors.Wrap(err, "failed to fetch licenses")
	}

	service.info = service.aggregateLicenses(licenses)

	return nil
}

// ReaggregareLicenseInfo re-calculates and updates the aggregated license
func (service *Service) ReaggregareLicenseInfo() error {
	licenses, err := service.Licenses()
	if err == nil {
		service.info = service.aggregateLicenses(licenses)
	}

	return errors.Wrap(err, "failed to fetch licenses to aggregate")
}

func RecalculateLicenseUsage(licenseService portaineree.LicenseService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(rw, r)

		if licenseService != nil {
			licenseService.ReaggregareLicenseInfo()
		}
	})
}

// aggregateLicenses takes a list of licenses and calculates a single combined license value.
// If there are no valid licenses this license will be empty and invalid.
func (service *Service) aggregateLicenses(licenses []liblicense.PortainerLicense) *portaineree.LicenseInfo {
	if len(licenses) == 0 {
		return &portaineree.LicenseInfo{Valid: false}
	}

	nodes := 0
	var expiresAt time.Time
	company := ""
	var licenseType liblicense.PortainerLicenseType
	hasValidLicenses := false

	for _, license := range licenses {
		l := ParseLicense(license)

		if isExpiredOrRevoked(l) {
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

		// if there is at least one Trial license, we'll consider the license as a Trial
		// otherwise, if there are only Essential licenses, we'll consider the license as an Essential
		// otherwise the license is a Subscription
		switch l.Type {
		case liblicense.PortainerLicenseTrial:
			licenseType = liblicense.PortainerLicenseTrial
		case liblicense.PortainerLicenseEssentials:
			if licenseType == 0 {
				licenseType = liblicense.PortainerLicenseEssentials
			}
		case liblicense.PortainerLicenseSubscription:
			if licenseType == 0 || licenseType == liblicense.PortainerLicenseEssentials {
				licenseType = liblicense.PortainerLicenseSubscription
			}
		}
	}

	expiresAtUnix := expiresAt.Unix()
	if expiresAt.IsZero() {
		expiresAtUnix = 0
	}

	licenseOveruseTimestamp, err := service.getLicenseOveruseTimestamp(licenseType, nodes)
	if err == nil {
		service.dataStore.Enforcement().UpdateOveruseStartedTimestamp(licenseOveruseTimestamp)
	}

	return &portaineree.LicenseInfo{
		Company:                 company,
		Nodes:                   nodes,
		ExpiresAt:               expiresAtUnix,
		Type:                    licenseType,
		Valid:                   hasValidLicenses,
		OveruseStartedTimestamp: licenseOveruseTimestamp,
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

	// replace with values that are encrypted in the license key
	l.Company = parsedLicense.Company
	l.Created = parsedLicense.Created
	l.ExpiresAfter = parsedLicense.ExpiresAfter
	l.ExpiresAt = parsedLicense.ExpiresAt
	l.Nodes = parsedLicense.Nodes
	l.ProductEdition = parsedLicense.ProductEdition
	l.Type = parsedLicense.Type
	l.Version = parsedLicense.Version

	return l
}

func licenseExpiresAt(license liblicense.PortainerLicense) time.Time {
	return master.ExpiresAt(license.Created, license.ExpiresAfter)
}

func isExpiredOrRevoked(license liblicense.PortainerLicense) bool {
	return master.Expired(license.Created, license.ExpiresAfter) || license.Revoked
}

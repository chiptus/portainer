package license

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"

	"github.com/pkg/errors"
)

// Service represents a service for managing portainer licenses
type Service struct {
	info            *portaineree.LicenseInfo
	dataStore       dataservices.DataStore
	shutdownCtx     context.Context
	snapshotService portaineree.SnapshotService
	expireAbsolute  bool
}

// NewService creates a new instance of Service
func NewService(dataStore dataservices.DataStore, shutdownCtx context.Context, snapshotService portaineree.SnapshotService, expireAbsolute bool) *Service {
	return &Service{
		info:            nil,
		dataStore:       dataStore,
		shutdownCtx:     shutdownCtx,
		snapshotService: snapshotService,
		expireAbsolute:  expireAbsolute,
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

// AddLicense attempts to add a license to instance.
func (service *Service) AddLicense(key string, force bool) ([]string, error) {
	// Validate the given license key and parse it into a license object.
	l := ParseLicense(key, service.expireAbsolute)
	if l.Revoked {
		return nil, fmt.Errorf("license is invalid")
	}
	valid, err := liblicense.ValidateLicense(&l)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("license key is invalid")
	}
	if isExpiredOrRevoked(l) {
		return nil, fmt.Errorf("license key is expired or revoked")
	}

	// Fetch a list of the existing licenses.
	licenses, err := service.Licenses()
	if err != nil {
		return nil, err
	}

	// If there are no existing licenses we accept their given license.
	if len(licenses) == 0 {
		err = service.dataStore.License().AddLicense(l.LicenseKey, &l)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, l)
		service.info = service.aggregateLicenses(licenses)

		return nil, nil
	}

	// Subscription licenses are stackable with other subscription licenses.
	// V2 Free licenses are stackable with V2 Subscription licenses.
	// With all other license types, you must remove your existing
	// licenses in order to add a new license. We warn the user of this.
	var conflicts []string
	switch {
	case l.Type == liblicense.PortainerLicenseSubscription && l.Version == 3:
		// V3 Subscription licenses can only be added if all existing licenses
		// are also subscription licenses.
		for _, l := range licenses {
			if l.Type != liblicense.PortainerLicenseSubscription {
				if force {
					err := service.dataStore.License().DeleteLicense(l.LicenseKey)
					if err != nil {
						return nil, err
					}
				}
				conflicts = append(conflicts,
					fmt.Sprintf(
						"%s (type %d %s - %s)",
						l.Company,
						l.Type,
						displayType(l.Type),
						displayNodes(l.Nodes),
					),
				)
			}
		}
	case l.Type == liblicense.PortainerLicenseSubscription && l.Version != 3:
		// V2 Subscription licenses can be added if all existing licenses
		// are subscription licenses OR V2 free licenses.
		for _, l := range licenses {
			if l.Type == liblicense.PortainerLicenseSubscription ||
				(l.Type == liblicense.PortainerLicenseFree && l.Version != 3) {
				continue
			}
			if force {
				err := service.dataStore.License().DeleteLicense(l.LicenseKey)
				if err != nil {
					return nil, err
				}
			}
			conflicts = append(conflicts,
				fmt.Sprintf(
					"%s (type %d %s - %s)",
					l.Company,
					l.Type,
					displayType(l.Type),
					displayNodes(l.Nodes),
				),
			)
		}
	case l.Type == liblicense.PortainerLicenseFree && l.Version != 3:
		// V2 Free licenses are stackable with other V2 Free licenses or V2
		// Subscription licenses.
		for _, l := range licenses {
			if l.Version != 3 {
				if l.Type == liblicense.PortainerLicenseSubscription {
					continue
				}
			}
			if force {
				err := service.dataStore.License().DeleteLicense(l.LicenseKey)
				if err != nil {
					return nil, err
				}
			}
			conflicts = append(conflicts,
				fmt.Sprintf(
					"%s (type %d %s - %s)",
					l.Company,
					l.Type,
					displayType(l.Type),
					displayNodes(l.Nodes),
				),
			)
		}
	default:
		for _, l := range licenses {
			if force {
				err := service.dataStore.License().DeleteLicense(l.LicenseKey)
				if err != nil {
					return nil, err
				}
			}
			conflicts = append(conflicts,
				fmt.Sprintf(
					"%s (type %d %s - %s)",
					l.Company,
					l.Type,
					displayType(l.Type),
					displayNodes(l.Nodes),
				),
			)
		}
	}

	if len(conflicts) != 0 && !force {
		return conflicts, nil
	}

	err = service.dataStore.License().AddLicense(l.LicenseKey, &l)
	if err != nil {
		return nil, err
	}

	// Fetch new list of licenses.
	licenses, err = service.Licenses()
	if err != nil {
		return nil, err
	}
	service.info = service.aggregateLicenses(licenses)

	return conflicts, nil
}

func displayType(t liblicense.PortainerLicenseType) string {
	switch t {
	case liblicense.PortainerLicenseTrial:
		return "Trial"
	case liblicense.PortainerLicenseSubscription:
		return "Subscription"
	case liblicense.PortainerLicenseFree:
		return "Free"
	case liblicense.PortainerLicensePersonal:
		return "Personal"
	case liblicense.PortainerLicenseStarter:
		return "Starter"
	default:
		return "Unknown"
	}
}

func displayNodes(n int) string {
	var count string
	if n == 0 {
		count = "unlimited"
	} else {
		count = strconv.Itoa(n)
	}

	return count + " nodes"
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
	err := service.dataStore.License().DeleteLicense(licenseKey)
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

// ReaggregateLicenseInfo re-calculates and updates the aggregated license
func (service *Service) ReaggregateLicenseInfo() error {
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
			licenseService.ReaggregateLicenseInfo()
		}
	})
}

// aggregateLicenses takes a list of licenses and calculates a single combined license value.
// If there are no valid licenses this license will be empty and invalid.
func (service *Service) aggregateLicenses(licenses []liblicense.PortainerLicense) *portaineree.LicenseInfo {
	// If we have no licenses return immediately.
	if len(licenses) == 0 {
		return &portaineree.LicenseInfo{Valid: false}
	}

	// If we have any trial licenses use the first one.
	if info, ok := trialLicenseInfo(licenses, service.expireAbsolute); ok {
		return &info
	}

	// If we have any subscription licenses use them.
	if info, ok := subLicenseInfo(licenses, service.expireAbsolute); ok {
		return &info
	}

	// Otherwise, use the first remaining license.
	var info portaineree.LicenseInfo
	info.Valid = !isExpiredOrRevoked(licenses[0])
	info.Nodes = licenses[0].Nodes
	if !info.Valid {
		info.Nodes = 0
	}

	info.Type = licenses[0].Type
	info.Company = licenses[0].Company
	info.ExpiresAt = licenses[0].ExpiresAt
	licenseOveruseTimestamp, err := service.getLicenseOveruseTimestamp(
		licenses[0].Type,
		licenses[0].Nodes,
	)
	info.OveruseStartedTimestamp = licenseOveruseTimestamp
	if err == nil {
		service.dataStore.Enforcement().UpdateOveruseStartedTimestamp(licenseOveruseTimestamp)
	}

	return &info
}

func trialLicenseInfo(licenses []liblicense.PortainerLicense, expireAbsolute bool) (portaineree.LicenseInfo, bool) {
	var info portaineree.LicenseInfo
	for _, l := range licenses {
		l := ParseLicense(l.LicenseKey, expireAbsolute)
		valid, err := liblicense.ValidateLicense(&l)
		if err != nil || !valid {
			continue
		}
		if isExpiredOrRevoked(l) {
			continue
		}

		if l.Type == liblicense.PortainerLicenseTrial {
			info.Valid = true
			info.Type = liblicense.PortainerLicenseTrial
			info.Company = l.Company
			info.Nodes = 0
			info.ExpiresAt = l.ExpiresAt
			return info, true
		}
	}

	return info, false
}

// subLicenseInfo adds all of a user's subscription licenses to produce a
// single aggregate value. As a special case, V2 Free licenses are added IF all
// other subscription licenses are also V2.
func subLicenseInfo(licenses []liblicense.PortainerLicense, expireAbsolute bool) (portaineree.LicenseInfo, bool) {
	var found bool
	var info portaineree.LicenseInfo
	info.ExpiresAt = math.MaxInt64

	var foundV3 bool
	var freeV2Nodes int
	for _, l := range licenses {
		valid, err := liblicense.ValidateLicense(&l)
		if err != nil || !valid {
			continue
		}
		if isExpiredOrRevoked(l) {
			continue
		}

		if l.Type == liblicense.PortainerLicenseFree && l.Version != 3 {
			freeV2Nodes = l.Nodes
			continue
		}

		if l.Type == liblicense.PortainerLicenseSubscription {
			if l.Version == 3 {
				foundV3 = true
			}

			l := ParseLicense(l.LicenseKey, expireAbsolute)
			if isExpiredOrRevoked(l) {
				continue
			}
			found = true
			info.Valid = true
			info.Type = liblicense.PortainerLicenseSubscription
			if l.Company != "" {
				info.Company = l.Company
			}

			info.Nodes += l.Nodes

			if l.ExpiresAt < info.ExpiresAt {
				info.ExpiresAt = l.ExpiresAt
			}
		}
	}

	if !foundV3 {
		info.Nodes += freeV2Nodes
	}

	return info, found
}

func ParseLicense(key string, expireAbsolute bool) liblicense.PortainerLicense {
	var l liblicense.PortainerLicense
	l.LicenseKey = key
	parsedLicense, err := liblicense.ParseLicenseKey(key)
	if err != nil {
		// If we can't even parse the license we simply revoke it.
		l.Revoked = true
		return l
	}

	// Use values encrypted in the license key.
	l.Company = parsedLicense.Company
	l.Created = parsedLicense.Created
	l.ExpiresAfter = parsedLicense.ExpiresAfter
	l.Nodes = parsedLicense.Nodes
	l.ProductEdition = parsedLicense.ProductEdition
	l.Type = parsedLicense.Type
	l.Version = parsedLicense.Version

	// ExpiresAt is normally rounded to the end of the nearest day.
	// If we were given a command line flag requesting the absolute expiration
	// time we instead override this value.
	l.ExpiresAt = parsedLicense.ExpiresAt
	if expireAbsolute {
		l.ExpiresAt = time.Unix(l.Created, 0).
			Add(time.Hour * time.Duration(l.ExpiresAfter) * 24).Unix()
	}

	return l
}

func licenseExpiresAt(license liblicense.PortainerLicense) time.Time {
	return liblicense.ExpiresAt(license.Created, license.ExpiresAfter)
}

func isExpiredOrRevoked(license liblicense.PortainerLicense) bool {
	now := time.Now()
	return now.After(time.Unix(license.ExpiresAt, 0)) || license.Revoked
}

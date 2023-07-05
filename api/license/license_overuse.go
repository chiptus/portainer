package license

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	nodeutil "github.com/portainer/portainer-ee/api/internal/nodes"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	"github.com/portainer/portainer/pkg/libhelm/time"

	"github.com/pkg/errors"
)

// NotOverused should be applied to a new endpoint provisioning,
// will ensure that the Essential license(5NF) will not be overused by adding an endpoint.
// It will respond with an error.
func NotOverused(licenseService portaineree.LicenseService, dataStore dataservices.DataStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if licenseService == nil || dataStore == nil {
			next.ServeHTTP(rw, r)
			return
		}

		licenseInfo := licenseService.Info()
		// license validation is only relevant for an Essential license
		if licenseInfo.Type == liblicense.PortainerLicenseFree {
			endpoints, err := dataStore.Endpoint().Endpoints()
			if err != nil {
				next.ServeHTTP(rw, r)
				return
			}

			for i := range endpoints {
				err = snapshot.FillSnapshotData(dataStore, &endpoints[i])
				if err != nil {
					next.ServeHTTP(rw, r)
					return
				}
			}

			if licenseIsAtTheLimit(licenseInfo.Nodes, endpoints) {
				httperror.WriteError(rw, http.StatusPaymentRequired, "You have exceeded the node allowance of your current license. Please contact your administrator.", nil)
				return
			}
		}

		next.ServeHTTP(rw, r)
	})
}

// ShouldEnforceOveruse returns true if the license limit was exceeded for longer than a grace period
func (service *Service) ShouldEnforceOveruse() bool {
	enforcementTimestamp := service.WillBeEnforcedAt()
	if enforcementTimestamp == 0 {
		return false
	}

	return time.Now().Unix() > enforcementTimestamp
}

// WillBeEnforcedAt returns a timestamp when the license overuse will be enforced.
// If the license isn't overused, it will return 0.
func (service *Service) WillBeEnforcedAt() int64 {
	overuserdStartedTimestamp := service.info.OveruseStartedTimestamp
	if overuserdStartedTimestamp == 0 {
		return 0
	}

	return overuserdStartedTimestamp + overuseGracePeriodInSeconds
}

// getLicenseOveruseTimestamp returns 0 if license isn't overused
// otherwise non-zero unit timestamp
func (service *Service) getLicenseOveruseTimestamp(licenseType liblicense.PortainerLicenseType, licensedNodesCount int) (int64, error) {
	enforcement, err := service.dataStore.Enforcement().LicenseEnforcement()
	if err != nil {
		return int64(0), errors.Wrap(err, "failed to fetch license enforcement information")
	}
	licenseOveruseStartedTimestamp := enforcement.LicenseOveruseStartedTimestamp

	// Only Free License node abuse can trigger license enforcement.
	if licenseType != liblicense.PortainerLicenseFree {
		return 0, nil
	}

	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		return 0, nil
	}

	for i := range endpoints {
		err = service.snapshotService.FillSnapshotData(&endpoints[i])
		if err != nil {
			return 0, nil
		}
	}

	licenseOverused := licenseIsOverused(licensedNodesCount, endpoints)

	if licenseOveruseStartedTimestamp == 0 && licenseOverused {
		licenseOveruseStartedTimestamp = time.Now().Unix()
	}

	if !licenseOverused {
		licenseOveruseStartedTimestamp = 0
	}

	return licenseOveruseStartedTimestamp, nil
}

// licenseIsOverused returns true if the license node quota is exceeded
func licenseIsOverused(allowedNodes int, endpoints []portaineree.Endpoint) bool {
	return nodeutil.NodesCount(endpoints) > allowedNodes
}

// licenseIsAtTheLimit validates that the license is not at its node limit or over it.
func licenseIsAtTheLimit(allowedNodes int, endpoints []portaineree.Endpoint) bool {
	return nodeutil.NodesCount(endpoints) >= allowedNodes
}

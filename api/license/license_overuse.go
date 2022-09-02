package license

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/portainer/libhelm/time"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/liblicense"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	nodeutil "github.com/portainer/portainer-ee/api/internal/nodes"
)

// NotOverused should be applied to a new endpoint provisioning,
// will ensure that the Essential license(5NF) will not be overused by adding an endpoint.
// It will responds with an error.
func NotOverused(licenseService portaineree.LicenseService, dataStore dataservices.DataStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if licenseService == nil || dataStore == nil {
			next.ServeHTTP(rw, r)
			return
		}

		licenseInfo := licenseService.Info()
		// license validation is only relevant for an Essential license
		if licenseInfo.Type == liblicense.PortainerLicenseEssentials {
			if licenseIsAtTheLimit(licenseInfo.Nodes, dataStore.Endpoint()) {
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

	// current only Essenstial License node abuse can trigger license enforcement
	if licenseType != liblicense.PortainerLicenseEssentials {
		return 0, nil
	}

	licenseOverused := licenseIsOverused(licensedNodesCount, service.dataStore.Endpoint())

	if licenseOveruseStartedTimestamp == 0 && licenseOverused {
		licenseOveruseStartedTimestamp = time.Now().Unix()
	}

	if !licenseOverused {
		licenseOveruseStartedTimestamp = 0
	}

	return licenseOveruseStartedTimestamp, nil
}

// licenseIsOverused returns true if the license node quota is exceeded
func licenseIsOverused(allowedNodes int, endpointService dataservices.EndpointService) bool {
	return nodesInUse(endpointService) > allowedNodes
}

// licenseIsAtTheLimit validates that the license is not at its node limit or over it.
func licenseIsAtTheLimit(allowedNodes int, endpointService dataservices.EndpointService) bool {
	return nodesInUse(endpointService) >= allowedNodes
}

// nodesInUse returns the number of nodes that are currently provisioned
func nodesInUse(endpointService dataservices.EndpointService) int {
	endpoints, err := endpointService.Endpoints()
	if err != nil {
		return 0
	}

	return nodeutil.NodesCount(endpoints)
}

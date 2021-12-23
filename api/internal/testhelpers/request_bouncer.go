package testhelpers

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
)

type testRequestBouncer struct {
}

// NewTestRequestBouncer creates new mock for requestBouncer
func NewTestRequestBouncer() *testRequestBouncer {
	return &testRequestBouncer{}
}

func (testRequestBouncer) AuthenticatedAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) RestrictedAccess(h http.Handler) http.Handler {
	return h
}

// PublicAccess defines a security check for public API environments(endpoints).
// No authentication is required to access these environments(endpoints).
func (testRequestBouncer) PublicAccess(h http.Handler) http.Handler {
	return h
}

// AdminAccess is an alias for RestrictedAddress
// It's not removed as it's used across our codebase and removing will cause conflicts with CE
func (testRequestBouncer) AdminAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) AuthorizedEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint, authorizationCheck bool) error {
	return nil
}

// AuthorizedEdgeEndpointOperation verifies that the request was received from a valid Edge environment(endpoint)
func (testRequestBouncer) AuthorizedEdgeEndpointOperation(r *http.Request, endpoint *portaineree.Endpoint) error {
	return nil
}

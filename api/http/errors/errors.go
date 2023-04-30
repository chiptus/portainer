package errors

import "errors"

var (
	// ErrEndpointAccessDenied Access denied to environment(endpoint) error
	ErrEndpointAccessDenied = errors.New("Access denied to environment")
	// ErrNoValidLicense Unauthorized error
	ErrNoValidLicense = errors.New("No valid Portainer License found")
	// ErrLicenseOverused License overused error
	ErrLicenseOverused = errors.New("Node limit exceeds the 5 node free license")
	// ErrUnauthorized Unauthorized error
	ErrUnauthorized = errors.New("Unauthorized")
	// ErrResourceAccessDenied Access denied to resource error
	ErrResourceAccessDenied = errors.New("Access denied to resource")
	// ErrNotAvailableInDemo feature is not allowed in demo
	ErrNotAvailableInDemo = errors.New("This feature is not available in the demo version of Portainer")
	// ErrNotAvailable feature is not enabled
	ErrNotAvailable = errors.New("This feature is not enabled. Please contact your administrator.")
	// ErrDisabled feature is disabled
	ErrDisabled = errors.New("This feature has been disabled. Contact Portainer support.")
)

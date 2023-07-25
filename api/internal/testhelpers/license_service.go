package testhelpers

import (
	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
)

type Licenseservice struct{}

func (l Licenseservice) Info() portaineree.LicenseInfo           { return portaineree.LicenseInfo{Valid: true} }
func (l Licenseservice) Licenses() []liblicense.PortainerLicense { return nil }
func (l Licenseservice) AddLicense(licenseKey string, force bool) ([]string, error) {
	return nil, nil
}
func (l Licenseservice) DeleteLicense(licenseKey string) error { return nil }
func (l Licenseservice) Start() error                          { return nil }
func (l Licenseservice) ShouldEnforceOveruse() bool            { return false }
func (l Licenseservice) WillBeEnforcedAt() int64               { return 0 }
func (l Licenseservice) SyncLicenses() error                   { return nil }

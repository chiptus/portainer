package license

import (
	"testing"

	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/pkg/liblicense"

	"github.com/stretchr/testify/assert"
)

func Test_DeleteLicense(t *testing.T) {
	t.Run("should be able to delete revoked license", func(t *testing.T) {
		_, store := datastore.MustNewTestStore(t, true, true)

		license1 := &liblicense.PortainerLicense{
			LicenseKey: "key1",
			Revoked:    true,
		}
		store.License().AddLicense(license1.LicenseKey, license1)
		license2 := &liblicense.PortainerLicense{
			LicenseKey: "key2",
		}
		store.License().AddLicense(license1.LicenseKey, license2)

		s := NewService(store, nil, nil, false)
		err := s.DeleteLicense(license1.LicenseKey)
		assert.NoError(t, err)
	})
}

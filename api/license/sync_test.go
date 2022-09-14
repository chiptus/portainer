package license

import (
	"context"
	"testing"
	"time"

	"github.com/portainer/liblicense"
	"github.com/portainer/liblicense/master"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/stretchr/testify/assert"
)

func Test_revokeInvalidLicenses(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()

	invalidLicense := createValidLicense(liblicense.PortainerLicenseSubscription, time.Now().Unix())
	olderTrialLicense := createValidLicense(liblicense.PortainerLicenseTrial, time.Now().Add(-time.Hour*24).Unix())
	newerTrialLicense := createValidLicense(liblicense.PortainerLicenseTrial, time.Now().Unix())
	olderEssentialsLicense := createValidLicense(liblicense.PortainerLicenseEssentials, time.Now().Add(-time.Hour*24).Unix())
	newerEssentialsLicense := createValidLicense(liblicense.PortainerLicenseEssentials, time.Now().Unix())
	olderSubscriptionLicense := createValidLicense(liblicense.PortainerLicenseSubscription, time.Now().Add(-time.Hour*24).Unix())
	newerSubscriptionLicense := createValidLicense(liblicense.PortainerLicenseSubscription, time.Now().Unix())

	store.License().AddLicense(invalidLicense.LicenseKey, invalidLicense)
	store.License().AddLicense(olderTrialLicense.LicenseKey, olderTrialLicense)
	store.License().AddLicense(newerTrialLicense.LicenseKey, newerTrialLicense)
	store.License().AddLicense(olderEssentialsLicense.LicenseKey, olderEssentialsLicense)
	store.License().AddLicense(newerEssentialsLicense.LicenseKey, newerEssentialsLicense)
	store.License().AddLicense(olderSubscriptionLicense.LicenseKey, olderSubscriptionLicense)
	store.License().AddLicense(newerSubscriptionLicense.LicenseKey, newerSubscriptionLicense)

	service := NewService(store, context.Background())
	err := service.revokeInvalidLicenses(func(l *liblicense.PortainerLicense) (bool, error) {
		return l.LicenseKey != invalidLicense.LicenseKey, nil
	})
	assert.NoError(t, err)
	licenses, _ := service.Licenses()
	for _, l := range licenses {
		switch l.LicenseKey {
		case invalidLicense.LicenseKey:
			assert.True(t, l.Revoked)
		case olderTrialLicense.LicenseKey:
			assert.True(t, l.Revoked, "Trial licenses override each other, so older ones are revoked")
		case newerTrialLicense.LicenseKey:
			assert.False(t, l.Revoked, "Trial licenses override each other, so older ones are revoked")
		case olderEssentialsLicense.LicenseKey:
			assert.True(t, l.Revoked, "Essential licenses override each other, so older ones are revoked")
		case newerEssentialsLicense.LicenseKey:
			assert.False(t, l.Revoked, "Essential licenses override each other, so older ones are revoked")
		case olderSubscriptionLicense.LicenseKey:
			assert.False(t, l.Revoked, "Subscription licenses stack, so they should not be revoked")
		case newerSubscriptionLicense.LicenseKey:
			assert.False(t, l.Revoked, "Subscription licenses stack, so they should not be revoked")
		}
	}
}

func createValidLicense(licenseType liblicense.PortainerLicenseType, createdTimestamp int64) *liblicense.PortainerLicense {
	return createValidLicenseNodes(licenseType, createdTimestamp, 1)
}

func createValidLicenseNodes(licenseType liblicense.PortainerLicenseType, createdTimestamp int64, nodes int) *liblicense.PortainerLicense {
	license := &liblicense.PortainerLicense{
		Type:         licenseType,
		Created:      createdTimestamp,
		ExpiresAfter: 10, // expired in 10 days
		Nodes:        nodes,
	}
	key, _ := master.GenerateLicense(license)
	license.LicenseKey = key
	return license
}

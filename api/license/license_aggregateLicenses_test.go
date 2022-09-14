package license

import (
	"testing"
	"time"

	"github.com/portainer/liblicense"
	"github.com/portainer/liblicense/master"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func Test_aggregateLicense_CalculateCorrectType(t *testing.T) {
	cases := []struct {
		name         string
		types        []liblicense.PortainerLicenseType
		expectedType liblicense.PortainerLicenseType
	}{
		{
			name:         "no license",
			types:        []liblicense.PortainerLicenseType{},
			expectedType: liblicense.PortainerLicenseType(0),
		},
		{
			name:         "only Trials",
			types:        []liblicense.PortainerLicenseType{liblicense.PortainerLicenseTrial},
			expectedType: liblicense.PortainerLicenseTrial,
		},
		{
			name:         "only Essentials",
			types:        []liblicense.PortainerLicenseType{liblicense.PortainerLicenseEssentials},
			expectedType: liblicense.PortainerLicenseEssentials,
		},
		{
			name:         "only Subscriptions",
			types:        []liblicense.PortainerLicenseType{liblicense.PortainerLicenseSubscription},
			expectedType: liblicense.PortainerLicenseSubscription,
		},
		{
			name: "Essentials and Subscriptions result in Subscription",
			types: []liblicense.PortainerLicenseType{
				liblicense.PortainerLicenseEssentials,
				liblicense.PortainerLicenseSubscription},
			expectedType: liblicense.PortainerLicenseSubscription,
		},
		{
			name: "Trials and Subscriptions result in Trial",
			types: []liblicense.PortainerLicenseType{
				liblicense.PortainerLicenseTrial,
				liblicense.PortainerLicenseSubscription},
			expectedType: liblicense.PortainerLicenseTrial,
		},
		{
			name: "Essentials and Trials result in Trial",
			types: []liblicense.PortainerLicenseType{
				liblicense.PortainerLicenseEssentials,
				liblicense.PortainerLicenseTrial},
			expectedType: liblicense.PortainerLicenseTrial,
		},
		{
			name: "Essentials and Subscriptions and Trials result in Trial",
			types: []liblicense.PortainerLicenseType{
				liblicense.PortainerLicenseEssentials,
				liblicense.PortainerLicenseSubscription,
				liblicense.PortainerLicenseTrial},
			expectedType: liblicense.PortainerLicenseTrial,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			licenses := []liblicense.PortainerLicense{}
			for _, licenseType := range tc.types {
				l := &liblicense.PortainerLicense{
					Type:         licenseType,
					Created:      time.Now().Unix(),
					ExpiresAfter: 10,
				}
				updateLicenseWithGeneratedKey(l)
				licenses = append(licenses, *l)
			}

			_, store, teardown := datastore.MustNewTestStore(t, true, true)
			defer teardown()
			service := NewService(store, nil)

			result := service.aggregateLicenses(licenses)
			assert.Equal(t, tc.expectedType, result.Type)
		})
	}
}

func Test_aggregateLicenses_aggregatesValidLicenses(t *testing.T) {
	expiredLicense := liblicense.PortainerLicense{
		Created:      time.Now().Add(-time.Hour * 24 * 10).Unix(),
		ExpiresAfter: 1,
		Nodes:        1,
	}
	updateLicenseWithGeneratedKey(&expiredLicense)

	incorrectKeyLicense := liblicense.PortainerLicense{
		Created:      time.Now().Unix(),
		ExpiresAfter: 1,
		Revoked:      true,
		Nodes:        10,
		LicenseKey:   "foo",
	}

	validLicense1 := liblicense.PortainerLicense{
		Created:      time.Now().Unix(),
		ExpiresAfter: 1,
		Nodes:        50,
	}
	updateLicenseWithGeneratedKey(&validLicense1)

	validLicense2 := liblicense.PortainerLicense{
		Created:      time.Now().Unix(),
		ExpiresAfter: 1,
		Nodes:        100,
	}
	updateLicenseWithGeneratedKey(&validLicense2)

	licenses := []liblicense.PortainerLicense{
		expiredLicense,
		validLicense1,
		incorrectKeyLicense,
		validLicense2,
	}

	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()
	service := NewService(store, nil)

	result := service.aggregateLicenses(licenses)
	assert.Equal(t, 150, result.Nodes)
}

func Test_aggregateLicenses_picksEarliestExpirationDate(t *testing.T) {
	expiresFirst := liblicense.PortainerLicense{
		Created:      time.Now().Unix(),
		ExpiresAfter: 1,
	}
	updateLicenseWithGeneratedKey(&expiresFirst)

	expiresLast := liblicense.PortainerLicense{
		Created:      time.Now().Unix(),
		ExpiresAfter: 10,
	}
	updateLicenseWithGeneratedKey(&expiresLast)

	expiresSecond := liblicense.PortainerLicense{
		Created:      time.Now().Add(time.Hour * 24).Unix(),
		ExpiresAfter: 5,
	}
	updateLicenseWithGeneratedKey(&expiresSecond)

	licenses := []liblicense.PortainerLicense{expiresFirst, expiresLast, expiresSecond}

	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()
	service := NewService(store, nil)

	result := service.aggregateLicenses(licenses)
	assert.Equal(t, master.ExpiresAt(expiresFirst.Created, expiresFirst.ExpiresAfter).Unix(), result.ExpiresAt)
}

func Test_aggregateLicenses_shouldSetOveruseTimestamp(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()

	endpoint := &portaineree.Endpoint{Snapshots: []portainer.DockerSnapshot{{NodeCount: 10}}, Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
	store.Endpoint().Create(endpoint)

	service := NewService(store, nil)
	enforcement, _ := service.dataStore.Enforcement().LicenseEnforcement()
	assert.Equal(t, int64(0), enforcement.LicenseOveruseStartedTimestamp)

	overusedLicense := createValidLicenseNodes(liblicense.PortainerLicenseEssentials, time.Now().Unix(), 1)

	aggregate := service.aggregateLicenses([]liblicense.PortainerLicense{*overusedLicense})
	assert.NotZero(t, aggregate.OveruseStartedTimestamp)

	enforcement, _ = service.dataStore.Enforcement().LicenseEnforcement()
	assert.Equal(t, aggregate.OveruseStartedTimestamp, enforcement.LicenseOveruseStartedTimestamp)
}

func updateLicenseWithGeneratedKey(license *liblicense.PortainerLicense) *liblicense.PortainerLicense {
	key, _ := master.GenerateLicense(license)
	license.LicenseKey = key
	return license
}

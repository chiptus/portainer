package license

import (
	"testing"
	"time"

	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
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
			types:        []liblicense.PortainerLicenseType{liblicense.PortainerLicenseFree},
			expectedType: liblicense.PortainerLicenseFree,
		},
		{
			name:         "only Subscriptions",
			types:        []liblicense.PortainerLicenseType{liblicense.PortainerLicenseSubscription},
			expectedType: liblicense.PortainerLicenseSubscription,
		},
		{
			name: "Essentials and Subscriptions",
			types: []liblicense.PortainerLicenseType{
				liblicense.PortainerLicenseFree,
				liblicense.PortainerLicenseSubscription,
			},
			expectedType: liblicense.PortainerLicenseSubscription,
		},
		{
			name: "Trial and Subscriptions",
			types: []liblicense.PortainerLicenseType{
				liblicense.PortainerLicenseTrial,
				liblicense.PortainerLicenseSubscription,
			},
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
					Version:      3,
				}
				updateLicenseWithGeneratedKey(l)
				licenses = append(licenses, *l)
			}

			_, store := datastore.MustNewTestStore(t, true, true)
			service := NewService(store, nil, nil, false)

			result := service.aggregateLicenses(licenses)
			assert.Equal(t, tc.expectedType, result.Type)
		})
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
		Version:      3,
	}
	key, _ := liblicense.GenerateLicense(license)
	license.LicenseKey = key
	return license
}

func Test_aggregateLicenses_shouldSetOveruseTimestamp(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portainer.EndpointID(1)}
	store.Endpoint().Create(endpoint)
	store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})

	snapshotService, _ := snapshot.NewService("1s", store, nil, nil, nil, nil, nil)

	service := NewService(store, nil, snapshotService, false)
	enforcement, _ := service.dataStore.Enforcement().LicenseEnforcement()
	assert.Equal(t, int64(0), enforcement.LicenseOveruseStartedTimestamp)

	overusedLicense := createValidLicenseNodes(liblicense.PortainerLicenseFree, time.Now().Unix(), 1)

	aggregate := service.aggregateLicenses([]liblicense.PortainerLicense{*overusedLicense})
	assert.NotZero(t, aggregate.OveruseStartedTimestamp)

	enforcement, _ = service.dataStore.Enforcement().LicenseEnforcement()
	assert.Equal(t, aggregate.OveruseStartedTimestamp, enforcement.LicenseOveruseStartedTimestamp)
}

func updateLicenseWithGeneratedKey(license *liblicense.PortainerLicense) *liblicense.PortainerLicense {
	key, _ := liblicense.GenerateLicense(license)
	license.LicenseKey = key
	return license
}

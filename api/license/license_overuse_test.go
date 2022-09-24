package license

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/portainer/liblicense"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func Test_getLicenseOveruseTimestamp(t *testing.T) {
	t.Run("shouldn't update if aggregate license isn't Essential", func(t *testing.T) {
		_, store, teardown := datastore.MustNewTestStore(t, true, true)
		defer teardown()

		endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
		store.Endpoint().Create(endpoint)
		store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})

		service := NewService(store, nil, nil)

		enforcement, _ := service.dataStore.Enforcement().LicenseEnforcement()
		assert.Equal(t, int64(0), enforcement.LicenseOveruseStartedTimestamp)

		overuserTimestamp, err := service.getLicenseOveruseTimestamp(liblicense.PortainerLicenseSubscription, 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), overuserTimestamp, "should remain zero even though there are less licensed nodes than in use")
	})

	t.Run("should set the timestamp if Essential license was overused", func(t *testing.T) {
		_, store, teardown := datastore.MustNewTestStore(t, true, true)
		defer teardown()

		endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
		store.Endpoint().Create(endpoint)
		store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})

		snapshotService, _ := snapshot.NewService("1s", store, nil, nil, nil, nil)
		service := NewService(store, nil, snapshotService)

		enforcement, _ := service.dataStore.Enforcement().LicenseEnforcement()
		assert.Equal(t, int64(0), enforcement.LicenseOveruseStartedTimestamp)

		overuserTimestamp, err := service.getLicenseOveruseTimestamp(liblicense.PortainerLicenseEssentials, 1)
		assert.NoError(t, err)
		assert.NotZero(t, overuserTimestamp, "should be set when there are less licensed nodes than in use")
	})

	t.Run("shouldn't drop the timestamp if Essential license stopped being overused", func(t *testing.T) {
		_, store, teardown := datastore.MustNewTestStore(t, true, true)
		defer teardown()

		endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
		store.Endpoint().Create(endpoint)
		store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})

		snapshotService, _ := snapshot.NewService("1s", store, nil, nil, nil, nil)
		service := NewService(store, nil, snapshotService)
		originalValue := time.Now().Add(-time.Hour * 24).Unix()
		service.dataStore.Enforcement().UpdateOveruseStartedTimestamp(originalValue)

		overuserTimestamp, err := service.getLicenseOveruseTimestamp(liblicense.PortainerLicenseEssentials, 15)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), overuserTimestamp, "should drop to zero when there are more licensed nodes than in use")
	})

	t.Run("should keep the old timestamp if Essential license was overused before", func(t *testing.T) {
		_, store, teardown := datastore.MustNewTestStore(t, true, true)
		defer teardown()

		endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
		store.Endpoint().Create(endpoint)
		store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})

		snapshotService, _ := snapshot.NewService("1s", store, nil, nil, nil, nil)
		service := NewService(store, nil, snapshotService)
		originalTimestamp := time.Now().Add(-time.Hour * 24).Unix()
		service.dataStore.Enforcement().UpdateOveruseStartedTimestamp(originalTimestamp)

		overuserTimestamp, err := service.getLicenseOveruseTimestamp(liblicense.PortainerLicenseEssentials, 1)
		assert.NoError(t, err)
		assert.Equal(t, originalTimestamp, overuserTimestamp)
	})

	t.Run("should drop the timestamp if overused Essential license being replaced with Subscription", func(t *testing.T) {
		_, store, teardown := datastore.MustNewTestStore(t, true, true)
		defer teardown()

		endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
		store.Endpoint().Create(endpoint)
		store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})

		snapshotService, _ := snapshot.NewService("1s", store, nil, nil, nil, nil)
		service := NewService(store, nil, snapshotService)
		originalTimestamp := time.Now().Add(-time.Hour * 24).Unix()
		service.dataStore.Enforcement().UpdateOveruseStartedTimestamp(originalTimestamp)

		overuserTimestamp, err := service.getLicenseOveruseTimestamp(liblicense.PortainerLicenseSubscription, 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), overuserTimestamp)
	})
}

func Test_licenseIsOverused(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()

	endpoint1 := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
	endpoint2 := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(2)}
	store.Endpoint().Create(endpoint1)
	store.Endpoint().Create(endpoint2)

	store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint1.ID, Docker: &portainer.DockerSnapshot{NodeCount: 10}})
	store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint2.ID, Docker: &portainer.DockerSnapshot{NodeCount: 1}})

	snapshots, _ := store.Snapshot().Snapshots()

	assert.True(t, licenseIsOverused(5, snapshots))
	assert.True(t, licenseIsOverused(10, snapshots))
	assert.False(t, licenseIsOverused(11, snapshots))
	assert.False(t, licenseIsOverused(15, snapshots))
}

func Test_ShouldEnforceOveruse(t *testing.T) {
	service := NewService(nil, nil, nil)

	t.Run("should return false if licenseOveruseStartedTimestamp is empty", func(t *testing.T) {
		service.info = &portaineree.LicenseInfo{
			OveruseStartedTimestamp: int64(0),
		}
		assert.False(t, service.ShouldEnforceOveruse())
	})

	t.Run("should return false if licenseOveruseStartedTimestamp is set but grace period hasn't finished", func(t *testing.T) {
		service.info = &portaineree.LicenseInfo{
			OveruseStartedTimestamp: time.Now().Add(10 - overuseGracePeriodInSeconds).Unix(),
		}
		assert.False(t, service.ShouldEnforceOveruse())
	})

	t.Run("should return false if licenseOveruseStartedTimestamp is set and grace period lapsed", func(t *testing.T) {
		service.info = &portaineree.LicenseInfo{
			OveruseStartedTimestamp: time.Now().Add(-10 - overuseGracePeriodInSeconds).Unix(),
		}
		assert.False(t, service.ShouldEnforceOveruse())
	})
}

func Test_NotOverused(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(t, true, true)
	defer teardown()

	endpoint := &portaineree.Endpoint{Type: portaineree.DockerEnvironment, ID: portaineree.EndpointID(1)}
	store.Endpoint().Create(endpoint)
	store.Snapshot().Create(&portaineree.Snapshot{EndpointID: endpoint.ID, Docker: &portainer.DockerSnapshot{NodeCount: 5}})

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(``))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("should return http error if 5NF license is at capacity or over", func(t *testing.T) {
		w := httptest.NewRecorder()

		licenseService := NewService(nil, nil, nil)
		licenseService.info = &portaineree.LicenseInfo{Type: liblicense.PortainerLicenseEssentials, Nodes: 5}

		NotOverused(licenseService, store, nextHandler).ServeHTTP(w, r)
		assert.Equal(t, http.StatusPaymentRequired, w.Code)
	})

	t.Run("should pass request through if NON-5NF license is at capacity or over", func(t *testing.T) {
		w := httptest.NewRecorder()

		licenseService := NewService(nil, nil, nil)
		licenseService.info = &portaineree.LicenseInfo{Type: liblicense.PortainerLicenseSubscription, Nodes: 5}

		NotOverused(licenseService, store, nextHandler).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should pass request through if 5NF license is below capacity", func(t *testing.T) {
		w := httptest.NewRecorder()

		licenseService := NewService(nil, nil, nil)
		licenseService.info = &portaineree.LicenseInfo{Type: liblicense.PortainerLicenseEssentials, Nodes: 6}

		NotOverused(licenseService, store, nextHandler).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
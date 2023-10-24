package endpointedge

import (
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog"
)

func setupBuildEdgeStacksTest(b testing.TB, endpointsCount int) (*Handler, error) {
	_, store := datastore.MustNewTestStore(b, true, false)

	edgeStackID := portainer.EdgeStackID(1)

	edgeStack := &portaineree.EdgeStack{
		ID:      edgeStackID,
		Name:    "myEdgeStack",
		Status:  make(map[portainer.EndpointID]portainer.EdgeStackStatus),
		Version: 2,
	}

	err := store.EdgeStack().Create(edgeStackID, edgeStack)
	if err != nil {
		return nil, err
	}

	for i := 1; i < endpointsCount; i++ {
		endpointID := portainer.EndpointID(i)

		err = store.Endpoint().Create(&portaineree.Endpoint{
			ID:   endpointID,
			Name: "env-" + strconv.Itoa(i),
			Type: portaineree.EdgeAgentOnDockerEnvironment,
		})
		if err != nil {
			return nil, err
		}

		err = store.EndpointRelation().Create(&portainer.EndpointRelation{
			EndpointID: endpointID,
			EdgeStacks: map[portainer.EdgeStackID]bool{
				edgeStackID: true,
			},
		})
		if err != nil {
			return nil, err
		}

		edgeStack.Status[endpointID] = portainer.EdgeStackStatus{
			EndpointID: portainer.EndpointID(endpointID),
			Status: []portainer.EdgeStackDeploymentStatus{
				{
					Type: portainer.EdgeStackStatusDeploymentReceived,
				},
			},
		}

		err = store.EdgeStack().UpdateEdgeStack(edgeStackID, edgeStack, true)
		if err != nil {
			return nil, err
		}
	}

	edgeService := edgeasync.NewService(store, nil)

	h := NewHandler(testhelpers.NewTestRequestBouncer(), store, nil, nil, edgeService, nil, nil, nil)

	return h, nil
}

func BenchmarkSetupBuildEdgeStacksTest(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	var h *Handler
	for i := 0; i < b.N; i++ {
		var err error
		h, err = setupBuildEdgeStacksTest(b, endpointsCount)
		if err != nil {
			b.Fatal(err)
		}
	}

	runtime.KeepAlive(h)
}

func BenchmarkBuildEdgeStacksWithCache(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	h, err := setupBuildEdgeStacksTest(b, endpointsCount)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	skipCache := false
	for i := 0; i < b.N; i++ {
		resp, err := h.buildEdgeStacks(h.DataStore, portainer.EndpointID(1), time.UTC, &skipCache)
		if err != nil {
			b.Fatal(err)
		}

		count += len(resp)
	}

	runtime.KeepAlive(count)
}

func BenchmarkBuildEdgeStacksNoCache(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	h, err := setupBuildEdgeStacksTest(b, endpointsCount)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	skipCache := true
	for i := 0; i < b.N; i++ {
		resp, err := h.buildEdgeStacks(h.DataStore, portainer.EndpointID(1), time.UTC, &skipCache)
		if err != nil {
			b.Fatal(err)
		}

		count += len(resp)
	}

	runtime.KeepAlive(count)
}

func BenchmarkBuildEdgeStacksParallelWithCache(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	h, err := setupBuildEdgeStacksTest(b, endpointsCount)
	if err != nil {
		b.Fatal(err)
	}

	runtime.GOMAXPROCS(64)

	b.ReportAllocs()
	b.ResetTimer()

	count := &atomic.Int64{}
	skipCache := false
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := h.buildEdgeStacks(h.DataStore, portainer.EndpointID(1), time.UTC, &skipCache)
			if err != nil {
				b.Fatal(err)
			}

			count.Add(int64(len(resp)))
		}
	})

	runtime.KeepAlive(count)
}

func BenchmarkBuildEdgeStacksParallelNoCache(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	h, err := setupBuildEdgeStacksTest(b, endpointsCount)
	if err != nil {
		b.Fatal(err)
	}

	runtime.GOMAXPROCS(64)

	b.ReportAllocs()
	b.ResetTimer()

	count := &atomic.Int64{}
	skipCache := true
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := h.buildEdgeStacks(h.DataStore, portainer.EndpointID(1), time.UTC, &skipCache)
			if err != nil {
				b.Fatal(err)
			}

			count.Add(int64(len(resp)))
		}
	})

	runtime.KeepAlive(count)
}

package endpointedge

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog"
)

func setupBuildEdgeStacksTest(b testing.TB, endpointsCount int) (*Handler, error) {
	_, store := datastore.MustNewTestStore(b, true, false)

	edgeStackID := portaineree.EdgeStackID(1)

	edgeStack := &portaineree.EdgeStack{
		ID:      edgeStackID,
		Name:    "myEdgeStack",
		Status:  make(map[portaineree.EndpointID]portainer.EdgeStackStatus),
		Version: 2,
	}

	err := store.EdgeStack().Create(edgeStackID, edgeStack)
	if err != nil {
		return nil, err
	}

	for i := 1; i < endpointsCount; i++ {
		endpointID := portaineree.EndpointID(i)

		err = store.Endpoint().Create(&portaineree.Endpoint{
			ID:   endpointID,
			Name: "env-" + strconv.Itoa(i),
			Type: portaineree.EdgeAgentOnDockerEnvironment,
		})
		if err != nil {
			return nil, err
		}

		err = store.EndpointRelation().Create(&portaineree.EndpointRelation{
			EndpointID: endpointID,
			EdgeStacks: map[portaineree.EdgeStackID]bool{
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

	h := NewHandler(nil, store, nil, nil, edgeService, nil, nil, nil)

	return h, nil
}

func BenchmarkBuildEdgeStacks(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	h, err := setupBuildEdgeStacksTest(b, endpointsCount)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	skipCache := false
	for i := 0; i < b.N; i++ {
		h.buildEdgeStacks(h.DataStore, portaineree.EndpointID(1), time.UTC, &skipCache)
	}
}

func BenchmarkBuildEdgeStacksParallel(b *testing.B) {
	const endpointsCount = 2000

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	h, err := setupBuildEdgeStacksTest(b, endpointsCount)
	if err != nil {
		b.Fatal(err)
	}

	runtime.GOMAXPROCS(64)

	b.ReportAllocs()
	b.ResetTimer()

	skipCache := false
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.buildEdgeStacks(h.DataStore, portaineree.EndpointID(1), time.UTC, &skipCache)
		}
	})
}

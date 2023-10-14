package endpoints

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"

	"github.com/gofrs/uuid"
)

func TestConcurrentEndpointDelete(t *testing.T) {
	N := 100

	// Setup N environments with 1 shared tag

	handler := mustSetupGlobalKeyHandler(t)

	handler.demoService = demo.NewService()
	handler.ProxyManager = proxy.NewManager(handler.DataStore, nil, nil, nil, nil, nil, nil, nil, nil)

	tagID := portainer.TagID(1)

	err := handler.DataStore.Tag().Create(&portainer.Tag{
		ID:             tagID,
		Name:           "concurrent-test-tag",
		Endpoints:      make(map[portainer.EndpointID]bool),
		EndpointGroups: make(map[portainer.EndpointGroupID]bool),
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < N; i++ {
		p := &endpointCreatePayload{
			Name:                 uuid.Must(uuid.NewV4()).String(),
			URL:                  "https://portainer.io:9443",
			EndpointCreationType: edgeAgentEnvironment,
			GroupID:              1,
			TagIDs:               []portainer.TagID{tagID},
		}

		endpoint, hErr := handler.createEndpoint(handler.DataStore, p)
		if hErr != nil {
			t.Fatal(hErr)
		}

		tag, err := handler.DataStore.Tag().Read(tagID)
		if err != nil {
			t.Fatal(err)
		}

		tag.Endpoints[endpoint.ID] = true

		handler.DataStore.Tag().Update(tagID, tag)
	}

	// Delete the environments concurrently

	tag, err := handler.DataStore.Tag().Read(tagID)
	if err != nil {
		t.Fatal(err)
	}

	endpointIDs := tag.Endpoints

	var wg sync.WaitGroup
	wg.Add(len(endpointIDs))

	for k := range endpointIDs {
		go func(id portainer.EndpointID) {
			url := "https://portainer.io:9443/endpoints/" + strconv.Itoa(int(id))

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				panic(err)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				panic("unexpected status code")
			}

			wg.Done()
		}(k)
	}

	wg.Wait()

	tag, err = handler.DataStore.Tag().Read(tagID)
	if err != nil {
		t.Fatal(err)
	}

	if len(tag.Endpoints) > 0 {
		t.Fail()
	}
}

func TestEndpointDeleteEdgeGroupsConcurrently(t *testing.T) {
	const endpointsCount = 100

	_, store := datastore.MustNewTestStore(t, true, false)

	handler := NewHandler(
		testhelpers.NewTestRequestBouncer(),
		testhelpers.NewUserActivityService(),
		store,
		nil,
		demo.NewService(),
		nil,
		nil,
	)
	handler.ProxyManager = proxy.NewManager(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler.AuthorizationService = authorization.NewService(store)

	// Create all the environments and add them to the same edge group

	var endpointIDs []portainer.EndpointID

	for i := 0; i < endpointsCount; i++ {
		endpointID := portainer.EndpointID(i) + 1

		err := store.Endpoint().Create(&portaineree.Endpoint{
			ID:   endpointID,
			Name: "env-" + strconv.Itoa(int(endpointID)),
			Type: portaineree.EdgeAgentOnDockerEnvironment,
		})
		if err != nil {
			t.Fatal("could not create endpoint:", err)
		}

		endpointIDs = append(endpointIDs, endpointID)
	}

	err := store.EdgeGroup().Create(&portaineree.EdgeGroup{
		ID:        1,
		Name:      "edgegroup-1",
		Endpoints: endpointIDs,
	})
	if err != nil {
		t.Fatal("could not create edge group:", err)
	}

	// Remove the environments concurrently

	var wg sync.WaitGroup
	wg.Add(len(endpointIDs))

	for _, endpointID := range endpointIDs {
		go func(ID portainer.EndpointID) {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodDelete, "/endpoints/"+strconv.Itoa(int(ID)), nil)
			if err != nil {
				t.Fail()
				return
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}(endpointID)
	}

	wg.Wait()

	// Check that the edge group is consistent

	edgeGroup, err := handler.DataStore.EdgeGroup().Read(1)
	if err != nil {
		t.Fatal("could not retrieve the edge group:", err)
	}

	if len(edgeGroup.Endpoints) > 0 {
		t.Fatal("the edge group is not consistent")
	}
}

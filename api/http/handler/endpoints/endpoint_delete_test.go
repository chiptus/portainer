package endpoints

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/proxy"

	"github.com/gofrs/uuid"
)

func TestConcurrentEndpointDelete(t *testing.T) {
	N := 100

	// Setup N environments with 1 shared tag

	handler, teardown, err := setupGlobalKeyHandler(t)
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	handler.demoService = demo.NewService()
	handler.ProxyManager = proxy.NewManager(handler.DataStore, nil, nil, nil, nil, nil, nil, nil, nil)

	tagID := portaineree.TagID(1)

	err = handler.DataStore.Tag().Create(&portaineree.Tag{
		ID:             tagID,
		Name:           "concurrent-test-tag",
		Endpoints:      make(map[portaineree.EndpointID]bool),
		EndpointGroups: make(map[portaineree.EndpointGroupID]bool),
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
			TagIDs:               []portaineree.TagID{tagID},
			IsEdgeDevice:         true,
		}

		endpoint, hErr := handler.createEndpoint(p)
		if hErr != nil {
			t.Fatal(hErr)
		}

		tag, err := handler.DataStore.Tag().Tag(tagID)
		if err != nil {
			t.Fatal(err)
		}

		tag.Endpoints[endpoint.ID] = true

		handler.DataStore.Tag().UpdateTag(tagID, tag)
	}

	// Delete the environments concurrently

	tag, err := handler.DataStore.Tag().Tag(tagID)
	if err != nil {
		t.Fatal(err)
	}

	endpointIDs := tag.Endpoints

	var wg sync.WaitGroup
	wg.Add(len(endpointIDs))

	for k := range endpointIDs {
		go func(id portaineree.EndpointID) {
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

	tag, err = handler.DataStore.Tag().Tag(tagID)
	if err != nil {
		t.Fatal(err)
	}

	if len(tag.Endpoints) > 0 {
		t.Fail()
	}
}

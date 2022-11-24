package tags

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer-ee/api/jwt"
)

func TestTagDeleteEdgeGroupsConcurrently(t *testing.T) {
	const tagsCount = 100

	_, store, teardown := datastore.MustNewTestStore(t, true, false)
	defer teardown()

	user := &portaineree.User{ID: 2, Username: "admin", Role: portaineree.AdministratorRole}
	err := store.User().Create(user)
	if err != nil {
		t.Fatal("could not create admin user:", err)
	}

	jwtService, err := jwt.NewService("1h", store)
	if err != nil {
		t.Fatal("could not initialize the JWT service:", err)
	}

	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	rawAPIKey, _, err := apiKeyService.GenerateApiKey(*user, "test")
	if err != nil {
		t.Fatal("could not generate API key:", err)
	}

	bouncer := security.NewRequestBouncer(store, nil, jwtService, apiKeyService, nil)

	handler := NewHandler(bouncer, testhelpers.NewUserActivityService())
	handler.DataStore = store

	// Create all the tags and add them to the same edge group

	var tagIDs []portaineree.TagID

	for i := 0; i < tagsCount; i++ {
		tagID := portaineree.TagID(i) + 1

		err = store.Tag().Create(&portaineree.Tag{
			ID:   tagID,
			Name: "tag-" + strconv.Itoa(int(tagID)),
		})
		if err != nil {
			t.Fatal("could not create tag:", err)
		}

		tagIDs = append(tagIDs, tagID)
	}

	err = store.EdgeGroup().Create(&portaineree.EdgeGroup{
		ID:     1,
		Name:   "edgegroup-1",
		TagIDs: tagIDs,
	})
	if err != nil {
		t.Fatal("could not create edge group:", err)
	}

	// Remove the tags concurrently

	var wg sync.WaitGroup
	wg.Add(len(tagIDs))

	for _, tagID := range tagIDs {
		go func(ID portaineree.TagID) {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodDelete, "/tags/"+strconv.Itoa(int(ID)), nil)
			if err != nil {
				t.Fail()
				return
			}
			req.Header.Add("X-Api-Key", rawAPIKey)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}(tagID)
	}

	wg.Wait()

	// Check that the edge group is consistent

	edgeGroup, err := handler.DataStore.EdgeGroup().EdgeGroup(1)
	if err != nil {
		t.Fatal("could not retrieve the edge group:", err)
	}

	if len(edgeGroup.TagIDs) > 0 {
		t.Fatal("the edge group is not consistent")
	}
}

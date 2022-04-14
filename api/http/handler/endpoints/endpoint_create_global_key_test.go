package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer/api/filesystem"
)

func TestGlobalKey(t *testing.T) {
	_, store, teardown := datastore.MustNewTestStore(true, true)
	defer teardown()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "portainer-test-global-key-*")
	if err != nil {
		t.Fatal("could not create a tmp dir:", err)
	}

	fs, err := filesystem.NewService(tmpDir, "")
	if err != nil {
		t.Fatal("could not start a new filesystem service:", err)
	}

	handler := NewHandler(
		helper.NewTestRequestBouncer(),
		helper.NewUserActivityService(),
		store,
		edge.NewService(store, fs),
	)

	ctx := context.Background()
	shutdownCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	handler.ReverseTunnelService = chisel.NewService(store, shutdownCtx)
	handler.AuthorizationService = authorization.NewService(store)

	portainerURL := "portainer.io"

	doRequest := func() *endpointCreateGlobalKeyResponse {
		edgeID := "test-edge-id"

		req, err := http.NewRequest(http.MethodPost, "https://"+portainerURL+":9443/endpoints/global-key", nil)
		if err != nil {
			t.Fatal("request error:", err)
		}
		req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, edgeID)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatal("expected a 200 response, found:", rec.Code)
		}

		p := &endpointCreateGlobalKeyResponse{}
		err = json.NewDecoder(rec.Body).Decode(p)
		if err != nil {
			t.Fatal("could not decode the response:", err)
		}

		if p.EndpointID <= 0 {
			t.Fatal("received invalid EndpointID:", p.EndpointID)
		}

		endpoint, err := store.Endpoint().Endpoint(p.EndpointID)
		if err != nil {
			t.Fatal("could not retrieve the created endpoint:", err)
		}

		if endpoint.URL != portainerURL {
			t.Fatalf("expected the Portainer URL to be '%s', received '%s'", portainerURL, endpoint.URL)
		}

		return p
	}

	// Test non-existing endpoint
	resp1 := doRequest()

	// Test already existing endpoint
	resp2 := doRequest()

	if resp1.EndpointID != resp2.EndpointID {
		t.Fatalf("expected EndpointID = %d, received: %d", resp1.EndpointID, resp2.EndpointID)
	}
}

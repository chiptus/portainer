package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/chisel"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/portainer/portainer/api/filesystem"
)

func setupGlobalKeyHandler(t *testing.T) (*Handler, func(), error) {
	_, store, storeTeardown := datastore.MustNewTestStore(t, true, true)

	ctx := context.Background()
	shutdownCtx, cancelFn := context.WithCancel(ctx)

	teardown := func() {
		cancelFn()
		storeTeardown()
	}

	tmpDir, err := os.MkdirTemp(t.TempDir(), "portainer-test-global-key-*")
	if err != nil {
		teardown()
		return nil, nil, fmt.Errorf("could not create a tmp dir: %w", err)
	}

	fs, err := filesystem.NewService(tmpDir, "")
	if err != nil {
		teardown()
		return nil, nil, fmt.Errorf("could not start a new filesystem service: %w", err)
	}

	settings := &portaineree.Settings{EdgePortainerURL: "https://portainer.domain.tld:9443"}
	settings.Edge.TunnelServerAddress = "portainer.domain.tld:8000"
	err = store.Settings().UpdateSettings(settings)
	if err != nil {
		teardown()
		return nil, nil, fmt.Errorf("could not update settings: %w", err)
	}

	handler := NewHandler(
		helper.NewTestRequestBouncer(),
		helper.NewUserActivityService(),
		store,
		edgeasync.NewService(store, fs),
		nil,
		nil,
		nil,
	)

	handler.ReverseTunnelService = chisel.NewService(store, shutdownCtx)
	handler.AuthorizationService = authorization.NewService(store)

	return handler, teardown, nil
}

func TestGlobalKey(t *testing.T) {
	handler, teardown, err := setupGlobalKeyHandler(t)
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	portainerURL := "https://portainer.domain.tld:9443"

	doRequest := func() *endpointCreateGlobalKeyResponse {
		edgeID := "test-edge-id"

		req, err := http.NewRequest(http.MethodPost, portainerURL+"/endpoints/global-key", nil)
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

		endpoint, err := handler.DataStore.Endpoint().Endpoint(p.EndpointID)
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

func TestEmptyGlobalKey(t *testing.T) {
	handler, teardown, err := setupGlobalKeyHandler(t)
	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://portainer.io:9443/endpoints/global-key", nil)
	if err != nil {
		t.Fatal("request error:", err)
	}
	req.Header.Set(portaineree.PortainerAgentEdgeIDHeader, "")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatal("expected a 400 response, found:", rec.Code)
	}
}

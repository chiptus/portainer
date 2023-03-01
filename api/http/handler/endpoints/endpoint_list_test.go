package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

type endpointListTest struct {
	title    string
	expected []portaineree.EndpointID
}

func Test_EndpointList_AgentVersion(t *testing.T) {

	version1Endpoint := portaineree.Endpoint{
		ID:      1,
		GroupID: 1,
		Type:    portaineree.AgentOnDockerEnvironment,
		Agent: struct {
			Version string "example:\"1.0.0\""
		}{
			Version: "1.0.0",
		},
	}
	version2Endpoint := portaineree.Endpoint{ID: 2, GroupID: 1, Type: portaineree.AgentOnDockerEnvironment, Agent: struct {
		Version string "example:\"1.0.0\""
	}{Version: "2.0.0"}}
	noVersionEndpoint := portaineree.Endpoint{ID: 3, Type: portaineree.AgentOnDockerEnvironment, GroupID: 1}
	notAgentEnvironments := portaineree.Endpoint{ID: 4, Type: portaineree.DockerEnvironment, GroupID: 1}

	handler, teardown := setup(t, []portaineree.Endpoint{
		notAgentEnvironments,
		version1Endpoint,
		version2Endpoint,
		noVersionEndpoint,
	})

	defer teardown()

	type endpointListAgentVersionTest struct {
		endpointListTest
		filter []string
	}

	tests := []endpointListAgentVersionTest{
		{
			endpointListTest{
				"should show version 1 agent endpoints and non-agent endpoints",
				[]portaineree.EndpointID{version1Endpoint.ID, notAgentEnvironments.ID},
			},
			[]string{version1Endpoint.Agent.Version},
		},
		{
			endpointListTest{
				"should show version 2 endpoints and non-agent endpoints",
				[]portaineree.EndpointID{version2Endpoint.ID, notAgentEnvironments.ID},
			},
			[]string{version2Endpoint.Agent.Version},
		},
		{
			endpointListTest{
				"should show version 1 and 2 endpoints and non-agent endpoints",
				[]portaineree.EndpointID{version2Endpoint.ID, notAgentEnvironments.ID, version1Endpoint.ID},
			},
			[]string{version2Endpoint.Agent.Version, version1Endpoint.Agent.Version},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			is := assert.New(t)
			query := ""
			for _, filter := range test.filter {
				query += fmt.Sprintf("agentVersions[]=%s&", filter)
			}

			req := buildEndpointListRequest(query)

			resp, err := doEndpointListRequest(req, handler, is)
			is.NoError(err)

			is.Equal(len(test.expected), len(resp))

			respIds := []portaineree.EndpointID{}

			for _, endpoint := range resp {
				respIds = append(respIds, endpoint.ID)
			}

			is.ElementsMatch(test.expected, respIds)
		})
	}
}

func Test_endpointList_edgeFilter(t *testing.T) {

	trustedEdgeAsync := portaineree.Endpoint{ID: 1, UserTrusted: true, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: true}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	untrustedEdgeAsync := portaineree.Endpoint{ID: 2, UserTrusted: false, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: true}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularUntrustedEdgeStandard := portaineree.Endpoint{ID: 3, UserTrusted: false, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: false}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularTrustedEdgeStandard := portaineree.Endpoint{ID: 4, UserTrusted: true, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: false}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularEndpoint := portaineree.Endpoint{ID: 5, GroupID: 1, Type: portaineree.DockerEnvironment}

	handler, teardown := setup(t, []portaineree.Endpoint{
		trustedEdgeAsync,
		untrustedEdgeAsync,
		regularUntrustedEdgeStandard,
		regularTrustedEdgeStandard,
		regularEndpoint,
	})

	defer teardown()

	type endpointListEdgeTest struct {
		endpointListTest
		edgeAsync           *bool
		edgeDeviceUntrusted bool
	}

	tests := []endpointListEdgeTest{
		{
			endpointListTest: endpointListTest{
				"should show all endpoints expect of the untrusted devices",
				[]portaineree.EndpointID{trustedEdgeAsync.ID, regularTrustedEdgeStandard.ID, regularEndpoint.ID},
			},
		},
		{
			endpointListTest: endpointListTest{
				"should show only trusted edge async agents and regular endpoints",
				[]portaineree.EndpointID{trustedEdgeAsync.ID, regularEndpoint.ID},
			},
			edgeAsync: BoolAddr(true),
		},
		{
			endpointListTest: endpointListTest{
				"should show only untrusted edge devices and regular endpoints",
				[]portaineree.EndpointID{untrustedEdgeAsync.ID, regularEndpoint.ID},
			},
			edgeAsync:           BoolAddr(true),
			edgeDeviceUntrusted: true,
		},
		{
			endpointListTest: endpointListTest{
				"should show no edge devices",
				[]portaineree.EndpointID{regularEndpoint.ID, regularTrustedEdgeStandard.ID},
			},
			edgeAsync: BoolAddr(false),
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			is := assert.New(t)

			query := fmt.Sprintf("edgeDeviceUntrusted=%v&", test.edgeDeviceUntrusted)
			if test.edgeAsync != nil {
				query += fmt.Sprintf("edgeAsync=%v&", *test.edgeAsync)
			}

			req := buildEndpointListRequest(query)
			resp, err := doEndpointListRequest(req, handler, is)
			is.NoError(err)

			is.Equal(len(test.expected), len(resp))

			respIds := []portaineree.EndpointID{}

			for _, endpoint := range resp {
				respIds = append(respIds, endpoint.ID)
			}

			is.ElementsMatch(test.expected, respIds)
		})
	}
}

func setup(t *testing.T, endpoints []portaineree.Endpoint) (handler *Handler, teardown func()) {
	is := assert.New(t)
	_, store, teardown := datastore.MustNewTestStore(t, true, true)

	for _, endpoint := range endpoints {
		err := store.Endpoint().Create(&endpoint)
		is.NoError(err, "error creating environment")
	}

	err := store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	bouncer := helper.NewTestRequestBouncer()
	handler = NewHandler(bouncer, helper.NewUserActivityService(), store, nil, nil, nil, nil)
	handler.ComposeStackManager = testhelpers.NewComposeStackManager()

	handler.SnapshotService, _ = snapshot.NewService("1s", store, nil, nil, nil, nil)

	return handler, teardown
}

func buildEndpointListRequest(query string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/endpoints?%s", query), nil)

	ctx := security.StoreTokenData(req, &portaineree.TokenData{ID: 1, Username: "admin", Role: 1})
	req = req.WithContext(ctx)

	restrictedCtx := security.StoreRestrictedRequestContext(req, &security.RestrictedRequestContext{UserID: 1, IsAdmin: true})
	req = req.WithContext(restrictedCtx)

	req.Header.Add("Authorization", "Bearer dummytoken")

	return req
}

func doEndpointListRequest(req *http.Request, h *Handler, is *assert.Assertions) ([]portaineree.Endpoint, error) {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	is.Equal(http.StatusOK, rr.Code, "Status should be 200")
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		return nil, err
	}

	resp := []portaineree.Endpoint{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

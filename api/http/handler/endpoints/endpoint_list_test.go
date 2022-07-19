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
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

type endpointListTest struct {
	title    string
	expected []portaineree.EndpointID
}

func Test_endpointList_edgeDeviceFilter(t *testing.T) {

	trustedEdgeDevice := portaineree.Endpoint{ID: 1, UserTrusted: true, IsEdgeDevice: true, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	untrustedEdgeDevice := portaineree.Endpoint{ID: 2, UserTrusted: false, IsEdgeDevice: true, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularUntrustedEdgeEndpoint := portaineree.Endpoint{ID: 3, UserTrusted: false, IsEdgeDevice: false, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularTrustedEdgeEndpoint := portaineree.Endpoint{ID: 4, UserTrusted: true, IsEdgeDevice: false, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularEndpoint := portaineree.Endpoint{ID: 5, UserTrusted: false, IsEdgeDevice: false, GroupID: 1, Type: portaineree.DockerEnvironment}

	handler, teardown := setup(t, []portaineree.Endpoint{
		trustedEdgeDevice,
		untrustedEdgeDevice,
		regularUntrustedEdgeEndpoint,
		regularTrustedEdgeEndpoint,
		regularEndpoint,
	})

	defer teardown()

	type endpointListEdgeDeviceTest struct {
		endpointListTest
		edgeDevice          *bool
		edgeDeviceUntrusted bool
	}

	tests := []endpointListEdgeDeviceTest{
		{
			endpointListTest: endpointListTest{
				"should show all endpoints expect of the untrusted devices",
				[]portaineree.EndpointID{trustedEdgeDevice.ID, regularUntrustedEdgeEndpoint.ID, regularTrustedEdgeEndpoint.ID, regularEndpoint.ID},
			},
			edgeDevice: nil,
		},
		{
			endpointListTest: endpointListTest{
				"should show only trusted edge devices and regular endpoints",
				[]portaineree.EndpointID{trustedEdgeDevice.ID, regularEndpoint.ID},
			},
			edgeDevice: BoolAddr(true),
		},
		{
			endpointListTest: endpointListTest{
				"should show only untrusted edge devices and regular endpoints",
				[]portaineree.EndpointID{untrustedEdgeDevice.ID, regularEndpoint.ID},
			},
			edgeDevice:          BoolAddr(true),
			edgeDeviceUntrusted: true,
		},
		{
			endpointListTest: endpointListTest{
				"should show no edge devices",
				[]portaineree.EndpointID{regularEndpoint.ID, regularUntrustedEdgeEndpoint.ID, regularTrustedEdgeEndpoint.ID},
			},
			edgeDevice: BoolAddr(false),
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			is := assert.New(t)

			query := fmt.Sprintf("edgeDeviceUntrusted=%v&", test.edgeDeviceUntrusted)
			if test.edgeDevice != nil {
				query += fmt.Sprintf("edgeDevice=%v&", *test.edgeDevice)
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
	_, store, teardown := datastore.MustNewTestStore(true, true)

	for _, endpoint := range endpoints {
		err := store.Endpoint().Create(&endpoint)
		is.NoError(err, "error creating environment")
	}

	err := store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	bouncer := helper.NewTestRequestBouncer()
	handler = NewHandler(bouncer, helper.NewUserActivityService(), store, nil, nil, nil)
	handler.ComposeStackManager = testhelpers.NewComposeStackManager()

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

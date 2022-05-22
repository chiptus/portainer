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

type endpointListEdgeDeviceTest struct {
	title    string
	expected []portaineree.EndpointID
	filter   string
}

func Test_endpointList(t *testing.T) {
	var err error
	is := assert.New(t)

	_, store, teardown := datastore.MustNewTestStore(true, true)
	defer teardown()

	trustedEndpoint := portaineree.Endpoint{ID: 1, UserTrusted: true, IsEdgeDevice: true, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	untrustedEndpoint := portaineree.Endpoint{ID: 2, UserTrusted: false, IsEdgeDevice: true, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularUntrustedEdgeEndpoint := portaineree.Endpoint{ID: 3, UserTrusted: false, IsEdgeDevice: false, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularTrustedEdgeEndpoint := portaineree.Endpoint{ID: 4, UserTrusted: true, IsEdgeDevice: false, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularEndpoint := portaineree.Endpoint{ID: 5, UserTrusted: false, IsEdgeDevice: false, GroupID: 1, Type: portaineree.DockerEnvironment}

	endpoints := []portaineree.Endpoint{
		trustedEndpoint,
		untrustedEndpoint,
		regularUntrustedEdgeEndpoint,
		regularTrustedEdgeEndpoint,
		regularEndpoint,
	}

	for _, endpoint := range endpoints {
		err = store.Endpoint().Create(&endpoint)
		is.NoError(err, "error creating environment")
	}

	err = store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	bouncer := helper.NewTestRequestBouncer()
	h := NewHandler(bouncer, helper.NewUserActivityService(), store, nil, nil)
	h.ComposeStackManager = testhelpers.NewComposeStackManager()

	tests := []endpointListEdgeDeviceTest{
		{
			"should show all edge endpoints",
			[]portaineree.EndpointID{trustedEndpoint.ID, untrustedEndpoint.ID, regularUntrustedEdgeEndpoint.ID, regularTrustedEdgeEndpoint.ID},
			EdgeDeviceFilterAll,
		},
		{
			"should show only trusted edge devices",
			[]portaineree.EndpointID{trustedEndpoint.ID, regularTrustedEdgeEndpoint.ID},
			EdgeDeviceFilterTrusted,
		},
		{
			"should show only untrusted edge devices",
			[]portaineree.EndpointID{untrustedEndpoint.ID, regularUntrustedEdgeEndpoint.ID},
			EdgeDeviceFilterUntrusted,
		},
		{
			"should show no edge devices",
			[]portaineree.EndpointID{regularEndpoint.ID, regularUntrustedEdgeEndpoint.ID, regularTrustedEdgeEndpoint.ID},
			EdgeDeviceFilterNone,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			is := assert.New(t)

			req := buildEndpointListRequest(test.filter)
			resp, err := doEndpointListRequest(req, h, is)
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

func buildEndpointListRequest(filter string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/endpoints?edgeDeviceFilter=%s", filter), nil)

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

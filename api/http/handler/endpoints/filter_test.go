package endpoints

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	helper "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

type filterTest struct {
	title    string
	expected []portaineree.EndpointID
	query    EnvironmentsQuery
}

func Test_Filter_edgeDeviceFilter(t *testing.T) {

	trustedEdgeDevice := portaineree.Endpoint{ID: 1, UserTrusted: true, IsEdgeDevice: true, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	untrustedEdgeDevice := portaineree.Endpoint{ID: 2, UserTrusted: false, IsEdgeDevice: true, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularUntrustedEdgeEndpoint := portaineree.Endpoint{ID: 3, UserTrusted: false, IsEdgeDevice: false, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularTrustedEdgeEndpoint := portaineree.Endpoint{ID: 4, UserTrusted: true, IsEdgeDevice: false, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularEndpoint := portaineree.Endpoint{ID: 5, GroupID: 1, Type: portaineree.DockerEnvironment}

	endpoints := []portaineree.Endpoint{
		trustedEdgeDevice,
		untrustedEdgeDevice,
		regularUntrustedEdgeEndpoint,
		regularTrustedEdgeEndpoint,
		regularEndpoint,
	}

	handler, teardown := setupFilterTest(t, endpoints)

	defer teardown()

	tests := []filterTest{
		{
			"should show all edge endpoints except of the untrusted devices",
			[]portaineree.EndpointID{trustedEdgeDevice.ID, regularUntrustedEdgeEndpoint.ID, regularTrustedEdgeEndpoint.ID},
			EnvironmentsQuery{
				types: []portaineree.EndpointType{portaineree.EdgeAgentOnDockerEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.EdgeAgentOnNomadEnvironment},
			},
		},
		{
			"should show only trusted edge devices and other regular endpoints",
			[]portaineree.EndpointID{trustedEdgeDevice.ID, regularEndpoint.ID},
			EnvironmentsQuery{
				edgeDevice: BoolAddr(true),
			},
		},
		{
			"should show only untrusted edge devices and other regular endpoints",
			[]portaineree.EndpointID{untrustedEdgeDevice.ID, regularEndpoint.ID},
			EnvironmentsQuery{
				edgeDevice:          BoolAddr(true),
				edgeDeviceUntrusted: true,
			},
		},
		{
			"should show no edge devices",
			[]portaineree.EndpointID{regularEndpoint.ID, regularUntrustedEdgeEndpoint.ID, regularTrustedEdgeEndpoint.ID},
			EnvironmentsQuery{
				edgeDevice: BoolAddr(false),
			},
		},
	}

	runTests(tests, t, handler, endpoints)
}

func runTests(tests []filterTest, t *testing.T, handler *Handler, endpoints []portaineree.Endpoint) {
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			runTest(t, test, handler, endpoints)
		})
	}
}

func runTest(t *testing.T, test filterTest, handler *Handler, endpoints []portaineree.Endpoint) {
	is := assert.New(t)

	filteredEndpoints, _, err := handler.filterEndpointsByQuery(endpoints, test.query, []portaineree.EndpointGroup{}, &portaineree.Settings{})

	is.NoError(err)

	is.Equal(len(test.expected), len(filteredEndpoints))

	respIds := []portaineree.EndpointID{}

	for _, endpoint := range filteredEndpoints {
		respIds = append(respIds, endpoint.ID)
	}

	is.ElementsMatch(test.expected, respIds)

}

func setupFilterTest(t *testing.T, endpoints []portaineree.Endpoint) (handler *Handler, teardown func()) {
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

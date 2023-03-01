package endpoints

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

type filterTest struct {
	title    string
	expected []portaineree.EndpointID
	query    EnvironmentsQuery
}

func Test_Filter_AgentVersion(t *testing.T) {

	version1Endpoint := portaineree.Endpoint{ID: 1, GroupID: 1,
		Type: portaineree.AgentOnDockerEnvironment,
		Agent: struct {
			Version string "example:\"1.0.0\""
		}{Version: "1.0.0"}}
	version2Endpoint := portaineree.Endpoint{ID: 2, GroupID: 1,
		Type: portaineree.AgentOnDockerEnvironment,
		Agent: struct {
			Version string "example:\"1.0.0\""
		}{Version: "2.0.0"}}
	noVersionEndpoint := portaineree.Endpoint{ID: 3, GroupID: 1,
		Type: portaineree.AgentOnDockerEnvironment,
	}
	notAgentEnvironments := portaineree.Endpoint{ID: 4, Type: portaineree.DockerEnvironment, GroupID: 1}

	endpoints := []portaineree.Endpoint{
		version1Endpoint,
		version2Endpoint,
		noVersionEndpoint,
		notAgentEnvironments,
	}

	handler, teardown := setupFilterTest(t, endpoints)

	defer teardown()

	tests := []filterTest{
		{
			"should show version 1 endpoints",
			[]portaineree.EndpointID{version1Endpoint.ID},
			EnvironmentsQuery{
				agentVersions: []string{version1Endpoint.Agent.Version},
				types:         []portaineree.EndpointType{portaineree.AgentOnDockerEnvironment},
			},
		},
		{
			"should show version 2 endpoints",
			[]portaineree.EndpointID{version2Endpoint.ID},
			EnvironmentsQuery{
				agentVersions: []string{version2Endpoint.Agent.Version},
				types:         []portaineree.EndpointType{portaineree.AgentOnDockerEnvironment},
			},
		},
		{
			"should show version 1 and 2 endpoints",
			[]portaineree.EndpointID{version2Endpoint.ID, version1Endpoint.ID},
			EnvironmentsQuery{
				agentVersions: []string{version2Endpoint.Agent.Version, version1Endpoint.Agent.Version},
				types:         []portaineree.EndpointType{portaineree.AgentOnDockerEnvironment},
			},
		},
	}

	runTests(tests, t, handler, endpoints)
}

func Test_Filter_edgeFilter(t *testing.T) {

	trustedEdgeAsync := portaineree.Endpoint{ID: 1, UserTrusted: true, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: true}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	untrustedEdgeAsync := portaineree.Endpoint{ID: 2, UserTrusted: false, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: true}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularUntrustedEdgeStandard := portaineree.Endpoint{ID: 3, UserTrusted: false, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: false}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularTrustedEdgeStandard := portaineree.Endpoint{ID: 4, UserTrusted: true, Edge: portaineree.EnvironmentEdgeSettings{AsyncMode: false}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularEndpoint := portaineree.Endpoint{ID: 5, GroupID: 1, Type: portaineree.DockerEnvironment}

	endpoints := []portaineree.Endpoint{
		trustedEdgeAsync,
		untrustedEdgeAsync,
		regularUntrustedEdgeStandard,
		regularTrustedEdgeStandard,
		regularEndpoint,
	}

	handler, teardown := setupFilterTest(t, endpoints)

	defer teardown()

	tests := []filterTest{
		{
			"should show all edge endpoints except of the untrusted edge",
			[]portaineree.EndpointID{trustedEdgeAsync.ID, regularTrustedEdgeStandard.ID},
			EnvironmentsQuery{
				types: []portaineree.EndpointType{portaineree.EdgeAgentOnDockerEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.EdgeAgentOnNomadEnvironment},
			},
		},
		{
			"should show only trusted edge devices and other regular endpoints",
			[]portaineree.EndpointID{trustedEdgeAsync.ID, regularEndpoint.ID},
			EnvironmentsQuery{
				edgeAsync: BoolAddr(true),
			},
		},
		{
			"should show only untrusted edge devices and other regular endpoints",
			[]portaineree.EndpointID{untrustedEdgeAsync.ID, regularEndpoint.ID},
			EnvironmentsQuery{
				edgeAsync:           BoolAddr(true),
				edgeDeviceUntrusted: true,
			},
		},
		{
			"should show no edge devices",
			[]portaineree.EndpointID{regularEndpoint.ID, regularTrustedEdgeStandard.ID},
			EnvironmentsQuery{
				edgeAsync: BoolAddr(false),
			},
		},
	}

	runTests(tests, t, handler, endpoints)
}

func runTests(tests []filterTest, t *testing.T, handler *Handler, endpoints []portaineree.Endpoint) {
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			runTest(t, test, handler, append([]portaineree.Endpoint{}, endpoints...))
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
	_, store, teardown := datastore.MustNewTestStore(t, true, true)

	for _, endpoint := range endpoints {
		err := store.Endpoint().Create(&endpoint)
		is.NoError(err, "error creating environment")
	}

	err := store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	bouncer := testhelpers.NewTestRequestBouncer()
	handler = NewHandler(bouncer, testhelpers.NewUserActivityService(), store, nil, nil, nil, nil)
	handler.ComposeStackManager = testhelpers.NewComposeStackManager()

	return handler, teardown
}

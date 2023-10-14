package endpoints

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

type filterTest struct {
	title    string
	expected []portainer.EndpointID
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

	handler := setupFilterTest(t, endpoints)

	tests := []filterTest{
		{
			"should show version 1 endpoints",
			[]portainer.EndpointID{version1Endpoint.ID},
			EnvironmentsQuery{
				agentVersions: []string{version1Endpoint.Agent.Version},
				types:         []portainer.EndpointType{portaineree.AgentOnDockerEnvironment},
			},
		},
		{
			"should show version 2 endpoints",
			[]portainer.EndpointID{version2Endpoint.ID},
			EnvironmentsQuery{
				agentVersions: []string{version2Endpoint.Agent.Version},
				types:         []portainer.EndpointType{portaineree.AgentOnDockerEnvironment},
			},
		},
		{
			"should show version 1 and 2 endpoints",
			[]portainer.EndpointID{version2Endpoint.ID, version1Endpoint.ID},
			EnvironmentsQuery{
				agentVersions: []string{version2Endpoint.Agent.Version, version1Endpoint.Agent.Version},
				types:         []portainer.EndpointType{portaineree.AgentOnDockerEnvironment},
			},
		},
	}

	runTests(tests, t, handler, endpoints)
}

func Test_Filter_edgeFilter(t *testing.T) {

	trustedEdgeAsync := portaineree.Endpoint{ID: 1, UserTrusted: true, Edge: portainer.EnvironmentEdgeSettings{AsyncMode: true}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	untrustedEdgeAsync := portaineree.Endpoint{ID: 2, UserTrusted: false, Edge: portainer.EnvironmentEdgeSettings{AsyncMode: true}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularUntrustedEdgeStandard := portaineree.Endpoint{ID: 3, UserTrusted: false, Edge: portainer.EnvironmentEdgeSettings{AsyncMode: false}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularTrustedEdgeStandard := portaineree.Endpoint{ID: 4, UserTrusted: true, Edge: portainer.EnvironmentEdgeSettings{AsyncMode: false}, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment}
	regularEndpoint := portaineree.Endpoint{ID: 5, GroupID: 1, Type: portaineree.DockerEnvironment}

	endpoints := []portaineree.Endpoint{
		trustedEdgeAsync,
		untrustedEdgeAsync,
		regularUntrustedEdgeStandard,
		regularTrustedEdgeStandard,
		regularEndpoint,
	}

	handler := setupFilterTest(t, endpoints)

	tests := []filterTest{
		{
			"should show all edge endpoints except of the untrusted edge",
			[]portainer.EndpointID{trustedEdgeAsync.ID, regularTrustedEdgeStandard.ID},
			EnvironmentsQuery{
				types: []portainer.EndpointType{portaineree.EdgeAgentOnDockerEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.EdgeAgentOnNomadEnvironment},
			},
		},
		{
			"should show only trusted edge devices and other regular endpoints",
			[]portainer.EndpointID{trustedEdgeAsync.ID, regularEndpoint.ID},
			EnvironmentsQuery{
				edgeAsync: BoolAddr(true),
			},
		},
		{
			"should show only untrusted edge devices and other regular endpoints",
			[]portainer.EndpointID{untrustedEdgeAsync.ID, regularEndpoint.ID},
			EnvironmentsQuery{
				edgeAsync:           BoolAddr(true),
				edgeDeviceUntrusted: true,
			},
		},
		{
			"should show no edge devices",
			[]portainer.EndpointID{regularEndpoint.ID, regularTrustedEdgeStandard.ID},
			EnvironmentsQuery{
				edgeAsync: BoolAddr(false),
			},
		},
	}

	runTests(tests, t, handler, endpoints)
}

func Test_Filter_excludeIDs(t *testing.T) {
	ids := []portainer.EndpointID{1, 2, 3, 4, 5, 6, 7, 8, 9}

	environments := slices.Map(ids, func(id portainer.EndpointID) portaineree.Endpoint {
		return portaineree.Endpoint{ID: id, GroupID: 1, Type: portaineree.DockerEnvironment}
	})

	handler := setupFilterTest(t, environments)

	tests := []filterTest{
		{
			title:    "should exclude IDs 2,5,8",
			expected: []portainer.EndpointID{1, 3, 4, 6, 7, 9},
			query: EnvironmentsQuery{
				excludeIds: []portainer.EndpointID{2, 5, 8},
			},
		},
	}

	runTests(tests, t, handler, environments)
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

	filteredEndpoints, _, err := handler.filterEndpointsByQuery(
		endpoints,
		test.query,
		[]portainer.EndpointGroup{},
		[]portaineree.EdgeGroup{},
		&portaineree.Settings{},
	)

	is.NoError(err)

	is.Equal(len(test.expected), len(filteredEndpoints))

	respIds := []portainer.EndpointID{}

	for _, endpoint := range filteredEndpoints {
		respIds = append(respIds, endpoint.ID)
	}

	is.ElementsMatch(test.expected, respIds)

}

func setupFilterTest(t *testing.T, endpoints []portaineree.Endpoint) *Handler {
	is := assert.New(t)
	_, store := datastore.MustNewTestStore(t, true, true)

	for _, endpoint := range endpoints {
		err := store.Endpoint().Create(&endpoint)
		is.NoError(err, "error creating environment")
	}

	err := store.User().Create(&portaineree.User{Username: "admin", Role: portaineree.AdministratorRole})
	is.NoError(err, "error creating a user")

	bouncer := testhelpers.NewTestRequestBouncer()
	handler := NewHandler(bouncer, testhelpers.NewUserActivityService(), store, nil, nil, nil, nil)
	handler.ComposeStackManager = testhelpers.NewComposeStackManager()

	return handler
}

package endpoints

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/edge"
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

	handler := setupFilterTest(t, endpoints)

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

	handler := setupFilterTest(t, endpoints)

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

func Test_Filter_edgeGroups(t *testing.T) {

	// create 6 endpoints
	env1_1 := portaineree.Endpoint{ID: 1, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment, UserTrusted: true}
	env1_2 := portaineree.Endpoint{ID: 2, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment, UserTrusted: true, TagIDs: []portaineree.TagID{1}}
	env1_3 := portaineree.Endpoint{ID: 3, GroupID: 1, Type: portaineree.EdgeAgentOnDockerEnvironment, UserTrusted: true, TagIDs: []portaineree.TagID{4, 5}}
	env2_2 := portaineree.Endpoint{ID: 4, GroupID: 2, Type: portaineree.EdgeAgentOnDockerEnvironment, UserTrusted: true, TagIDs: []portaineree.TagID{2}}
	env2_3 := portaineree.Endpoint{ID: 5, GroupID: 2, Type: portaineree.EdgeAgentOnDockerEnvironment, UserTrusted: true, TagIDs: []portaineree.TagID{3, 4, 5}}
	env3_3 := portaineree.Endpoint{ID: 6, GroupID: 3, Type: portaineree.EdgeAgentOnDockerEnvironment, UserTrusted: true, TagIDs: []portaineree.TagID{4, 5}}

	endpoints := []portaineree.Endpoint{
		env1_1,
		env1_2,
		env1_3,
		env2_2,
		env2_3,
		env3_3,
	}

	handler := setupFilterTest(t, endpoints)

	// create 5 tags
	tags := []portaineree.Tag{
		{ID: 1, Name: "tag1", Endpoints: map[portaineree.EndpointID]bool{env1_2.ID: true}, EndpointGroups: make(map[portaineree.EndpointGroupID]bool)},
		{ID: 2, Name: "tag2", Endpoints: map[portaineree.EndpointID]bool{env1_2.ID: true, env2_2.ID: true}, EndpointGroups: make(map[portaineree.EndpointGroupID]bool)},
		{ID: 3, Name: "tag3", Endpoints: map[portaineree.EndpointID]bool{env3_3.ID: true}, EndpointGroups: make(map[portaineree.EndpointGroupID]bool)},
		{ID: 4, Name: "tag4", Endpoints: map[portaineree.EndpointID]bool{env1_3.ID: true, env2_3.ID: true, env3_3.ID: true}, EndpointGroups: make(map[portaineree.EndpointGroupID]bool)},
		{ID: 5, Name: "tag5", Endpoints: map[portaineree.EndpointID]bool{env1_3.ID: true, env2_3.ID: true, env3_3.ID: true}, EndpointGroups: make(map[portaineree.EndpointGroupID]bool)},
	}

	for _, tag := range tags {
		err := handler.DataStore.Tag().Create(&tag)
		assert.NoError(t, err)
	}

	// create 3 edge groups, one static, one dynamic with partial match, one dynamic with full match, each of them should have 3 endpoints, one is shared with one of the other groups, one is shared with the other group and one is unique
	edgeGroup1 := portaineree.EdgeGroup{ID: 1, Name: "edgeGroup1", Dynamic: false, Endpoints: []portaineree.EndpointID{env1_1.ID, env1_2.ID, env1_3.ID}}
	edgeGroup2 := portaineree.EdgeGroup{ID: 2, Name: "edgeGroup2", Dynamic: true, PartialMatch: true, TagIDs: []portaineree.TagID{1, 2, 3}} // should match env1_2 and env2_2 and env2_3
	edgeGroup3 := portaineree.EdgeGroup{ID: 3, Name: "edgeGroup3", Dynamic: true, PartialMatch: false, TagIDs: []portaineree.TagID{4, 5}}   // should match env1_3 and env2_3 and env3_3

	edgeGroups := []portaineree.EdgeGroup{
		edgeGroup1,
		edgeGroup2,
		edgeGroup3,
	}

	expectedEdgeGroupEndpoints := map[portaineree.EdgeGroupID][]portaineree.EndpointID{
		1: {env1_1.ID, env1_2.ID, env1_3.ID},
		2: {env1_2.ID, env2_2.ID, env2_3.ID},
		3: {env1_3.ID, env2_3.ID, env3_3.ID},
	}

	for _, edgeGroup := range edgeGroups {
		err := handler.DataStore.EdgeGroup().Create(&edgeGroup)
		assert.NoError(t, err)
	}

	// check that the endpoints are related correctly:

	for _, edgeGroup := range edgeGroups {
		edgeGroupEndpoints := edge.EdgeGroupRelatedEndpoints(&edgeGroup, endpoints, []portaineree.EndpointGroup{{ID: 1, TagIDs: []portaineree.TagID{}}})
		assert.Equal(t, expectedEdgeGroupEndpoints[edgeGroup.ID], edgeGroupEndpoints, "edge group %d", edgeGroup.ID)
	}

	tests := []filterTest{
		{
			"should show all endpoints",
			[]portaineree.EndpointID{env1_1.ID, env1_2.ID, env1_3.ID, env2_2.ID, env2_3.ID, env3_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{1, 2, 3},
			},
		},
		{
			"should show only endpoints from edge group 1",
			[]portaineree.EndpointID{env1_1.ID, env1_2.ID, env1_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{1},
			},
		},
		{
			"should show only endpoints from edge group 2",
			[]portaineree.EndpointID{env1_2.ID, env2_2.ID, env2_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{2},
			},
		},
		{
			"should show only endpoints from edge group 3",
			[]portaineree.EndpointID{env1_3.ID, env2_3.ID, env3_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{3},
			},
		},
		{
			"should show only endpoints from edge group 1 and 2",
			[]portaineree.EndpointID{env1_1.ID, env1_2.ID, env1_3.ID, env2_2.ID, env2_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{1, 2},
			},
		},
		{
			"should show only endpoints from edge group 1 and 3",
			[]portaineree.EndpointID{env1_1.ID, env1_2.ID, env1_3.ID, env2_3.ID, env3_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{1, 3},
			},
		},
		{
			"should show only endpoints from edge group 2 and 3",
			[]portaineree.EndpointID{env1_2.ID, env2_2.ID, env2_3.ID, env3_3.ID, env1_3.ID},
			EnvironmentsQuery{
				edgeGroupIds: []portaineree.EdgeGroupID{2, 3},
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

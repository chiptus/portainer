package endpoints

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/stretchr/testify/assert"
)

func Test_updateEdgeGroups(t *testing.T) {

	createGroups := func(store *datastore.Store, names []string) ([]portaineree.EdgeGroup, error) {
		groups := make([]portaineree.EdgeGroup, len(names))
		for index, name := range names {
			group := &portaineree.EdgeGroup{
				Name:      name,
				Dynamic:   false,
				TagIDs:    make([]portaineree.TagID, 0),
				Endpoints: make([]portaineree.EndpointID, 0),
			}

			err := store.EdgeGroup().Create(group)
			if err != nil {
				return nil, err
			}

			groups[index] = *group
		}

		return groups, nil
	}

	checkGroups := func(store *datastore.Store, is *assert.Assertions, groupIDs []portaineree.EdgeGroupID, endpointID portaineree.EndpointID) {
		for _, groupID := range groupIDs {
			group, err := store.EdgeGroup().EdgeGroup(groupID)
			is.NoError(err)

			for _, endpoint := range group.Endpoints {
				if endpoint == endpointID {
					return
				}
			}
			is.Fail("expected endpoint to be in group")
		}
	}

	groupsByName := func(groups []portaineree.EdgeGroup, groupNames []string) []portaineree.EdgeGroup {
		result := make([]portaineree.EdgeGroup, len(groupNames))
		for i, tagName := range groupNames {
			for j, tag := range groups {
				if tag.Name == tagName {
					result[i] = groups[j]
					break
				}
			}
		}

		return result
	}

	type testCase struct {
		title              string
		endpoint           *portaineree.Endpoint
		groupNames         []string
		endpointGroupNames []string
		groupsToApply      []string
		shouldNotBeUpdated bool
	}

	testFn := func(t *testing.T, testCase testCase) {
		is := assert.New(t)
		_, store := datastore.MustNewTestStore(t, true, true)

		err := store.Endpoint().Create(testCase.endpoint)
		is.NoError(err)

		groups, err := createGroups(store, testCase.groupNames)
		is.NoError(err)

		endpointGroups := groupsByName(groups, testCase.endpointGroupNames)
		for _, group := range endpointGroups {
			group.Endpoints = append(group.Endpoints, testCase.endpoint.ID)

			err = store.EdgeGroup().UpdateEdgeGroup(group.ID, &group)
			is.NoError(err)
		}

		expectedGroups := groupsByName(groups, testCase.groupsToApply)
		expectedIDs := make([]portaineree.EdgeGroupID, len(expectedGroups))
		for i, tag := range expectedGroups {
			expectedIDs[i] = tag.ID
		}

		err = store.UpdateTx(func(tx dataservices.DataStoreTx) error {
			updated, err := updateEnvironmentEdgeGroups(tx, expectedIDs, testCase.endpoint.ID)
			is.NoError(err)

			is.Equal(testCase.shouldNotBeUpdated, !updated)

			return nil
		})

		is.NoError(err)

		checkGroups(store, is, expectedIDs, testCase.endpoint.ID)
	}

	testCases := []testCase{
		{
			title:              "applying edge groups to an endpoint without edge groups",
			endpoint:           &portaineree.Endpoint{},
			groupNames:         []string{"edge group1", "edge group2", "edge group3"},
			endpointGroupNames: []string{},
			groupsToApply:      []string{"edge group1", "edge group2", "edge group3"},
		},
		{
			title:              "applying edge groups to an endpoint with edge groups",
			endpoint:           &portaineree.Endpoint{},
			groupNames:         []string{"edge group1", "edge group2", "edge group3", "edge group4", "edge group5", "edge group6"},
			endpointGroupNames: []string{"edge group1", "edge group2", "edge group3"},
			groupsToApply:      []string{"edge group4", "edge group5", "edge group6"},
		},
		{
			title:              "applying edge groups to an endpoint with edge groups that are already applied",
			endpoint:           &portaineree.Endpoint{},
			groupNames:         []string{"edge group1", "edge group2", "edge group3"},
			endpointGroupNames: []string{"edge group1", "edge group2", "edge group3"},
			groupsToApply:      []string{"edge group1", "edge group2", "edge group3"},
			shouldNotBeUpdated: true,
		},
		{
			title:              "adding new edge groups to an endpoint with edge groups ",
			endpoint:           &portaineree.Endpoint{},
			groupNames:         []string{"edge group1", "edge group2", "edge group3", "edge group4", "edge group5", "edge group6"},
			endpointGroupNames: []string{"edge group1", "edge group2", "edge group3"},
			groupsToApply:      []string{"edge group1", "edge group2", "edge group3", "edge group4", "edge group5", "edge group6"},
		},
		{
			title:              "mixing edge groups that are already applied and new edge groups",
			endpoint:           &portaineree.Endpoint{},
			groupNames:         []string{"edge group1", "edge group2", "edge group3", "edge group4", "edge group5", "edge group6"},
			endpointGroupNames: []string{"edge group1", "edge group2", "edge group3"},
			groupsToApply:      []string{"edge group2", "edge group4", "edge group5"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			testFn(t, testCase)
		})
	}
}

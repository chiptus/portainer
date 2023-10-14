package staggers

import (
	"reflect"
	"testing"

	portainer "github.com/portainer/portainer/api"
)

func TestBuildStaggerQueueWithFixedDeviceNumber(t *testing.T) {
	arr := []portainer.EndpointID{1, 2, 3, 4, 5, 6, 7}

	tt := []struct {
		deviceNumber int
		expected     [][]portainer.EndpointID
	}{
		{
			deviceNumber: 2,
			expected: [][]portainer.EndpointID{
				{1, 2},
				{3, 4},
				{5, 6},
				{7},
			},
		},
		{
			deviceNumber: 3,
			expected: [][]portainer.EndpointID{
				{1, 2, 3},
				{4, 5, 6},
				{7},
			},
		},
		{
			deviceNumber: 1,
			expected: [][]portainer.EndpointID{
				{1},
				{2},
				{3},
				{4},
				{5},
				{6},
				{7},
			},
		},
	}

	for _, tc := range tt {
		queue := buildStaggerQueueWithFixedDeviceNumber(arr, tc.deviceNumber)
		if !reflect.DeepEqual(queue, tc.expected) {
			t.Errorf("Expected stagger queue %v, got %v", tc.expected, queue)
		}
	}
}

func TestBuildStaggerQueueWithIncrementalDeviceNumber(t *testing.T) {
	arr := []portainer.EndpointID{1, 2, 3, 4, 5, 6, 7, 8, 9}

	tt := []struct {
		deviceStartFrom   int
		deviceIncremental int
		expected          [][]portainer.EndpointID
	}{
		{
			deviceStartFrom:   1,
			deviceIncremental: 2,
			expected: [][]portainer.EndpointID{
				{1},
				{2, 3},
				{4, 5, 6, 7},
				{8, 9},
			},
		},
		{
			deviceStartFrom:   1,
			deviceIncremental: 3,
			expected: [][]portainer.EndpointID{
				{1},
				{2, 3, 4},
				{5, 6, 7, 8, 9},
			},
		},
		{
			deviceStartFrom:   2,
			deviceIncremental: 2,
			expected: [][]portainer.EndpointID{
				{1, 2},
				{3, 4, 5, 6},
				{7, 8, 9},
			},
		},
	}

	for _, tc := range tt {
		queue := buildStaggerQueueWithIncrementalDeviceNumber(arr, tc.deviceStartFrom, tc.deviceIncremental)
		if !reflect.DeepEqual(queue, tc.expected) {
			t.Errorf("Expected stagger queue %v, got %v", tc.expected, queue)
		}
	}
}

func TestCalculateNextStaggerCheckIntervalForAsyncUpdate(t *testing.T) {
	tt := []struct {
		edge     portainer.EnvironmentEdgeSettings
		expected int
	}{
		{
			edge: portainer.EnvironmentEdgeSettings{
				PingInterval:     10,
				CommandInterval:  20,
				SnapshotInterval: 30,
			},
			expected: 5,
		},
		{
			edge: portainer.EnvironmentEdgeSettings{
				PingInterval:     10,
				CommandInterval:  10,
				SnapshotInterval: 10,
			},
			expected: 5,
		},
		{
			edge: portainer.EnvironmentEdgeSettings{
				PingInterval:     -1,
				CommandInterval:  -1,
				SnapshotInterval: -1,
			},
			expected: 30,
		},
	}

	for _, tc := range tt {
		interval := calculateNextStaggerCheckIntervalForAsyncUpdate(&tc.edge)
		if interval != tc.expected {
			t.Errorf("Expected interval %d, got %d", tc.expected, interval)
		}
	}
}

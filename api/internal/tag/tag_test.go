package tag

import (
	"reflect"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
)

func TestIntersection(t *testing.T) {
	cases := []struct {
		name     string
		setA     tagSet
		setB     tagSet
		expected tagSet
	}{
		{
			name:     "positive numbers set intersection",
			setA:     Set([]portaineree.TagID{1, 2, 3, 4, 5}),
			setB:     Set([]portaineree.TagID{4, 5, 6, 7}),
			expected: Set([]portaineree.TagID{4, 5}),
		},
		{
			name:     "empty setA intersection",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{}),
			expected: Set([]portaineree.TagID{}),
		},
		{
			name:     "empty setB intersection",
			setA:     Set([]portaineree.TagID{}),
			setB:     Set([]portaineree.TagID{1, 2, 3}),
			expected: Set([]portaineree.TagID{}),
		},
		{
			name:     "no common elements sets intersection",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{4, 5, 6}),
			expected: Set([]portaineree.TagID{}),
		},
		{
			name:     "equal sets intersection",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{1, 2, 3}),
			expected: Set([]portaineree.TagID{1, 2, 3}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Intersection(tc.setA, tc.setB)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	cases := []struct {
		name     string
		setA     tagSet
		setB     tagSet
		expected tagSet
	}{
		{
			name:     "non-duplicate set union",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{4, 5, 6}),
			expected: Set([]portaineree.TagID{1, 2, 3, 4, 5, 6}),
		},
		{
			name:     "empty setA union",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{}),
			expected: Set([]portaineree.TagID{1, 2, 3}),
		},
		{
			name:     "empty setB union",
			setA:     Set([]portaineree.TagID{}),
			setB:     Set([]portaineree.TagID{1, 2, 3}),
			expected: Set([]portaineree.TagID{1, 2, 3}),
		},
		{
			name:     "duplicate elements in set union",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{3, 4, 5}),
			expected: Set([]portaineree.TagID{1, 2, 3, 4, 5}),
		},
		{
			name:     "equal sets union",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{1, 2, 3}),
			expected: Set([]portaineree.TagID{1, 2, 3}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Union(tc.setA, tc.setB)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	cases := []struct {
		name     string
		setA     tagSet
		setB     tagSet
		expected bool
	}{
		{
			name:     "setA contains setB",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{1, 2}),
			expected: true,
		},
		{
			name:     "setA equals to setB",
			setA:     Set([]portaineree.TagID{1, 2}),
			setB:     Set([]portaineree.TagID{1, 2}),
			expected: true,
		},
		{
			name:     "setA contains parts of setB",
			setA:     Set([]portaineree.TagID{1, 2}),
			setB:     Set([]portaineree.TagID{1, 2, 3}),
			expected: false,
		},
		{
			name:     "setA does not contain setB",
			setA:     Set([]portaineree.TagID{1, 2}),
			setB:     Set([]portaineree.TagID{3, 4}),
			expected: false,
		},
		{
			name:     "setA is empty and setB is not empty",
			setA:     Set([]portaineree.TagID{}),
			setB:     Set([]portaineree.TagID{1, 2}),
			expected: false,
		},
		{
			name:     "setA is not empty and setB is empty",
			setA:     Set([]portaineree.TagID{1, 2}),
			setB:     Set([]portaineree.TagID{}),
			expected: false,
		},
		{
			name:     "setA is empty and setB is empty",
			setA:     Set([]portaineree.TagID{}),
			setB:     Set([]portaineree.TagID{}),
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Contains(tc.setA, tc.setB)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestDifference(t *testing.T) {
	cases := []struct {
		name     string
		setA     tagSet
		setB     tagSet
		expected tagSet
	}{
		{
			name:     "positive numbers set difference",
			setA:     Set([]portaineree.TagID{1, 2, 3, 4, 5}),
			setB:     Set([]portaineree.TagID{4, 5, 6, 7}),
			expected: Set([]portaineree.TagID{1, 2, 3}),
		},
		{
			name:     "empty set difference",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{}),
			expected: Set([]portaineree.TagID{1, 2, 3}),
		},
		{
			name:     "equal sets difference",
			setA:     Set([]portaineree.TagID{1, 2, 3}),
			setB:     Set([]portaineree.TagID{1, 2, 3}),
			expected: Set([]portaineree.TagID{}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Difference(tc.setA, tc.setB)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

package tag

import portaineree "github.com/portainer/portainer-ee/api"

type tagSet map[portaineree.TagID]bool

// Set converts an array of ids to a set
func Set(tagIDs []portaineree.TagID) tagSet {
	set := map[portaineree.TagID]bool{}
	for _, tagID := range tagIDs {
		set[tagID] = true
	}
	return set
}

// Intersection returns a set intersection of the provided sets
func Intersection(sets ...tagSet) tagSet {
	intersection := tagSet{}
	if len(sets) == 0 {
		return intersection
	}
	setA := sets[0]
	for tag := range setA {
		inAll := true
		for _, setB := range sets {
			if !setB[tag] {
				inAll = false
				break
			}
		}

		if inAll {
			intersection[tag] = true
		}
	}

	return intersection
}

// Union returns a set union of provided sets
func Union(sets ...tagSet) tagSet {
	union := tagSet{}
	for _, set := range sets {
		for tag := range set {
			union[tag] = true
		}
	}
	return union
}

// Contains return true if setA contains setB
func Contains(setA tagSet, setB tagSet) bool {
	if len(setA) == 0 || len(setB) == 0 {
		return false
	}

	for tag := range setB {
		if !setA[tag] {
			return false
		}
	}
	return true
}

// Difference returns the set difference tagsA - tagsB
func Difference(setA tagSet, setB tagSet) tagSet {
	set := tagSet{}

	for tag := range setA {
		if !setB[tag] {
			set[tag] = true
		}
	}

	return set
}

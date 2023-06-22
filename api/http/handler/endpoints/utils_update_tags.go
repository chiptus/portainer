package endpoints

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/set"
)

// updateEnvironmentTags updates the tags associated to an environment
func updateEnvironmentTags(tx dataservices.DataStoreTx, newTags []portaineree.TagID, oldTags []portaineree.TagID, environmentID portaineree.EndpointID) (bool, error) {
	payloadTagSet := set.ToSet(newTags)
	environmentTagSet := set.ToSet(oldTags)
	union := set.Union(payloadTagSet, environmentTagSet)
	intersection := set.Intersection(payloadTagSet, environmentTagSet)

	if len(union) <= len(intersection) {
		return false, nil
	}

	updateSet := func(tagIDs set.Set[portaineree.TagID], updateItem func(*portaineree.Tag)) error {
		for tagID := range tagIDs {
			tag, err := tx.Tag().Read(tagID)
			if err != nil {
				return errors.WithMessage(err, "Unable to find a tag inside the database")
			}

			updateItem(tag)

			err = tx.Tag().Update(tagID, tag)
			if err != nil {
				return errors.WithMessage(err, "Unable to persist tag changes inside the database")
			}
		}

		return nil
	}

	removeTags := environmentTagSet.Difference(payloadTagSet)
	err := updateSet(removeTags, func(tag *portaineree.Tag) {
		delete(tag.Endpoints, environmentID)
	})
	if err != nil {
		return false, err
	}

	addTags := payloadTagSet.Difference(environmentTagSet)
	err = updateSet(addTags, func(tag *portaineree.Tag) {
		tag.Endpoints[environmentID] = true
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

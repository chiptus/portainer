package migrator

import portaineree "github.com/portainer/portainer-ee/api"

func (migrator *Migrator) assignEdgeGroupsToEdgeUpdatesForDB100() error {
	updates, err := migrator.edgeUpdateService.List()
	if err != nil {
		return err
	}

	for idx := range updates {
		update := updates[idx]
		edgeStack, err := migrator.edgeStackService.EdgeStack(update.EdgeStackID)
		if err != nil {
			return err
		}

		environmentIds := make([]portaineree.EndpointID, len(update.EnvironmentsPreviousVersions))
		i := 0
		for id := range update.EnvironmentsPreviousVersions {
			environmentIds[i] = id
			i++
		}

		edgeGroup := &portaineree.EdgeGroup{
			Name:         edgeStack.Name,
			Endpoints:    environmentIds,
			EdgeUpdateID: int(update.ID),
		}

		err = migrator.edgeGroupService.Create(edgeGroup)
		if err != nil {
			return err
		}

		update.EdgeGroupIDs = edgeStack.EdgeGroups
		err = migrator.edgeUpdateService.Update(update.ID, &update)
		if err != nil {
			return err
		}

		edgeStack.EdgeGroups = []portaineree.EdgeGroupID{edgeGroup.ID}
		err = migrator.edgeStackService.UpdateEdgeStack(edgeStack.ID, edgeStack)
		if err != nil {
			return err
		}
	}

	return nil
}

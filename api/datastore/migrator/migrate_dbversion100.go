package migrator

import (
	"fmt"
	"net"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/url"
	"github.com/rs/zerolog/log"
)

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

// rebuildEdgeStackFileSystemWithVersionForDB100 creates the edge stack version folder if needed.
// This is needed for backward compatibility with edge stacks created before the
// edge stack version folder was introduced.
func (migrator *Migrator) rebuildEdgeStackFileSystemWithVersionForDB100() error {
	edgeStacks, err := migrator.edgeStackService.EdgeStacks()
	if err != nil {
		return err
	}

	for _, edgeStack := range edgeStacks {
		if edgeStack.GitConfig != nil {
			// skip if the edge stack is deployed by git repository
			continue
		}

		edgeStackIdentifier := strconv.Itoa(int(edgeStack.ID))
		edgeStackVersionFolder := migrator.fileService.GetEdgeStackProjectPathByVersion(edgeStackIdentifier, edgeStack.Version)
		err = migrator.fileService.SafeMoveDirectory(edgeStack.ProjectPath, edgeStackVersionFolder)
		if err != nil {
			return fmt.Errorf("failed to copy edge stack %d project folder: %w", edgeStack.ID, err)
		}
	}
	return nil
}

func (migrator *Migrator) updateTunnelServerAddressForDB100() error {
	settings, err := migrator.settingsService.Settings()
	if err != nil {
		return err
	}

	if settings.EdgePortainerURL != "" && settings.Edge.TunnelServerAddress == "" {
		u, err := url.ParseURL(settings.EdgePortainerURL)
		if err != nil {
			return err
		}

		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}

		settings.Edge.TunnelServerAddress = net.JoinHostPort(host, *migrator.flags.TunnelPort)
		log.
			Info().
			Str("EdgePortainerURL", settings.EdgePortainerURL).
			Str("TunnelServerAddress", settings.Edge.TunnelServerAddress).
			Msg("TunnelServerAddress updated")
	}

	return migrator.settingsService.UpdateSettings(settings)
}

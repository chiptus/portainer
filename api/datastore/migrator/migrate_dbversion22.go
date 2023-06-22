package migrator

import (
	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/rs/zerolog/log"
)

func (m *Migrator) updateTagsToDBVersion23() error {
	log.Info().Msg("updating tags")

	tags, err := m.tagService.ReadAll()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		tag.EndpointGroups = make(map[portaineree.EndpointGroupID]bool)
		tag.Endpoints = make(map[portaineree.EndpointID]bool)

		err = m.tagService.Update(tag.ID, &tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateEndpointsAndEndpointGroupsToDBVersion23() error {
	log.Info().Msg("updating endpoints and endpoint groups")

	tags, err := m.tagService.ReadAll()
	if err != nil {
		return err
	}

	tagsNameMap := make(map[string]portaineree.Tag)
	for _, tag := range tags {
		tagsNameMap[tag.Name] = tag
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		endpointTags := make([]portaineree.TagID, 0)
		for _, tagName := range endpoint.Tags {
			tag, ok := tagsNameMap[tagName]
			if ok {
				endpointTags = append(endpointTags, tag.ID)
				tag.Endpoints[endpoint.ID] = true
			}
		}
		endpoint.TagIDs = endpointTags
		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}

		relation := &portaineree.EndpointRelation{
			EndpointID: endpoint.ID,
			EdgeStacks: map[portaineree.EdgeStackID]bool{},
		}

		err = m.endpointRelationService.Create(relation)
		if err != nil {
			return err
		}
	}

	endpointGroups, err := m.endpointGroupService.ReadAll()
	if err != nil {
		return err
	}

	for _, endpointGroup := range endpointGroups {
		endpointGroupTags := make([]portaineree.TagID, 0)
		for _, tagName := range endpointGroup.Tags {
			tag, ok := tagsNameMap[tagName]
			if ok {
				endpointGroupTags = append(endpointGroupTags, tag.ID)
				tag.EndpointGroups[endpointGroup.ID] = true
			}
		}
		endpointGroup.TagIDs = endpointGroupTags
		err = m.endpointGroupService.Update(endpointGroup.ID, &endpointGroup)
		if err != nil {
			return err
		}
	}

	for _, tag := range tagsNameMap {
		err = m.tagService.Update(tag.ID, &tag)
		if err != nil {
			return err
		}
	}

	return nil
}

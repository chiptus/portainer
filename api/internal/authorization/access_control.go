package authorization

import (
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
)

// NewAdministratorsOnlyResourceControl will create a new administrators only resource control associated to the resource specified by the
// identifier and type parameters.
func NewAdministratorsOnlyResourceControl(resourceIdentifier string, resourceType portaineree.ResourceControlType) *portaineree.ResourceControl {
	return &portaineree.ResourceControl{
		Type:               resourceType,
		ResourceID:         resourceIdentifier,
		SubResourceIDs:     []string{},
		UserAccesses:       []portaineree.UserResourceAccess{},
		TeamAccesses:       []portaineree.TeamResourceAccess{},
		AdministratorsOnly: true,
		Public:             false,
		System:             false,
	}
}

// NewPrivateResourceControl will create a new private resource control associated to the resource specified by the
// identifier and type parameters. It automatically assigns it to the user specified by the userID parameter.
func NewPrivateResourceControl(resourceIdentifier string, resourceType portaineree.ResourceControlType, userID portaineree.UserID) *portaineree.ResourceControl {
	return &portaineree.ResourceControl{
		Type:           resourceType,
		ResourceID:     resourceIdentifier,
		SubResourceIDs: []string{},
		UserAccesses: []portaineree.UserResourceAccess{
			{
				UserID:      userID,
				AccessLevel: portaineree.ReadWriteAccessLevel,
			},
		},
		TeamAccesses:       []portaineree.TeamResourceAccess{},
		AdministratorsOnly: false,
		Public:             false,
		System:             false,
	}
}

// NewSystemResourceControl will create a new public resource control with the System flag set to true.
// These kind of resource control are not persisted and are created on the fly by the Portainer API.
func NewSystemResourceControl(resourceIdentifier string, resourceType portaineree.ResourceControlType) *portaineree.ResourceControl {
	return &portaineree.ResourceControl{
		Type:               resourceType,
		ResourceID:         resourceIdentifier,
		SubResourceIDs:     []string{},
		UserAccesses:       []portaineree.UserResourceAccess{},
		TeamAccesses:       []portaineree.TeamResourceAccess{},
		AdministratorsOnly: false,
		Public:             true,
		System:             true,
	}
}

// NewPublicResourceControl will create a new public resource control.
func NewPublicResourceControl(resourceIdentifier string, resourceType portaineree.ResourceControlType) *portaineree.ResourceControl {
	return &portaineree.ResourceControl{
		Type:               resourceType,
		ResourceID:         resourceIdentifier,
		SubResourceIDs:     []string{},
		UserAccesses:       []portaineree.UserResourceAccess{},
		TeamAccesses:       []portaineree.TeamResourceAccess{},
		AdministratorsOnly: false,
		Public:             true,
		System:             false,
	}
}

// NewRestrictedResourceControl will create a new resource control with user and team accesses restrictions.
func NewRestrictedResourceControl(resourceIdentifier string, resourceType portaineree.ResourceControlType, userIDs []portaineree.UserID, teamIDs []portaineree.TeamID) *portaineree.ResourceControl {
	userAccesses := make([]portaineree.UserResourceAccess, 0)
	teamAccesses := make([]portaineree.TeamResourceAccess, 0)

	for _, id := range userIDs {
		access := portaineree.UserResourceAccess{
			UserID:      id,
			AccessLevel: portaineree.ReadWriteAccessLevel,
		}

		userAccesses = append(userAccesses, access)
	}

	for _, id := range teamIDs {
		access := portaineree.TeamResourceAccess{
			TeamID:      id,
			AccessLevel: portaineree.ReadWriteAccessLevel,
		}

		teamAccesses = append(teamAccesses, access)
	}

	return &portaineree.ResourceControl{
		Type:               resourceType,
		ResourceID:         resourceIdentifier,
		SubResourceIDs:     []string{},
		UserAccesses:       userAccesses,
		TeamAccesses:       teamAccesses,
		AdministratorsOnly: false,
		Public:             false,
		System:             false,
	}
}

// DecorateStacks will iterate through a list of stacks, check for an associated resource control for each
// stack and decorate the stack element if a resource control is found.
func DecorateStacks(stacks []portaineree.Stack, resourceControls []portaineree.ResourceControl) []portaineree.Stack {
	for idx, stack := range stacks {

		resourceControl := GetResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl, resourceControls)
		if resourceControl != nil {
			stacks[idx].ResourceControl = resourceControl
		}
	}

	return stacks
}

// DecorateCustomTemplates will iterate through a list of custom templates, check for an associated resource control for each
// template and decorate the template element if a resource control is found.
func DecorateCustomTemplates(templates []portaineree.CustomTemplate, resourceControls []portaineree.ResourceControl) []portaineree.CustomTemplate {
	for idx, template := range templates {

		resourceControl := GetResourceControlByResourceIDAndType(strconv.Itoa(int(template.ID)), portaineree.CustomTemplateResourceControl, resourceControls)
		if resourceControl != nil {
			templates[idx].ResourceControl = resourceControl
		}
	}

	return templates
}

// FilterAuthorizedStacks returns a list of decorated stacks filtered through resource control access checks.
func FilterAuthorizedStacks(stacks []portaineree.Stack, user *portaineree.User, userTeamIDs []portaineree.TeamID) []portaineree.Stack {
	authorizedStacks := make([]portaineree.Stack, 0)

	for _, stack := range stacks {
		_, isEndpointAdmin := user.EndpointAuthorizations[stack.EndpointID][portaineree.EndpointResourcesAccess]
		if isEndpointAdmin {
			authorizedStacks = append(authorizedStacks, stack)
			continue
		}

		if stack.ResourceControl != nil && UserCanAccessResource(user.ID, userTeamIDs, stack.ResourceControl) {
			authorizedStacks = append(authorizedStacks, stack)
		}
	}

	return authorizedStacks
}

// FilterAuthorizedCustomTemplates returns a list of decorated custom templates filtered through resource control access checks.
func FilterAuthorizedCustomTemplates(customTemplates []portaineree.CustomTemplate, user *portaineree.User, userTeamIDs []portaineree.TeamID) []portaineree.CustomTemplate {
	authorizedTemplates := make([]portaineree.CustomTemplate, 0)

	for _, customTemplate := range customTemplates {
		if customTemplate.CreatedByUserID == user.ID || (customTemplate.ResourceControl != nil && UserCanAccessResource(user.ID, userTeamIDs, customTemplate.ResourceControl)) {
			authorizedTemplates = append(authorizedTemplates, customTemplate)
		}
	}

	return authorizedTemplates
}

// UserCanAccessResource will valid that a user has permissions defined in the specified resource control
// based on its identifier and the team(s) he is part of.
func UserCanAccessResource(userID portaineree.UserID, userTeamIDs []portaineree.TeamID, resourceControl *portaineree.ResourceControl) bool {
	if resourceControl == nil {
		return false
	}

	for _, authorizedUserAccess := range resourceControl.UserAccesses {
		if userID == authorizedUserAccess.UserID {
			return true
		}
	}

	for _, authorizedTeamAccess := range resourceControl.TeamAccesses {
		for _, userTeamID := range userTeamIDs {
			if userTeamID == authorizedTeamAccess.TeamID {
				return true
			}
		}
	}

	return resourceControl.Public
}

// GetResourceControlByResourceIDAndType retrieves the first matching resource control in a set of resource controls
// based on the specified id and resource type parameters.
func GetResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType, resourceControls []portaineree.ResourceControl) *portaineree.ResourceControl {
	for _, resourceControl := range resourceControls {
		if resourceID == resourceControl.ResourceID && resourceType == resourceControl.Type {
			return &resourceControl
		}
		for _, subResourceID := range resourceControl.SubResourceIDs {
			if resourceID == subResourceID {
				return &resourceControl
			}
		}
	}
	return nil
}

package resourcecontrol

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker/consts"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
)

func FindResourceControl(resourceIdentifier string, resourceType portaineree.ResourceControlType, resourceLabelsObject map[string]interface{}, resourceControls []portaineree.ResourceControl, endpointId portaineree.EndpointID) (*portaineree.ResourceControl, error) {
	resourceControl := authorization.GetResourceControlByResourceIDAndType(resourceIdentifier, resourceType, resourceControls)
	if resourceControl != nil {
		return resourceControl, nil
	}

	if resourceLabelsObject != nil {
		if resourceLabelsObject[consts.SwarmServiceIdLabel] != nil {
			inheritedServiceIdentifier := resourceLabelsObject[consts.SwarmServiceIdLabel].(string)
			resourceControl = authorization.GetResourceControlByResourceIDAndType(inheritedServiceIdentifier, portaineree.ServiceResourceControl, resourceControls)

			if resourceControl != nil {
				return resourceControl, nil
			}
		}

		if resourceLabelsObject[consts.SwarmStackNameLabel] != nil {
			stackName := resourceLabelsObject[consts.SwarmStackNameLabel].(string)
			stackResourceID := stackutils.ResourceControlID(endpointId, stackName)
			resourceControl = authorization.GetResourceControlByResourceIDAndType(stackResourceID, portaineree.StackResourceControl, resourceControls)

			if resourceControl != nil {
				return resourceControl, nil
			}
		}

		if resourceLabelsObject[consts.ComposeStackNameLabel] != nil {
			stackName := resourceLabelsObject[consts.ComposeStackNameLabel].(string)
			stackResourceID := stackutils.ResourceControlID(endpointId, stackName)
			resourceControl = authorization.GetResourceControlByResourceIDAndType(stackResourceID, portaineree.StackResourceControl, resourceControls)

			if resourceControl != nil {
				return resourceControl, nil
			}
		}
	}

	return nil, nil
}

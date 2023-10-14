package docker

import (
	"net/http"
	"strings"

	"github.com/portainer/portainer-ee/api/docker/consts"
	portainer "github.com/portainer/portainer/api"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/utils"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"

	"github.com/rs/zerolog/log"
)

const (
	resourceLabelForPortainerTeamResourceControl   = "io.portainer.accesscontrol.teams"
	resourceLabelForPortainerUserResourceControl   = "io.portainer.accesscontrol.users"
	resourceLabelForPortainerPublicResourceControl = "io.portainer.accesscontrol.public"
)

type (
	resourceLabelsObjectSelector func(map[string]interface{}) map[string]interface{}
	resourceOperationParameters  struct {
		resourceIdentifierAttribute string
		resourceType                portainer.ResourceControlType
		labelsObjectSelector        resourceLabelsObjectSelector
	}
)

func getUniqueElements(items string) []string {
	result := []string{}
	seen := make(map[string]struct{})
	for _, item := range strings.Split(items, ",") {
		v := strings.TrimSpace(item)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; !ok {
			result = append(result, v)
			seen[v] = struct{}{}
		}
	}

	return result
}

func (transport *Transport) newResourceControlFromPortainerLabels(labelsObject map[string]interface{}, resourceID string, resourceType portainer.ResourceControlType) (*portainer.ResourceControl, error) {
	if labelsObject[resourceLabelForPortainerPublicResourceControl] != nil {
		resourceControl := authorization.NewPublicResourceControl(resourceID, resourceType)

		err := transport.dataStore.ResourceControl().Create(resourceControl)
		if err != nil {
			return nil, err
		}

		return resourceControl, nil
	}

	teamNames := make([]string, 0)
	userNames := make([]string, 0)
	if labelsObject[resourceLabelForPortainerTeamResourceControl] != nil {
		concatenatedTeamNames := labelsObject[resourceLabelForPortainerTeamResourceControl].(string)
		teamNames = getUniqueElements(concatenatedTeamNames)
	}

	if labelsObject[resourceLabelForPortainerUserResourceControl] != nil {
		concatenatedUserNames := labelsObject[resourceLabelForPortainerUserResourceControl].(string)
		userNames = getUniqueElements(concatenatedUserNames)
	}

	if len(teamNames) > 0 || len(userNames) > 0 {
		teamIDs := make([]portainer.TeamID, 0)
		userIDs := make([]portainer.UserID, 0)

		for _, name := range teamNames {
			team, err := transport.dataStore.Team().TeamByName(name)
			if err != nil {
				log.Warn().
					Str("name", name).
					Str("resource_id", resourceID).
					Msg("unknown team name in access control label, ignoring access control rule for this team")

				continue
			}

			teamIDs = append(teamIDs, team.ID)
		}

		for _, name := range userNames {
			user, err := transport.dataStore.User().UserByUsername(name)
			if err != nil {
				log.Warn().
					Str("name", name).
					Str("resource_id", resourceID).
					Msg("unknown user name in access control label, ignoring access control rule for this user")

				continue
			}

			userIDs = append(userIDs, user.ID)
		}

		resourceControl := authorization.NewRestrictedResourceControl(resourceID, resourceType, userIDs, teamIDs)

		err := transport.dataStore.ResourceControl().Create(resourceControl)
		if err != nil {
			return nil, err
		}

		return resourceControl, nil
	}

	return nil, nil
}

func (transport *Transport) createPrivateResourceControl(resourceIdentifier string, resourceType portainer.ResourceControlType, userID portainer.UserID) (*portainer.ResourceControl, error) {
	resourceControl := authorization.NewPrivateResourceControl(resourceIdentifier, resourceType, userID)

	err := transport.dataStore.ResourceControl().Create(resourceControl)
	if err != nil {
		log.Error().
			Str("resource", resourceIdentifier).
			Err(err).
			Msg("unable to persist resource control")

		return nil, err
	}

	return resourceControl, nil
}

func (transport *Transport) getInheritedResourceControlFromServiceOrStack(resourceIdentifier, nodeName string, resourceType portainer.ResourceControlType, resourceControls []portainer.ResourceControl) (*portainer.ResourceControl, error) {
	client, err := transport.dockerClientFactory.CreateClient(transport.endpoint, nodeName, nil)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	switch resourceType {
	case portaineree.ContainerResourceControl:
		return getInheritedResourceControlFromContainerLabels(client, transport.endpoint.ID, resourceIdentifier, resourceControls)
	case portaineree.NetworkResourceControl:
		return getInheritedResourceControlFromNetworkLabels(client, transport.endpoint.ID, resourceIdentifier, resourceControls)
	case portaineree.VolumeResourceControl:
		return getInheritedResourceControlFromVolumeLabels(client, transport.endpoint.ID, resourceIdentifier, resourceControls)
	case portaineree.ServiceResourceControl:
		return getInheritedResourceControlFromServiceLabels(client, transport.endpoint.ID, resourceIdentifier, resourceControls)
	case portaineree.ConfigResourceControl:
		return getInheritedResourceControlFromConfigLabels(client, transport.endpoint.ID, resourceIdentifier, resourceControls)
	case portaineree.SecretResourceControl:
		return getInheritedResourceControlFromSecretLabels(client, transport.endpoint.ID, resourceIdentifier, resourceControls)
	}

	return nil, nil
}

func (transport *Transport) applyAccessControlOnResource(parameters *resourceOperationParameters, responseObject map[string]interface{}, response *http.Response, executor *operationExecutor) error {
	if responseObject[parameters.resourceIdentifierAttribute] == nil {
		log.Warn().
			Str("identifier_attribute", parameters.resourceIdentifierAttribute).
			Msg("unable to find resource identifier property in resource object")

		return nil
	}

	if parameters.resourceType == portaineree.NetworkResourceControl {
		systemResourceControl := findSystemNetworkResourceControl(responseObject)
		if systemResourceControl != nil {
			responseObject = decorateObject(responseObject, systemResourceControl)
			return utils.RewriteResponse(response, responseObject, http.StatusOK)
		}
	}

	resourceIdentifier := responseObject[parameters.resourceIdentifierAttribute].(string)
	resourceLabelsObject := parameters.labelsObjectSelector(responseObject)

	resourceControl, err := transport.findResourceControl(resourceIdentifier, parameters.resourceType, resourceLabelsObject, executor.operationContext.resourceControls)
	if err != nil {
		return err
	}

	if resourceControl == nil && (executor.operationContext.isAdmin || executor.operationContext.endpointResourceAccess) {
		return utils.RewriteResponse(response, responseObject, http.StatusOK)
	}

	if executor.operationContext.isAdmin || executor.operationContext.endpointResourceAccess || (resourceControl != nil && authorization.UserCanAccessResource(executor.operationContext.userID, executor.operationContext.userTeamIDs, resourceControl)) {
		responseObject = decorateObject(responseObject, resourceControl)
		return utils.RewriteResponse(response, responseObject, http.StatusOK)
	}

	return utils.RewriteAccessDeniedResponse(response)
}

func (transport *Transport) applyAccessControlOnResourceList(parameters *resourceOperationParameters, resourceData []interface{}, executor *operationExecutor) ([]interface{}, error) {
	if executor.operationContext.isAdmin || executor.operationContext.endpointResourceAccess {
		return transport.decorateResourceList(parameters, resourceData, executor.operationContext.resourceControls)
	}

	return transport.filterResourceList(parameters, resourceData, executor.operationContext)
}

func (transport *Transport) decorateResourceList(parameters *resourceOperationParameters, resourceData []interface{}, resourceControls []portainer.ResourceControl) ([]interface{}, error) {
	decoratedResourceData := make([]interface{}, 0)

	for _, resource := range resourceData {
		resourceObject := resource.(map[string]interface{})

		if resourceObject[parameters.resourceIdentifierAttribute] == nil {
			log.Warn().
				Str("identifier_attribute", parameters.resourceIdentifierAttribute).
				Msg("unable to find resource identifier property in resource list element")

			continue
		}

		if parameters.resourceType == portaineree.NetworkResourceControl {
			systemResourceControl := findSystemNetworkResourceControl(resourceObject)
			if systemResourceControl != nil {
				resourceObject = decorateObject(resourceObject, systemResourceControl)
				decoratedResourceData = append(decoratedResourceData, resourceObject)
				continue
			}
		}

		resourceIdentifier := resourceObject[parameters.resourceIdentifierAttribute].(string)
		resourceLabelsObject := parameters.labelsObjectSelector(resourceObject)

		resourceControl, err := transport.findResourceControl(resourceIdentifier, parameters.resourceType, resourceLabelsObject, resourceControls)
		if err != nil {
			return nil, err
		}

		if resourceControl != nil {
			resourceObject = decorateObject(resourceObject, resourceControl)
		}

		decoratedResourceData = append(decoratedResourceData, resourceObject)
	}

	return decoratedResourceData, nil
}

func (transport *Transport) filterResourceList(parameters *resourceOperationParameters, resourceData []interface{}, context *restrictedDockerOperationContext) ([]interface{}, error) {
	filteredResourceData := make([]interface{}, 0)

	for _, resource := range resourceData {
		resourceObject := resource.(map[string]interface{})
		if resourceObject[parameters.resourceIdentifierAttribute] == nil {
			log.Warn().
				Str("identifier_attribute", parameters.resourceIdentifierAttribute).
				Msg("unable to find resource identifier property in resource list element")

			continue
		}

		resourceIdentifier := resourceObject[parameters.resourceIdentifierAttribute].(string)
		resourceLabelsObject := parameters.labelsObjectSelector(resourceObject)

		if parameters.resourceType == portaineree.NetworkResourceControl {
			systemResourceControl := findSystemNetworkResourceControl(resourceObject)
			if systemResourceControl != nil {
				resourceObject = decorateObject(resourceObject, systemResourceControl)
				filteredResourceData = append(filteredResourceData, resourceObject)
				continue
			}
		}

		resourceControl, err := transport.findResourceControl(resourceIdentifier, parameters.resourceType, resourceLabelsObject, context.resourceControls)
		if err != nil {
			return nil, err
		}

		if resourceControl == nil {
			if context.isAdmin || context.endpointResourceAccess {
				filteredResourceData = append(filteredResourceData, resourceObject)
			}
			continue
		}

		if context.isAdmin || context.endpointResourceAccess || authorization.UserCanAccessResource(context.userID, context.userTeamIDs, resourceControl) {
			resourceObject = decorateObject(resourceObject, resourceControl)
			filteredResourceData = append(filteredResourceData, resourceObject)
		}
	}

	return filteredResourceData, nil
}

func (transport *Transport) findResourceControl(resourceIdentifier string, resourceType portainer.ResourceControlType, resourceLabelsObject map[string]interface{}, resourceControls []portainer.ResourceControl) (*portainer.ResourceControl, error) {
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
			stackResourceID := stackutils.ResourceControlID(transport.endpoint.ID, stackName)
			resourceControl = authorization.GetResourceControlByResourceIDAndType(stackResourceID, portaineree.StackResourceControl, resourceControls)

			if resourceControl != nil {
				return resourceControl, nil
			}
		}

		if resourceLabelsObject[consts.ComposeStackNameLabel] != nil {
			stackName := resourceLabelsObject[consts.ComposeStackNameLabel].(string)
			stackResourceID := stackutils.ResourceControlID(transport.endpoint.ID, stackName)
			resourceControl = authorization.GetResourceControlByResourceIDAndType(stackResourceID, portaineree.StackResourceControl, resourceControls)

			if resourceControl != nil {
				return resourceControl, nil
			}
		}

		return transport.newResourceControlFromPortainerLabels(resourceLabelsObject, resourceIdentifier, resourceType)
	}

	return nil, nil
}

func getStackResourceIDFromLabels(resourceLabelsObject map[string]string, endpointID portainer.EndpointID) string {
	if resourceLabelsObject[consts.SwarmStackNameLabel] != "" {
		stackName := resourceLabelsObject[consts.SwarmStackNameLabel]
		return stackutils.ResourceControlID(endpointID, stackName)
	}

	if resourceLabelsObject[consts.ComposeStackNameLabel] != "" {
		stackName := resourceLabelsObject[consts.ComposeStackNameLabel]
		return stackutils.ResourceControlID(endpointID, stackName)
	}

	return ""
}

func decorateObject(object map[string]interface{}, resourceControl *portainer.ResourceControl) map[string]interface{} {
	if object["Portainer"] == nil {
		object["Portainer"] = make(map[string]interface{})
	}

	portainerMetadata := object["Portainer"].(map[string]interface{})
	portainerMetadata["ResourceControl"] = resourceControl
	return object
}

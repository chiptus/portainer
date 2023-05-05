package edgestacks

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

func hasKubeEndpoint(endpointService dataservices.EndpointService, endpointIDs []portaineree.EndpointID) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, endpointutils.IsKubernetesEndpoint)
}

func hasDockerEndpoint(endpointService dataservices.EndpointService, endpointIDs []portaineree.EndpointID) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, endpointutils.IsDockerEndpoint)
}

func hasEndpointPredicate(endpointService dataservices.EndpointService, endpointIDs []portaineree.EndpointID, predicate func(*portaineree.Endpoint) bool) (bool, error) {
	for _, endpointID := range endpointIDs {
		endpoint, err := endpointService.Endpoint(endpointID)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve environment from database: %w", err)
		}

		if predicate(endpoint) {
			return true, nil
		}
	}

	return false, nil
}

func hasWrongEnvironmentType(endpointService dataservices.EndpointService, endpointIDs []portaineree.EndpointID, deploymentType portaineree.EdgeStackDeploymentType) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, func(e *portaineree.Endpoint) bool {
		switch deploymentType {
		case portaineree.EdgeStackDeploymentKubernetes:
			return !endpointutils.IsKubernetesEndpoint(e)
		case portaineree.EdgeStackDeploymentCompose:
			return !endpointutils.IsDockerEndpoint(e)
		case portaineree.EdgeStackDeploymentNomad:
			return !endpointutils.IsNomadEndpoint(e)
		}

		return e.Type != portaineree.DockerEnvironment
	})
}

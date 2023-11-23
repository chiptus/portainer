package security

import (
	"net/http"
	"regexp"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

func portainerAgentOperationAuthorization(url, method string) portainer.Authorization {
	var dockerHubRule = regexp.MustCompile(`/docker/v2/dockerhub/\d+`)
	if dockerHubRule.MatchString(url) && method == http.MethodGet {
		return portaineree.OperationPortainerDockerHubInspect
	}
	return portaineree.OperationPortainerUndefined
}

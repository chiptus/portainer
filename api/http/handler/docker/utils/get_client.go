package utils

import (
	"net/http"

	dockerclient "github.com/docker/docker/client"
	portaineree "github.com/portainer/portainer-ee/api"
	prclient "github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
)

// GetClient returns a Docker client based on the request context
func GetClient(r *http.Request, dockerClientFactory *prclient.ClientFactory) (*dockerclient.Client, *httperror.HandlerError) {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return nil, httperror.NotFound("Unable to find an environment on request context", err)
	}

	agentTargetHeader := r.Header.Get(portaineree.PortainerAgentTargetHeader)

	cli, err := dockerClientFactory.CreateClient(endpoint, agentTargetHeader, nil)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to connect to the Docker daemon", err)
	}

	return cli, nil
}

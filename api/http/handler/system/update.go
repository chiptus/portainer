package system

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/platform"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
)

var platformToEndpointType = map[platform.ContainerPlatform]portainer.EndpointType{
	platform.PlatformDockerStandalone: portaineree.DockerEnvironment,
	platform.PlatformDockerSwarm:      portaineree.DockerEnvironment,
	platform.PlatformKubernetes:       portaineree.KubernetesLocalEnvironment,
}

// @id systemUpdate
// @summary Update Portainer to latest version
// @description Update Portainer to latest version
// @description **Access policy**: administrator
// @tags system
// @produce json
// @success 204 {object} status "Success"
// @router /system/update [post]
func (handler *Handler) systemUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	environment, err := handler.guessLocalEndpoint()
	if err != nil {
		return httperror.InternalServerError("Failed to guess local endpoint", err)
	}

	err = handler.updateService.Update(environment, "latest")
	if err != nil {
		return httperror.InternalServerError("Failed to update Portainer", err)
	}

	return response.Empty(w)
}

func (handler *Handler) guessLocalEndpoint() (*portaineree.Endpoint, error) {
	platform, err := platform.DetermineContainerPlatform()
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine container platform")
	}

	endpointType, ok := platformToEndpointType[platform]
	if !ok {
		return nil, errors.New("failed to determine endpoint type")
	}

	endpoints, err := handler.dataStore.Endpoint().Endpoints()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve endpoints")
	}

	for _, endpoint := range endpoints {
		if endpoint.Type == endpointType {
			return &endpoint, nil
		}
	}

	return nil, errors.New("failed to find local endpoint")
}

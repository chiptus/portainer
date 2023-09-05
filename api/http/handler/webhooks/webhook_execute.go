package webhooks

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	dockertypes "github.com/docker/docker/api/types"
)

// @summary Execute a webhook
// @description Acts on a passed in token UUID to restart the docker service
// @description **Access policy**: public
// @tags webhooks
// @param id path string true "Webhook token"
// @success 202 "Webhook executed"
// @failure 400
// @failure 500
// @router /webhooks/{id} [post]
func (handler *Handler) webhookExecute(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	webhookToken, err := request.RetrieveRouteVariableValue(r, "token")

	if err != nil {
		return httperror.InternalServerError("Invalid service id parameter", err)
	}

	webhook, err := handler.DataStore.Webhook().WebhookByToken(webhookToken)

	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a webhook with this token", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to retrieve webhook from the database", err)
	}

	resourceID := webhook.ResourceID
	endpointID := webhook.EndpointID
	registryID := webhook.RegistryID
	webhookType := webhook.WebhookType

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	imageTag, _ := request.RetrieveQueryParameter(r, "tag", true)

	agentTargetHeader := r.Header.Get(portaineree.PortainerAgentTargetHeader)

	switch webhookType {
	case portaineree.ServiceWebhook:
		return handler.executeServiceWebhook(w, endpoint, resourceID, registryID, imageTag)
	case portaineree.ContainerWebhook:
		return handler.executeContainerWebhook(w, endpoint, webhook, imageTag, agentTargetHeader)
	default:
		return httperror.InternalServerError("Unsupported webhook type", errors.New("Webhooks for this resource are not currently supported"))
	}
}

func (handler *Handler) executeServiceWebhook(
	w http.ResponseWriter,
	endpoint *portaineree.Endpoint,
	resourceID string,
	registryID portaineree.RegistryID,
	imageTag string,
) *httperror.HandlerError {
	dockerClient, err := handler.DockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return httperror.InternalServerError("Error creating docker client", err)
	}
	defer dockerClient.Close()

	service, _, err := dockerClient.ServiceInspectWithRaw(context.Background(), resourceID, dockertypes.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return httperror.InternalServerError("Error looking up service", err)
	}

	service.Spec.TaskTemplate.ForceUpdate++

	var imageName = strings.Split(service.Spec.TaskTemplate.ContainerSpec.Image, "@sha")[0]

	if imageTag != "" {
		var tagIndex = strings.LastIndex(imageName, ":")
		if tagIndex == -1 {
			tagIndex = len(imageName)
		}
		service.Spec.TaskTemplate.ContainerSpec.Image = imageName[:tagIndex] + ":" + imageTag
	} else {
		service.Spec.TaskTemplate.ContainerSpec.Image = imageName
	}

	serviceUpdateOptions := dockertypes.ServiceUpdateOptions{
		QueryRegistry: true,
	}

	if registryID != 0 {
		registry, err := handler.DataStore.Registry().Read(registryID)
		if err != nil {
			return httperror.InternalServerError("Error getting registry", err)
		}

		if registry.Authentication {
			registryutils.EnsureRegTokenValid(handler.DataStore, registry)
			serviceUpdateOptions.EncodedRegistryAuth, err = registryutils.GetRegistryAuthHeader(registry)
			if err != nil {
				return httperror.InternalServerError("Error getting registry auth header", err)
			}
		}
	}
	if imageTag != "" {
		rc, err := dockerClient.ImagePull(context.Background(), service.Spec.TaskTemplate.ContainerSpec.Image, dockertypes.ImagePullOptions{RegistryAuth: serviceUpdateOptions.EncodedRegistryAuth})
		if err != nil {
			return httperror.NotFound("Error pulling image with the specified tag", err)
		}
		defer func(rc io.ReadCloser) {
			_ = rc.Close()
		}(rc)
	}
	_, err = dockerClient.ServiceUpdate(context.Background(), resourceID, service.Version, service.Spec, serviceUpdateOptions)

	if err != nil {
		return httperror.InternalServerError("Error updating service", err)
	}
	return response.Empty(w)
}

func (handler *Handler) executeContainerWebhook(w http.ResponseWriter, endpoint *portaineree.Endpoint, webhook *portaineree.Webhook, imageTag, nodeName string) *httperror.HandlerError {
	newContainer, err := handler.containerService.Recreate(context.Background(), endpoint, webhook.ResourceID, true, imageTag, nodeName)
	if err != nil {
		return httperror.InternalServerError("Error updating service", err)
	}
	webhook.ResourceID = newContainer.ID
	err = handler.DataStore.Webhook().Update(webhook.ID, webhook)
	if err != nil {
		return httperror.InternalServerError("Error updating webhook", err)
	}
	return response.Empty(w)
}

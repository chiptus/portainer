package stacks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker/images"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	log "github.com/sirupsen/logrus"
)

var (
	_imageStatusCache *cache.Cache
)

func init() {
	_imageStatusCache = cache.New(5*time.Second, 5*time.Second)
}

func (handler *Handler) stackImagesStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if err != nil {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(stack.EndpointID))
	if err != nil {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	}

	if !endpointutils.IsDockerEndpoint(endpoint) {
		return httperror.InternalServerError("Not a docker endpoint", nil)
	}

	status, err := handler.stackImageDigest(r.Context(), stack, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable get the status of this stack", err)
	}

	return response.JSON(w, &images.StatusResponse{Status: status, Message: ""})
}

func (handler *Handler) stackImageDigest(ctx context.Context, stack *portaineree.Stack, endpoint *portaineree.Endpoint) (images.Status, error) {
	if stack.Status == portaineree.StackStatusInactive {
		return images.Skipped, nil
	}

	cacheKey := fmt.Sprintf("%d_%s", stack.EndpointID, stack.Name)
	status, err := getStatus(cacheKey)
	if err == nil {
		return status, nil
	}

	status, err = handler.imageStatus(ctx, stack, endpoint)
	if err != nil {
		_imageStatusCache.Set(cacheKey, images.Error, 0)
	}

	_imageStatusCache.Set(cacheKey, status, 0)

	return status, nil
}

func (handler *Handler) imageStatus(ctx context.Context, stack *portaineree.Stack, endpoint *portaineree.Endpoint) (images.Status, error) {
	if !endpointutils.IsDockerEndpoint(endpoint) {
		return images.Skipped, nil
	}
	cli, err := handler.DockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return images.Error, err
	}
	var imageList []string
	if stack.Type == portaineree.DockerComposeStack {
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			Filters: filters.NewArgs(filters.Arg("label", "com.docker.compose.project="+stack.Name)),
			All:     true,
		})
		if err != nil {
			return images.Error, err
		}
		for _, container := range containers {
			imageList = append(imageList, container.Image)
		}

	} else if stack.Type == portaineree.DockerSwarmStack {
		services, err := cli.ServiceList(ctx, types.ServiceListOptions{
			Filters: filters.NewArgs(filters.Arg("label", "com.docker.stack.namespace="+stack.Name)),
			Status:  false,
		})
		if err != nil {
			return images.Error, err
		}
		for _, service := range services {
			imageList = append(imageList, service.Spec.TaskTemplate.ContainerSpec.Image)
		}
	}

	if len(imageList) == 0 {
		log.Debugf("No containers or services under this stack: %s", stack.Name)
		return images.Skipped, nil
	}
	statusClient := images.NewClientWithRegistry(images.NewRegistryClient(handler.DataStore), cli)
	statuses := make([]images.Status, len(imageList))
	for i, resourceImage := range imageList {
		s, err := statusClient.Status(ctx, resourceImage)
		if err != nil {
			statuses[i] = images.Error
			log.Debugf("error when fetching image status for stacks: %v", err)
		}
		statuses[i] = s
		if s == images.Outdated || s == images.Processing {
			break
		}
	}
	if contains(statuses, images.Outdated) {
		return images.Outdated, nil
	} else if contains(statuses, images.Processing) {
		return images.Processing, nil
	} else if contains(statuses, images.Error) {
		return images.Error, nil
	} else if contains(statuses, images.Skipped) {
		return images.Skipped, nil
	} else {
		return images.Updated, nil
	}
}

func getStatus(stack string) (images.Status, error) {
	cacheData, ok := _imageStatusCache.Get(stack)
	if ok {
		status, ok := cacheData.(images.Status)
		if ok {
			return status, nil
		}
	}
	return "", errors.New("no digest found in cache")
}

func contains(statuses []images.Status, status images.Status) bool {
	if statuses == nil && len(statuses) == 0 {
		return false
	}

	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

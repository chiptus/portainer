package stacks

import (
	"context"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerdocker "github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/docker/consts"
	"github.com/portainer/portainer-ee/api/docker/images"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/rs/zerolog/log"
)

// @id stackImagesStatus
// @summary Fetch image status for stack
// @description
// @description **Access policy**:
// @tags docker
// @security jwt
// @accept json
// @produce json
// @param environmentId path int true "Environment identifier"
// @param id path int true "Stack identifier"
// @success 200 "Success"
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @router /docker/{environmentId}/stacks/{id}/images_status [get]
func (handler *Handler) stackImagesStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().Read(portaineree.StackID(stackID))
	if err != nil {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	}

	status, err := images.CachedResourceImageStatus(stack.Name)
	if err == nil {
		return response.JSON(w, &images.StatusResponse{Status: status, Message: ""})
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(stack.EndpointID))
	if err != nil {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	}

	if !endpointutils.IsDockerEndpoint(endpoint) {
		return httperror.InternalServerError("Not a docker endpoint", nil)
	}

	status, err = handler.stackImageStatus(r.Context(), stack, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to get the status of this stack", err)
	}

	if status != images.Preparing {
		images.CacheResourceImageStatus(stack.Name, status)
	}

	return response.JSON(w, &images.StatusResponse{Status: status, Message: ""})
}

func (handler *Handler) stackImageStatus(ctx context.Context, stack *portaineree.Stack, endpoint *portaineree.Endpoint) (images.Status, error) {
	if stack.Status == portaineree.StackStatusInactive {
		return images.Skipped, nil
	}

	status, err := handler.imageStatus(ctx, stack, endpoint)
	if err != nil {
		log.Error().Err(err).Msg("Unable to get image status")
		status = images.Error
	}

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

	statusClient := images.NewClientWithRegistry(images.NewRegistryClient(handler.DataStore), handler.DockerClientFactory)
	if stack.Type == portaineree.DockerComposeStack {
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("label", consts.ComposeStackNameLabel+"="+stack.Name)),
		})
		if err != nil {
			log.Warn().Err(err).Str("stackName", stack.Name).Msg("cannot list container for the compose stack")
			return images.Error, err
		}

		nonExistedOrStoppedContainers := make([]types.Container, 0)
		for _, container := range containers {
			if container.State == "exited" || container.State == "stopped" {
				continue
			}
			nonExistedOrStoppedContainers = append(nonExistedOrStoppedContainers, container)
		}
		if len(nonExistedOrStoppedContainers) == 0 {
			log.Debug().Str("stackName", stack.Name).Msg("No containers or services under this stack")
			return images.Preparing, nil
		}

		return statusClient.ContainersImageStatus(ctx, containers, endpoint), nil
	} else if stack.Type == portaineree.DockerSwarmStack {
		services, err := cli.ServiceList(ctx, types.ServiceListOptions{
			Filters: filters.NewArgs(filters.Arg("label", consts.SwarmStackNameLabel+"="+stack.Name)),
			Status:  false,
		})
		if err != nil {
			log.Warn().Str("stackName", stack.Name).Msg("cannot list services for the swarm stack")
			return images.Error, err
		}

		statuses := make([]images.Status, len(services))
		for i, service := range services {
			serviceID := service.ID
			s, err := statusClient.ServiceImageStatus(ctx, serviceID, endpoint)
			if err != nil {
				statuses[i] = images.Error
				log.Debug().Err(err).Msg("error when fetching image status for stack")
				continue
			}

			statuses[i] = s
			if s == images.Outdated || s == images.Processing {
				break
			}
		}
		return images.FigureOut(statuses), nil
	}

	return images.Skipped, nil
}

func EvictComposeStackImageStatusCache(ctx context.Context, endpoint *portaineree.Endpoint, stackName string, dockerClientFactory *portainerdocker.ClientFactory) {
	cli, err := dockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		log.Error().Err(err).Int("endpointId", int(endpoint.ID)).Msg("cannot create docker client for endpoint.")
		return
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", consts.ComposeStackNameLabel+"="+stackName)),
		All:     true,
	})
	if err != nil {
		log.Warn().Err(err).Str("stackName", stackName).Msg("cannot list container for the compose stack")
		return
	}

	if len(containers) == 0 {
		log.Debug().Str("stackName", stackName).Msg("No containers or services under this stack")
		return
	}
	for _, c := range containers {
		images.EvictImageStatus(c.ID)
	}
}

func EvictSwarmStackImageStatusCache(ctx context.Context, endpoint *portaineree.Endpoint, stackName string, dockerClientFactory *portainerdocker.ClientFactory) {
	cli, err := dockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		log.Error().Err(err).Int("endpointId", int(endpoint.ID)).Msg("cannot create docker client for endpoint")
		return
	}

	services, err := cli.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(filters.Arg("label", consts.SwarmStackNameLabel+"="+stackName)),
		Status:  false,
	})
	if err != nil {
		log.Warn().Err(err).Str("stackName", stackName).Msg("cannot list services for the swarm stack")
		return
	}

	if len(services) == 0 {
		log.Debug().Str("stackName", stackName).Msg("No services under this stack")
		return
	}
	for _, s := range services {
		images.EvictImageStatus(s.ID)
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			Filters: filters.NewArgs(filters.Arg("label", consts.SwarmServiceIdLabel+"="+s.ID)),
			All:     true,
		})
		if err != nil {
			log.Warn().Err(err).Str("serviceId", s.ID).Msg("cannot list container for the service")
			return
		}

		if len(containers) == 0 {
			log.Debug().Str("serviceId", s.ID).Msg("No containers or services under this service")
			return
		}
		for _, c := range containers {
			images.EvictImageStatus(c.ID)
		}
	}
}

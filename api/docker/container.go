package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	portaineree "github.com/portainer/portainer-ee/api"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Recreate a container, only can be trigger by a webhook
func (factory *ClientFactory) Recreate(ctx context.Context, endpoint *portaineree.Endpoint, containerId string, forcePullImage bool, imageTag string) (*types.ContainerJSON, error) {
	cli, err := factory.CreateClient(endpoint, "", nil)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	defer func(cli *client.Client) {
	_:
		cli.Close()
	}(cli)
	// 0. retrieve the container inspect
	container, _, err := cli.ContainerInspectWithRaw(ctx, containerId, true)
	if err != nil {
		return nil, err
	}

	image := strings.Split(container.Config.Image, "@sha")[0]
	if imageTag != "" {
		var tagIndex = strings.LastIndex(image, ":")
		if tagIndex == -1 {
			tagIndex = len(image)
		}
		container.Config.Image = image[:tagIndex] + ":" + imageTag
		image = container.Config.Image
	}
	// 1. pull image if you need force pull
	if imageTag != "" || forcePullImage {
		if err != nil {
			return nil, err
		}
		log.Debugf("Starting to pull the image: %s", image)
		_, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			return nil, err
		}
	}
	// 2. stop the current container
	log.Debugf("Starting to stop the container: %s", containerId)
	err = cli.ContainerStop(ctx, containerId, nil)
	if err != nil {
		return nil, err
	}
	// 3. rename the current container
	log.Debugf("Starting to rename the container: %s", containerId)
	err = cli.ContainerRename(ctx, containerId, container.Name+"-old")
	if err != nil {
		return nil, err
	}
	// 4. create a new container
	log.Debugf("Starting to create a new container with the same name: %s", strings.Split(container.Name, "/")[1])
	create, err := cli.ContainerCreate(ctx, container.Config, container.HostConfig,
		&network.NetworkingConfig{EndpointsConfig: container.NetworkSettings.Networks}, nil, container.Name)
	if err != nil {
		return nil, err
	}
	newContainerId := create.ID
	// 5. network connect with bridge(not sure)
	log.Debugf("Starting to connect network: %s", newContainerId)
	err = cli.NetworkConnect(ctx, container.HostConfig.NetworkMode.NetworkName(), newContainerId, nil)
	if err != nil {
		return nil, err
	}
	// 6. start the new container
	log.Debugf("Starting the new container: %s", newContainerId)
	err = cli.ContainerStart(ctx, newContainerId, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}
	// 7. delete the old container
	log.Debugf("Starting to remove the old container: %s", containerId)
	_ = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})

	newContainer, _, err := cli.ContainerInspectWithRaw(ctx, newContainerId, true)
	if err != nil {
		return nil, err
	}
	return &newContainer, nil
}

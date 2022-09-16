package docker

import (
	"context"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/images"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ContainerService struct {
	factory   *ClientFactory
	dataStore dataservices.DataStore
	sr        *serviceRestore
}

func NewContainerService(factory *ClientFactory, dataStore dataservices.DataStore) *ContainerService {
	return &ContainerService{
		factory:   factory,
		dataStore: dataStore,
		sr:        &serviceRestore{},
	}
}

// Recreate a container, only can be trigger by a webhook
func (c *ContainerService) Recreate(ctx context.Context, endpoint *portaineree.Endpoint, containerId string, forcePullImage bool, imageTag, nodeName string) (*types.ContainerJSON, error) {
	cli, err := c.factory.CreateClient(endpoint, nodeName, nil)
	if err != nil {
		return nil, errors.Wrap(err, "create client error")
	}

	defer func(cli *client.Client) {
	_:
		cli.Close()
	}(cli)

	log.Debug().Str("container_id", containerId).Msg("starting to fetch container information")

	container, _, err := cli.ContainerInspectWithRaw(ctx, containerId, true)
	if err != nil {
		return nil, errors.Wrap(err, "fetch container information error")
	}

	log.Debug().Str("image", container.Config.Image).Msg("starting to parse image")
	img, err := images.ParseImage(images.ParseImageOptions{
		Name: container.Config.Image,
	})
	if err != nil {
		return nil, errors.Wrap(err, "parse image error")
	}

	if imageTag != "" {
		err = img.WithTag(imageTag)
		if err != nil {
			return nil, errors.Wrap(err, "set image tag error")
		}

		log.Debug().Str("image", container.Config.Image).Msg("new image with tag")

		container.Config.Image = img.FullName()
	}

	// 1. pull image if you need force pull
	if forcePullImage {
		puller := images.NewPuller(cli, images.NewRegistryClient(c.dataStore), c.dataStore)
		err = puller.Pull(ctx, img)
		if err != nil {
			return nil, errors.Wrap(err, "pull image error")
		}
	}

	// 2. stop the current container
	log.Debug().Str("container_id", containerId).Msg("starting to stop the container")
	err = cli.ContainerStop(ctx, containerId, nil)
	if err != nil {
		return nil, errors.Wrap(err, "stop container error")
	}

	// 3. rename the current container
	log.Debug().Str("container_id", containerId).Msg("starting to rename the container")
	err = cli.ContainerRename(ctx, containerId, container.Name+"-old")
	if err != nil {
		return nil, errors.Wrap(err, "rename container error")
	}

	c.sr.enable()
	defer c.sr.close()
	defer c.sr.restore()

	c.sr.push(func() {
		log.Debug().Str("container_id", containerId).Str("container", container.Name).Msg("restoring the container")
		cli.ContainerRename(ctx, containerId, container.Name)
		cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
	})

	// 4. create a new container
	log.Debug().Str("container", strings.Split(container.Name, "/")[1]).Msg("starting to create a new container")
	create, err := cli.ContainerCreate(ctx, container.Config, container.HostConfig, &network.NetworkingConfig{EndpointsConfig: container.NetworkSettings.Networks}, nil, container.Name)

	c.sr.push(func() {
		log.Debug().Str("container_id", create.ID).Msg("removing the new container")
		cli.ContainerStop(ctx, create.ID, nil)
		cli.ContainerRemove(ctx, create.ID, types.ContainerRemoveOptions{})
	})

	if err != nil {
		return nil, errors.Wrap(err, "create container error")
	}

	newContainerId := create.ID

	// 5. network connect with bridge(not sure)
	log.Debug().Str("container_id", newContainerId).Msg("starting to connect network to container")
	err = cli.NetworkConnect(ctx, container.HostConfig.NetworkMode.NetworkName(), newContainerId, nil)
	if err != nil {
		return nil, errors.Wrap(err, "connect container network error")
	}

	// 6. start the new container
	log.Debug().Str("container_id", newContainerId).Msg("starting the new container")
	err = cli.ContainerStart(ctx, newContainerId, types.ContainerStartOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "start container error")
	}

	// 7. delete the old container
	log.Debug().Str("container_id", containerId).Msg("starting to remove the old container")
	_ = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})

	c.sr.disable()

	newContainer, _, err := cli.ContainerInspectWithRaw(ctx, newContainerId, true)
	if err != nil {
		return nil, errors.Wrap(err, "fetch container information error")
	}

	return &newContainer, nil
}

type serviceRestore struct {
	restoreC chan struct{}
	fs       []func()
}

func (sr *serviceRestore) enable() {
	sr.restoreC = make(chan struct{}, 1)
	sr.fs = make([]func(), 0)
	sr.restoreC <- struct{}{}
}

func (sr *serviceRestore) disable() {
	select {
	case <-sr.restoreC:
	default:
	}
}

func (sr *serviceRestore) push(f func()) {
	sr.fs = append(sr.fs, f)
}

func (sr *serviceRestore) restore() {
	select {
	case <-sr.restoreC:
		l := len(sr.fs)
		if l > 0 {
			for i := l - 1; i >= 0; i-- {
				sr.fs[i]()
			}
		}
	default:
	}
}

func (sr *serviceRestore) close() {
	if sr == nil || sr.restoreC == nil {
		return
	}

	select {
	case <-sr.restoreC:
	default:
	}

	close(sr.restoreC)
}

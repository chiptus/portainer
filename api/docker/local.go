package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func ScanAndCleanUpGhostUpdaterContainers(ctx context.Context) error {
	return withCli(func(cli *client.Client) error {
		foundRunningContainer := false
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("label", "io.portainer.updater=true")),
		})
		if err != nil {
			return fmt.Errorf("failed to list containers: %w", err)
		}
		for _, container := range containers {
			if container.State == "exited" {
				err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
				if err != nil {
					return fmt.Errorf("failed to remove container: %w", err)
				}

				if container.NetworkSettings != nil {
					for _, networkSetting := range container.NetworkSettings.Networks {
						err = cli.NetworkRemove(ctx, networkSetting.NetworkID)
						if err != nil {
							return fmt.Errorf("failed to remove network: %w", err)
						}
					}
				}
			} else if container.State == "running" {
				foundRunningContainer = true
			}
		}

		if foundRunningContainer {
			return errors.New("Found running updater container. Retry after 30 seconds.")
		}
		return nil
	})
}

func withCli(callback func(cli *client.Client) error) error {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	defer cli.Close()

	return callback(cli)
}

// Retry executes the given function f up to maxRetries times with a delay of delayBetweenRetries
func Retry(ctx context.Context, maxRetries int, delayBetweenRetries time.Duration, f func(ctx context.Context) error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = f(ctx)
		if err == nil {
			return nil
		}
		log.Warn().Err(err).Int("retry time", i).Msg("failed to clean up updater stack")
		time.Sleep(delayBetweenRetries)
	}
	return err
}

package update

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cbroglie/mustache"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/pkg/errors"
	dockerclient "github.com/portainer/portainer/api/docker/client"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/pkg/libstack"
	"github.com/rs/zerolog/log"
)

func (service *service) updateDocker(version, envType string) error {
	ctx := context.TODO()
	templateName := filesystem.JoinPaths(service.assetsPath, "mustache-templates", mustacheUpdateDockerTemplateFile)

	portainerImagePrefix := os.Getenv(portainerImagePrefixEnvVar)
	if portainerImagePrefix == "" {
		portainerImagePrefix = "portainer/portainer-ee"
	}

	image := fmt.Sprintf("%s:%s", portainerImagePrefix, version)

	skipPullImage := os.Getenv(skipPullImageEnvVar)

	err := service.checkImageForDocker(ctx, image, skipPullImage != "")
	if err != nil {
		return err
	}

	composeFile, err := mustache.RenderFile(templateName, map[string]string{
		"image":           image,
		"skip_pull_image": skipPullImage,
		"updater_image":   os.Getenv(updaterImageEnvVar),
		"envType":         envType,
	})

	log.Debug().
		Str("composeFile", composeFile).
		Msg("Compose file for update")

	if err != nil {
		return errors.Wrap(err, "failed to render update template")
	}

	tmpDir := os.TempDir()
	timeId := time.Now().Unix()
	filePath := filesystem.JoinPaths(tmpDir, fmt.Sprintf("update-%d.yml", timeId))

	r := bytes.NewReader([]byte(composeFile))

	err = filesystem.CreateFile(filePath, r)
	if err != nil {
		return errors.Wrap(err, "failed to create update compose file")
	}

	projectName := fmt.Sprintf(
		"portainer-update-%d-%s",
		timeId,
		strings.Replace(version, ".", "-", -1))

	err = service.composeDeployer.Deploy(
		ctx,
		[]string{filePath},
		libstack.DeployOptions{
			ForceRecreate:        true,
			AbortOnContainerExit: true,
			Options: libstack.Options{
				ProjectName: projectName,
			},
		},
	)

	// optimally, server was restarted by the updater, so we should not reach this point

	if err != nil {
		return errors.Wrap(err, "failed to deploy update stack")
	}

	return errors.New("update failed: server should have been restarted by the updater")
}

func (service *service) checkImageForDocker(ctx context.Context, image string, skipPullImage bool) error {
	cli, err := dockerclient.CreateClientFromEnv()
	if err != nil {
		return errors.Wrap(err, "failed to create docker client")
	}

	if skipPullImage {
		filters := filters.NewArgs()
		filters.Add("reference", image)
		images, err := cli.ImageList(ctx, types.ImageListOptions{
			Filters: filters,
		})

		if err != nil {
			return errors.Wrap(err, "failed to list images")
		}

		if len(images) == 0 {
			return errors.Errorf("image %s not found locally", image)
		}

		return nil
	}

	// check if available on registry
	_, err = cli.DistributionInspect(ctx, image, "")
	if err != nil {
		return errors.Errorf("image %s not found on registry", image)
	}

	return nil
}

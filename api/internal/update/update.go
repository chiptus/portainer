package update

import (
	"fmt"

	"github.com/pkg/errors"
	libstack "github.com/portainer/docker-compose-wrapper"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer/api/platform"
)

const (
	// mustacheUpgradeDockerTemplateFile represents the name of the template file for the docker upgrade
	mustacheUpdateDockerTemplateFile = "update-docker.yml.mustache"

	// portainerImagePrefixEnvVar represents the name of the environment variable used to define the image prefix for portainer-updater
	// useful if there's a need to test PR images
	portainerImagePrefixEnvVar = "UPGRADE_PORTAINER_IMAGE_PREFIX"
	// skipPullImageEnvVar represents the name of the environment variable used to define if the image pull should be skipped
	// useful if there's a need to test local images
	skipPullImageEnvVar = "UPGRADE_SKIP_PULL_PORTAINER_IMAGE"
	// updaterImageEnvVar represents the name of the environment variable used to define the updater image
	// useful if there's a need to test a different updater
	updaterImageEnvVar = "UPGRADE_UPDATER_IMAGE"
)

type Service interface {
	Update(environment *portaineree.Endpoint, version string) error
}

type service struct {
	composeDeployer         libstack.Deployer
	kubernetesClientFactory *cli.ClientFactory

	platform   platform.ContainerPlatform
	isUpdating bool

	assetsPath string
}

func NewService(assetsPath string, composeDeployer libstack.Deployer, kubernetesClientFactory *cli.ClientFactory) (*service, error) {
	platform, err := platform.DetermineContainerPlatform()
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine container platform")
	}

	return &service{
		composeDeployer: composeDeployer,
		platform:        platform,
		assetsPath:      assetsPath,
	}, nil
}

func (service *service) Update(environment *portaineree.Endpoint, version string) error {
	service.isUpdating = true

	switch service.platform {
	case platform.PlatformDockerStandalone:
		return service.updateDocker(version, "standalone")
	case platform.PlatformDockerSwarm:
		return service.updateDocker(version, "swarm")
	case platform.PlatformKubernetes:
		return service.updateKubernetes(environment, version)
	}

	return fmt.Errorf("unsupported platform %s", service.platform)
}

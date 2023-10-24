package updateschedules

import (
	"fmt"
	"os"
	"path"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/cbroglie/mustache"
	"github.com/pkg/errors"
)

const (
	// mustache template directory name
	mustacheTemplateDir = "mustache-templates"

	// mustacheUpdateEdgeStackTemplateFile represents the name of the edge stack template file for edge updates
	mustacheUpdateEdgeStackTemplateFile = "edge-update.yml.mustache"

	// mustacheUpdateNomadEdgeStackTemplateFile represents the name of the edge stack template file for Nomad edge updates
	mustacheUpdateNomadEdgeStackTemplateFile = "nomad-edge-update.hcl.mustache"

	// agentImagePrefixEnvVar represents the name of the environment variable used to define the agent image prefix for portainer-updater
	// useful if there's a need to test PR images
	agentImagePrefixEnvVar = "EDGE_UPDATE_AGENT_IMAGE_PREFIX"
	// skipPullAgentImageEnvVar represents the name of the environment variable used to define if the agent image pull should be skipped
	// useful if there's a need to test local images
	skipPullAgentImageEnvVar = "EDGE_UPDATE_SKIP_PULL_AGENT_IMAGE"
	// updaterImageEnvVar represents the name of the environment variable used to define the updater image
	// useful if there's a need to test a different updater
	updaterImageEnvVar = "EDGE_UPDATE_UPDATER_IMAGE"
)

func (service *Service) createEdgeStack(
	tx dataservices.DataStoreTx,
	scheduleID edgetypes.UpdateScheduleID,
	relatedEnvironments []portainer.EndpointID,
	registryID portainer.RegistryID,
	version, scheduledTime string,
	endpointType portainer.EndpointType,
) (portainer.EdgeStackID, error) {

	agentImagePrefix := os.Getenv(agentImagePrefixEnvVar)
	if agentImagePrefix == "" {
		agentImagePrefix = "portainer/agent"
	}

	prePullImage := false
	rePullImage := false
	registries := []portainer.RegistryID{}
	registryURL := ""
	if registryID != 0 {
		prePullImage = true
		rePullImage = true
		registries = append(registries, registryID)

		registry, err := tx.Registry().Read(registryID)
		if err != nil {
			return 0, errors.WithMessage(err, "failed to retrieve registry")
		}

		registryURL = registry.URL
		agentImagePrefix = fmt.Sprintf("%s/agent", registryURL)
	}

	agentImage := fmt.Sprintf("%s:%s", agentImagePrefix, version)

	deploymentConfig, err := getDeploymentConfig(endpointType, service.assetsPath)
	if err != nil {
		return 0, err
	}

	edgeGroup := &portaineree.EdgeGroup{
		Name:         buildEdgeStackName(scheduleID),
		Endpoints:    relatedEnvironments,
		EdgeUpdateID: int(scheduleID),
	}

	err = tx.EdgeGroup().Create(edgeGroup)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to create edge group for update schedule")
	}

	buildEdgeStackArgs := edgestacks.BuildEdgeStackArgs{
		Registries:            registries,
		ScheduledTime:         scheduledTime,
		UseManifestNamespaces: false,
		PrePullImage:          prePullImage,
		RePullImage:           rePullImage,
		RetryDeploy:           false,
	}

	stack, err := service.edgeStacksService.BuildEdgeStack(
		tx,
		buildEdgeStackName(scheduleID),
		deploymentConfig.Type,
		[]portainer.EdgeGroupID{edgeGroup.ID},
		buildEdgeStackArgs,
	)
	if err != nil {
		return 0, err
	}

	stack.EdgeUpdateID = int(scheduleID)

	_, err = service.edgeStacksService.PersistEdgeStack(
		tx,
		stack,
		func(stackFolder string, relatedEndpointIds []portainer.EndpointID) (configPath string, manifestPath string, projectPath string, err error) {
			skipPullAgentImage := ""
			env := os.Getenv(skipPullAgentImageEnvVar)
			if env != "" {
				skipPullAgentImage = "1"
			}

			updaterImage := os.Getenv(updaterImageEnvVar)
			if updaterImage == "" && registryURL != "" {
				updaterImage = fmt.Sprintf("%s/portainer-updater:latest", registryURL)
			}

			deploymentFile, err := mustache.RenderFile(deploymentConfig.TemplatePath, map[string]string{
				"agent_image_name":      agentImage,
				"schedule_id":           strconv.Itoa(int(scheduleID)),
				"skip_pull_agent_image": skipPullAgentImage,
				"updater_image":         updaterImage,
			})

			if err != nil {
				return "", "", "", errors.WithMessage(err, "failed to render edge stack template")
			}

			configPath = deploymentConfig.Path

			initialStackFileVersion := 1 // When creating a new stack, the version is always 1
			projectPath = service.fileService.GetEdgeStackProjectPath(stackFolder)

			_, err = service.fileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, configPath, initialStackFileVersion, []byte(deploymentFile))
			if err != nil {
				return "", "", "", err
			}

			return configPath, "", projectPath, nil
		})

	if err != nil {
		return 0, err
	}

	return stack.ID, nil
}

func buildEdgeStackName(scheduleId edgetypes.UpdateScheduleID) string {
	return fmt.Sprintf("edge-update-schedule-%d", scheduleId)
}

type DeploymentConfig struct {
	Type         portainer.EdgeStackDeploymentType
	TemplatePath string
	Path         string
}

// getDeploymentConfig returns a struct of deployment configuration based on the provided endpoint type
func getDeploymentConfig(endpointType portainer.EndpointType, assetsPath string) (DeploymentConfig, error) {
	config := DeploymentConfig{}

	switch endpointType {
	case portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.EdgeAgentOnDockerEnvironment:
		config.Type = portaineree.EdgeStackDeploymentCompose
		config.TemplatePath = path.Join(assetsPath, mustacheTemplateDir, mustacheUpdateEdgeStackTemplateFile)
		config.Path = filesystem.ComposeFileDefaultName

	case portaineree.EdgeAgentOnNomadEnvironment:
		config.Type = portaineree.EdgeStackDeploymentNomad
		config.TemplatePath = path.Join(assetsPath, mustacheTemplateDir, mustacheUpdateNomadEdgeStackTemplateFile)
		config.Path = eefs.NomadJobFileDefaultName

	default:
		return config, errors.New("endpoint type is not supported")
	}

	return config, nil
}

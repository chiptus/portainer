package edgeupdateschedules

import (
	"fmt"
	"os"
	"path"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
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

func (handler *Handler) createUpdateEdgeStack(
	scheduleID edgetypes.UpdateScheduleID,
	groupIDs []portaineree.EdgeGroupID,
	registryID portaineree.RegistryID,
	version, scheduledTime string,
	endpointType portaineree.EndpointType) (portaineree.EdgeStackID, error) {

	agentImagePrefix := os.Getenv(agentImagePrefixEnvVar)
	if agentImagePrefix == "" {
		agentImagePrefix = "portainer/agent"
	}

	prePullImage := false
	rePullImage := false
	registries := []portaineree.RegistryID{}
	var (
		registry *portaineree.Registry
		err      error
	)
	if registryID != 0 {
		prePullImage = true
		rePullImage = true
		registries = append(registries, registryID)

		registry, err = handler.dataStore.Registry().Registry(registryID)
		if err != nil {
			return 0, errors.WithMessage(err, "failed to retrieve registry")
		}

		agentImagePrefix = fmt.Sprintf("%s/agent", registry.URL)
	}

	agentImage := fmt.Sprintf("%s:%s", agentImagePrefix, version)

	deploymentConfig, err := getDeploymentConfig(endpointType, handler.assetsPath)
	if err != nil {
		return 0, err
	}

	stack, err := handler.edgeStacksService.BuildEdgeStack(
		handler.dataStore,
		buildEdgeStackName(scheduleID),
		deploymentConfig.Type,
		groupIDs,
		registries,
		scheduledTime,
		false,
		prePullImage,
		rePullImage,
		false,
	)
	if err != nil {
		return 0, err
	}

	stack.EdgeUpdateID = int(scheduleID)

	_, err = handler.edgeStacksService.PersistEdgeStack(handler.dataStore, stack, func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (configPath string, manifestPath string, projectPath string, err error) {
		skipPullAgentImage := ""
		env := os.Getenv(skipPullAgentImageEnvVar)
		if env != "" {
			skipPullAgentImage = "1"
		}

		updaterImage := os.Getenv(updaterImageEnvVar)
		if updaterImage == "" && registry != nil && registry.URL != "" {
			updaterImage = fmt.Sprintf("%s/portainer-updater:latest", registry.URL)
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

		projectPath, err = handler.fileService.StoreEdgeStackFileFromBytes(stackFolder, configPath, []byte(deploymentFile))
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
	Type         portaineree.EdgeStackDeploymentType
	TemplatePath string
	Path         string
}

// getDeploymentConfig returns a struct of deployment configuration based on the provided endpoint type
func getDeploymentConfig(endpointType portaineree.EndpointType, assetsPath string) (DeploymentConfig, error) {
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

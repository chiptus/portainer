package edgeupdateschedules

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/cbroglie/mustache"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"

	"github.com/portainer/portainer/api/filesystem"
)

const (
	// mustacheUpdateEdgeStackTemplateFile represents the name of the edge stack template file for edge updates
	mustacheUpdateEdgeStackTemplateFile = "edge-update.yml.mustache"

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

func (handler *Handler) createUpdateEdgeStack(scheduleID edgetypes.UpdateScheduleID, groupIDs []portaineree.EdgeGroupID, version string, scheduledTime string) (portaineree.EdgeStackID, error) {
	agentImagePrefix := os.Getenv(agentImagePrefixEnvVar)
	if agentImagePrefix == "" {
		agentImagePrefix = "portainer/agent"
	}

	agentImage := fmt.Sprintf("%s:%s", agentImagePrefix, version)

	stack, err := handler.edgeStacksService.BuildEdgeStack(buildEdgeStackName(scheduleID), portaineree.EdgeStackDeploymentCompose, groupIDs, []portaineree.RegistryID{}, scheduledTime)
	if err != nil {
		return 0, err
	}

	stack.EdgeUpdateID = int(scheduleID)

	_, err = handler.edgeStacksService.PersistEdgeStack(stack, func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (composePath string, manifestPath string, projectPath string, err error) {
		templateName := path.Join(handler.assetsPath, mustacheUpdateEdgeStackTemplateFile)
		skipPullAgentImage := ""
		env := os.Getenv(skipPullAgentImageEnvVar)
		if env != "" {
			skipPullAgentImage = "1"
		}

		composeFile, err := mustache.RenderFile(templateName, map[string]string{
			"agent_image_name":      agentImage,
			"schedule_id":           strconv.Itoa(int(scheduleID)),
			"skip_pull_agent_image": skipPullAgentImage,
			"updater_image":         os.Getenv(updaterImageEnvVar),
		})

		if err != nil {
			return "", "", "", errors.WithMessage(err, "failed to render edge stack template")
		}

		composePath = filesystem.ComposeFileDefaultName

		projectPath, err = handler.fileService.StoreEdgeStackFileFromBytes(stackFolder, composePath, []byte(composeFile))
		if err != nil {
			return "", "", "", err
		}

		return composePath, "", projectPath, nil
	})

	if err != nil {
		return 0, err
	}

	return stack.ID, nil
}

func buildEdgeStackName(scheduleId edgetypes.UpdateScheduleID) string {
	return fmt.Sprintf("edge-update-schedule-%d", scheduleId)
}

package edgestacks

import (
	"fmt"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/rs/zerolog/log"
)

func (handler *Handler) updateStackVersion(stack *portaineree.EdgeStack, deploymentType portaineree.EdgeStackDeploymentType, config []byte, oldGitHash string, relatedEnvironmentsIDs []portaineree.EndpointID) error {
	stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
		Version:    stack.Version,
		ConfigHash: oldGitHash,
	}
	stack.Version = stack.Version + 1
	stack.Status = newStatus(stack.Status, relatedEnvironmentsIDs)

	if stack.GitConfig != nil {
		return nil
	}

	return handler.storeStackFile(stack, deploymentType, config)
}

func (handler *Handler) storeStackFile(stack *portaineree.EdgeStack, deploymentType portaineree.EdgeStackDeploymentType, config []byte) error {

	if deploymentType != stack.DeploymentType {
		// deployment type was changed - need to delete all old files
		err := handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to clear old files")
		}

		stack.EntryPoint = ""
		stack.ManifestPath = ""
		stack.PreviousDeploymentInfo = nil
		stack.DeploymentType = deploymentType
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	entryPoint := ""
	if deploymentType == portaineree.EdgeStackDeploymentCompose {
		if stack.EntryPoint == "" {
			stack.EntryPoint = filesystem.ComposeFileDefaultName
		}

		entryPoint = stack.EntryPoint
	}

	if deploymentType == portaineree.EdgeStackDeploymentKubernetes {
		if stack.ManifestPath == "" {
			stack.ManifestPath = filesystem.ManifestFileDefaultName
		}

		entryPoint = stack.ManifestPath
	}

	if deploymentType == portaineree.EdgeStackDeploymentNomad {
		if stack.EntryPoint == "" {
			stack.EntryPoint = eefs.NomadJobFileDefaultName
		}

		entryPoint = stack.EntryPoint
	}

	_, err := handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, entryPoint, stack.Version, config)
	if err != nil {
		return fmt.Errorf("unable to persist updated Compose file with version on disk: %w", err)
	}

	return nil
}

func newStatus(oldStatus map[portaineree.EndpointID]portainer.EdgeStackStatus, relatedEnvironmentsIDs []portaineree.EndpointID) map[portaineree.EndpointID]portainer.EdgeStackStatus {
	status := map[portaineree.EndpointID]portainer.EdgeStackStatus{}

	for _, environmentID := range relatedEnvironmentsIDs {
		newEnvStatus := portainer.EdgeStackStatus{
			Details: portainer.EdgeStackStatusDetails{
				Pending: true,
			},
			DeploymentInfo: portainer.StackDeploymentInfo{},
		}

		oldEnvStatus, ok := oldStatus[environmentID]
		if ok {
			newEnvStatus.DeploymentInfo = oldEnvStatus.DeploymentInfo
		}

		status[environmentID] = newEnvStatus
	}

	return status
}

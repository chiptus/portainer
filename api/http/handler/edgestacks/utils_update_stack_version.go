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

func (handler *Handler) updateStackVersion(stack *portaineree.EdgeStack, deploymentType portaineree.EdgeStackDeploymentType, config []byte, oldGitHash string, relatedEnvironmentsIDs []portaineree.EndpointID, rollbackTo *int) error {
	stack.Version = stack.Version + 1
	stack.Status = newStatus(stack.Status, relatedEnvironmentsIDs)

	if stack.GitConfig != nil {
		if rollbackTo != nil {
			// rollback operation
			err := rollbackVersion(stack, rollbackTo)
			if err != nil {
				return err
			}

		} else {
			// update operation
			// Only update stack file version when git hash has changed
			if oldGitHash != stack.GitConfig.ConfigHash {
				stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
					FileVersion: stack.StackFileVersion,
					ConfigHash:  oldGitHash,
				}
				stack.StackFileVersion++
			}
		}

		return nil
	}

	return handler.storeStackFile(stack, deploymentType, config, rollbackTo)
}

func (handler *Handler) storeStackFile(stack *portaineree.EdgeStack, deploymentType portaineree.EdgeStackDeploymentType, config []byte, rollbackTo *int) error {
	if deploymentType != stack.DeploymentType {
		if rollbackTo != nil {
			return fmt.Errorf("unable to rollback to a different deployment type")
		}

		// deployment type was changed - need to delete all old files
		err := handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to clear old files")
		}

		stack.EntryPoint = ""
		stack.ManifestPath = ""
		stack.PreviousDeploymentInfo = nil
		stack.DeploymentType = deploymentType

		// increment stack file version and clean up previous deployment info
		stack.PreviousDeploymentInfo = nil
		stack.StackFileVersion++
	} else {
		if rollbackTo != nil {
			// rollback operation
			err := rollbackVersion(stack, rollbackTo)
			if err != nil {
				return err
			}

			hasChanged, err := handler.hasFileContentChanged(stack, string(config))
			if err != nil {
				return fmt.Errorf("unable to check if file content matches with stack file v%d: %w", stack.StackFileVersion, err)
			}

			if hasChanged {
				return fmt.Errorf("file content in payload doesn't match stack file version v%d: %w", stack.StackFileVersion, err)
			}

		} else {
			// update operation
			hasChanged, err := handler.hasFileContentChanged(stack, string(config))
			if err != nil {
				return fmt.Errorf("unable to check if stack file content has changed: %w", err)
			}

			if !hasChanged {
				return nil
			}

			stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
				FileVersion: stack.StackFileVersion,
			}
			stack.StackFileVersion++
		}
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

	_, err := handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, entryPoint, stack.StackFileVersion, config)
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

// hasFileContentChanged checks if the file content has changed for non-git deployed stacks
func (handler *Handler) hasFileContentChanged(stack *portaineree.EdgeStack, updatedFileContent string) (bool, error) {
	if stack.GitConfig != nil {
		return false, fmt.Errorf("hasFileContentChanged should not be called for git deployed stacks")
	}

	if len(updatedFileContent) == 0 {
		return false, fmt.Errorf("updatedFileContent is empty")
	}

	projectVersionPath := handler.FileService.FormProjectPathByVersion(stack.ProjectPath, stack.StackFileVersion, "")

	currentFileContent, err := handler.FileService.GetFileContent(projectVersionPath, stack.EntryPoint)
	if err != nil {
		return false, err
	}

	if string(currentFileContent) != updatedFileContent {
		return true, nil
	}
	return false, nil
}

func rollbackVersion(stack *portaineree.EdgeStack, rollbackTo *int) error {
	if rollbackTo == nil {
		return fmt.Errorf("rollback version is not set")
	}

	if stack.PreviousDeploymentInfo == nil {
		return fmt.Errorf("no previous deployment info")
	}

	// If we support multiple rollback versions, we can remove this check or
	// change it to check if the rollback version is in the list of previous versions
	if *rollbackTo != stack.PreviousDeploymentInfo.FileVersion {
		return fmt.Errorf("rollback version is not the same as previous deployment version")
	}

	stack.StackFileVersion = *rollbackTo
	if stack.GitConfig != nil {
		// rollback for git edge stack
		stack.GitConfig.ConfigHash = stack.PreviousDeploymentInfo.ConfigHash
		// for git deployed stack rollback, we can't find the previous commit hash of rollback version
		// so we must remove the previous deployment info after rollback
		stack.PreviousDeploymentInfo = nil
	} else {
		// rollback for non-git edge stack
		if stack.PreviousDeploymentInfo.FileVersion == 1 {
			stack.PreviousDeploymentInfo = nil
		} else {
			stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
				FileVersion: stack.PreviousDeploymentInfo.FileVersion - 1,
			}
		}
	}

	return nil
}

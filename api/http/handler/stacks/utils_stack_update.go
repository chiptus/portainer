package stacks

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

func (handler *Handler) hasFileContentChanged(stack *portaineree.Stack, updatedFileContent string) (bool, error) {
	commitHash := ""
	if stack.GitConfig != nil {
		commitHash = stack.GitConfig.ConfigHash
	}

	projectVersionPath := handler.FileService.FormProjectPathByVersion(stack.ProjectPath, stack.StackFileVersion, commitHash)

	currentFileContent, err := handler.FileService.GetFileContent(projectVersionPath, stack.EntryPoint)
	if err != nil {
		return false, err
	}

	if string(currentFileContent) != updatedFileContent {
		return true, nil
	}
	return false, nil
}

func rollbackVersion(stack *portaineree.Stack, rollbackTo *int) error {
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

		return nil
	}

	// rollback for non-git edge stack
	if stack.PreviousDeploymentInfo.FileVersion == 1 {
		stack.PreviousDeploymentInfo = nil
	} else {
		stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
			FileVersion: stack.PreviousDeploymentInfo.FileVersion - 1,
		}
	}

	return nil
}

func (handler *Handler) updateStackFileVersion(stack *portaineree.Stack, configFile string, rollbackTo *int) error {
	if rollbackTo != nil {
		// rollback opertaion
		err := rollbackVersion(stack, rollbackTo)
		if err != nil {
			return fmt.Errorf("unable to rollback stack: %w", err)
		}

		// Check again if the payload file content matches with the file content of previous version
		// Do not move this check out of the rollback block, because the stack file version has updated
		hasChanged, err := handler.hasFileContentChanged(stack, configFile)
		if err != nil {
			return fmt.Errorf("unable to check if file content changed: %w", err)
		}

		if hasChanged {
			return fmt.Errorf("file content in payload doesn't match stack file version rolling back: %w", err)
		}

		return nil
	}

	// update operation
	hasContentChanged, err := handler.hasFileContentChanged(stack, configFile)
	if err != nil {
		return fmt.Errorf("unable to check if file content changed: %w", err)
	}

	if hasContentChanged {
		stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
			FileVersion: stack.StackFileVersion,
		}

		stack.StackFileVersion++
	}

	return nil
}

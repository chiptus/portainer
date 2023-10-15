package migrator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/rs/zerolog/log"
)

// rebuildEdgeStackFileSystemWithVersion creates the edge stack version folder if needed.
// This is needed for backward compatibility with edge stacks created before the
// edge stack version folder was introduced.
func (migrator *Migrator) rebuildEdgeStackFileSystemWithVersion() error {
	edgeStacks, err := migrator.edgeStackService.EdgeStacks()
	if err != nil {
		return err
	}

	for _, edgeStack := range edgeStacks {
		if edgeStack.StackFileVersion > 0 {
			err := RemoveGitStackVersionFolders(edgeStack.ProjectPath, edgeStack.GitConfig, edgeStack.PreviousDeploymentInfo, migrator.fileService.FormProjectPathByVersion)
			if err != nil {
				log.Info().Err(err).Msg("failed to remove old edge stack version folders")
			}

			continue
		}

		commitHash := ""
		if edgeStack.GitConfig != nil {
			commitHash = edgeStack.GitConfig.ConfigHash
		}

		edgeStackIdentifier := strconv.Itoa(int(edgeStack.ID))

		edgeStack.StackFileVersion = edgeStack.Version
		edgeStackVersionFolder := migrator.fileService.GetEdgeStackProjectPathByVersion(edgeStackIdentifier, edgeStack.StackFileVersion, commitHash)

		// Conduct the source folder checks to avoid unnecessary error return
		// In the normal case, the source folder should exist, However, there is a chance that
		// the edge stack folder was deleted by the user, but the edge stack id is still in the
		// database. In this case, we should skip folder migration
		sourceExists, err := migrator.fileService.FileExists(edgeStack.ProjectPath)
		if err != nil {
			log.Warn().
				Err(err).
				Int("edgeStackID", int(edgeStack.ID)).
				Msg("failed to check if edge stack project folder exists")
			continue
		}
		if !sourceExists {
			log.Debug().
				Int("edgeStackID", int(edgeStack.ID)).
				Msg("edge stack project folder does not exist, skipping")
			continue
		}

		/*
			We do not need to check if the target folder exists or not, because
			1. There is a chance the edge stack folder already included a version folder that matches
			with our version folder name. But it was added by user or existed in git repository originally.
			In that case, we should still add our version folder as the parent folder. For example:

			Original:                                       After migration:

			└── edge-stacks                                     └── edge-stacks
				└── 1                                               └── 1
					├── docker-compose.yml                              └── v1
					└── v1                                                  ├── docker-compose.yml
																			└── v1
			 2. As the migration function will be only invoked once when the database is upgraded
			 from lower version to 100, we do not need to worry about nested subfolders being created
			 multiple times. For example: /edge-stacks/2/v1/v1/v1/v1/docker-compose.yml
		*/

		err = migrator.fileService.SafeMoveDirectory(edgeStack.ProjectPath, edgeStackVersionFolder)
		if err != nil {
			return fmt.Errorf("failed to copy edge stack %d project folder: %w", edgeStack.ID, err)
		}

		err = migrator.edgeStackService.UpdateEdgeStackFunc(edgeStack.ID, func(edgeStack *portaineree.EdgeStack) {
			edgeStack.StackFileVersion = edgeStack.Version
		})
		if err != nil {
			return fmt.Errorf("failed to update edge stack %d file version: %w", edgeStack.ID, err)
		}
	}
	return nil
}

// rebuildStackFileSystemWithVersion creates the regular stack version folder if needed.
// This is needed for backward compatibility with regular stacks created before the
// regular stack version folder was introduced.
func (migrator *Migrator) rebuildStackFileSystemWithVersion() error {
	stacks, err := migrator.stackService.ReadAll()
	if err != nil {
		return err
	}

	for _, stack := range stacks {
		if stack.StackFileVersion > 0 {
			// we only keep the latest version folder for stack
			err := RemoveGitStackVersionFolders(stack.ProjectPath, stack.GitConfig, nil, migrator.fileService.FormProjectPathByVersion)
			if err != nil {
				log.Info().Err(err).Msg("failed to remove old stack version folders")
			}

			continue
		}

		commitHash := ""
		if stack.GitConfig != nil {
			commitHash = stack.GitConfig.ConfigHash
		}

		stackIdentifier := strconv.Itoa(int(stack.ID))

		stack.StackFileVersion = 1
		stackVersionFolder := migrator.fileService.GetStackProjectPathByVersion(stackIdentifier, stack.StackFileVersion, commitHash)

		// Conduct the source folder checks to avoid unnecessary error return, same
		// as the above edge stack migration.
		sourceExists, err := migrator.fileService.FileExists(stack.ProjectPath)
		if err != nil {
			log.Warn().
				Err(err).
				Int("stackID", int(stack.ID)).
				Msg("failed to check if stack project folder exists")
			continue
		}
		if !sourceExists {
			log.Debug().
				Int("stackID", int(stack.ID)).
				Msg("stack project folder does not exist, skipping")
			continue
		}

		err = migrator.fileService.SafeMoveDirectory(stack.ProjectPath, stackVersionFolder)
		if err != nil {
			return fmt.Errorf("failed to copy stack %d project folder: %w", stack.ID, err)
		}

		err = migrator.stackService.Update(stack.ID, &stack)
		if err != nil {
			return fmt.Errorf("failed to update stack %d file version: %w", stack.ID, err)
		}

	}
	return nil
}

func RemoveGitStackVersionFolders(projectPath string, gitConfig *gittypes.RepoConfig, prevInfo *portainer.StackDeploymentInfo, formPath func(string, int, string) string) error {
	// If the stack version folder is already migrated, we can remove
	// stack version folders that are no longer referenced.
	if gitConfig != nil {
		keptPaths := []string{
			formPath(projectPath, 0, gitConfig.ConfigHash),
		}

		if prevInfo != nil {
			keptPaths = append(keptPaths, formPath(projectPath, 0, prevInfo.ConfigHash))
		}

		err := removeUnreferencedGitStackVersionFolders(projectPath, keptPaths)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeUnreferencedGitStackVersionFolders(root string, keptPaths []string) error {
	foldersToBeRemoved := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		if path == root {
			return nil
		}

		for _, keptPath := range keptPaths {
			if path == keptPath {
				return filepath.SkipDir
			}
		}

		foldersToBeRemoved = append(foldersToBeRemoved, path)
		return nil
	})
	if err != nil {
		return err
	}

	for _, path := range foldersToBeRemoved {
		err = os.RemoveAll(path)
		if err != nil {
			log.Info().Err(err).Msg("failed to remove old stack version folders that are no longer referenced")
		}
	}
	return nil
}

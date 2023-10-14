package stackbuilders

import (
	"strconv"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
)

type GitMethodStackBuildProcess interface {
	// Set general stack information
	SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess
	// Set unique stack information, e.g. swarm stack has swarmID, kubernetes stack has namespace
	SetUniqueInfo(payload *StackPayload) GitMethodStackBuildProcess
	// Deploy stack based on the configuration
	Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess
	// Save the stack information to database
	SaveStack() (*portaineree.Stack, *httperror.HandlerError)
	// Get reponse from http request. Use if it is needed
	GetResponse() string
	// Set git repository configuration
	SetGitRepository(payload *StackPayload, userID portainer.UserID) GitMethodStackBuildProcess
	// Set auto update setting
	SetAutoUpdate(payload *StackPayload) GitMethodStackBuildProcess
}

type GitMethodStackBuilder struct {
	StackBuilder
	userActivityService portaineree.UserActivityService
	gitService          portainer.GitService
	scheduler           *scheduler.Scheduler
}

func (b *GitMethodStackBuilder) SetGeneralInfo(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess {
	stackID := b.dataStore.Stack().GetNextIdentifier()
	b.stack.ID = portainer.StackID(stackID)
	b.stack.EndpointID = endpoint.ID
	b.stack.AdditionalFiles = payload.AdditionalFiles
	b.stack.Status = portaineree.StackStatusActive
	b.stack.CreationDate = time.Now().Unix()
	b.stack.AutoUpdate = payload.AutoUpdate
	b.stack.SupportRelativePath = payload.SupportRelativePath
	b.stack.FilesystemPath = payload.FilesystemPath
	return b
}

func (b *GitMethodStackBuilder) SetUniqueInfo(payload *StackPayload) GitMethodStackBuildProcess {

	return b
}

func (b *GitMethodStackBuilder) SetGitRepository(payload *StackPayload, userID portainer.UserID) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	var repoConfig gittypes.RepoConfig
	gitAuthentication, err := b.extractGitAuthenticationFromPayload(&payload.RepositoryConfigPayload, userID)
	if err != nil {
		return b
	}

	repoConfig.Authentication = gitAuthentication
	repoConfig.URL = payload.URL
	repoConfig.ReferenceName = payload.ReferenceName
	repoConfig.TLSSkipVerify = payload.TLSSkipVerify

	repoConfig.ConfigFilePath = payload.ComposeFile
	if payload.ComposeFile == "" {
		repoConfig.ConfigFilePath = filesystem.ComposeFileDefaultName
	}
	// If a manifest file is specified (for kube git apps), then use it instead of the default compose file name
	if payload.ManifestFile != "" {
		repoConfig.ConfigFilePath = payload.ManifestFile
	}

	// Set the project path on the disk
	stackFolder := strconv.Itoa(int(b.stack.ID))
	getProjectPath := func(enableVersionFolder bool, commitHash string) string {
		if enableVersionFolder {
			return b.fileService.GetStackProjectPathByVersion(stackFolder, 0, commitHash)
		}
		return b.fileService.GetStackProjectPath(stackFolder)
	}

	commitHash, err := stackutils.DownloadGitRepository(repoConfig, b.gitService, true, getProjectPath)
	if err != nil {
		b.err = httperror.InternalServerError(err.Error(), err)
		return b
	}

	// This projectPath should be non-versioned project path. e.g. /data/compose/1
	b.stack.ProjectPath = b.fileService.GetStackProjectPath(stackFolder)

	// Update the latest commit id
	repoConfig.ConfigHash = commitHash
	b.stack.GitConfig = &repoConfig
	return b
}

func (b *GitMethodStackBuilder) Deploy(payload *StackPayload, endpoint *portaineree.Endpoint) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	// Deploy the stack
	err := b.deploymentConfiger.Deploy()
	if err != nil {
		b.err = httperror.InternalServerError(err.Error(), err)
		return b
	}

	return b
}

func (b *GitMethodStackBuilder) SetAutoUpdate(payload *StackPayload) GitMethodStackBuildProcess {
	if b.hasError() {
		return b
	}

	if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
		jobID, err := deployments.StartAutoupdate(b.stack.ID,
			b.stack.AutoUpdate.Interval,
			b.scheduler,
			b.stackDeployer,
			b.dataStore,
			b.gitService,
			b.userActivityService)
		if err != nil {
			b.err = err
			return b
		}

		b.stack.AutoUpdate.JobID = jobID
	}
	return b
}

func (b *GitMethodStackBuilder) GetResponse() string {
	return ""
}

func (b *GitMethodStackBuilder) extractGitAuthenticationFromPayload(payload *RepositoryConfigPayload, requestUserID portainer.UserID) (*gittypes.GitAuthentication, error) {
	if payload.Authentication {
		repositoryUsername := ""
		repositoryPassword := ""
		repositoryGitCredentialID := 0
		if payload.GitCredentialID != 0 {
			credential, err := b.dataStore.GitCredential().Read(portaineree.GitCredentialID(payload.GitCredentialID))
			if err != nil {
				b.err = httperror.InternalServerError("git credential not found", err)
				return nil, err
			}

			// When creating the stack with an existing git credential, the git credential must be owned by the calling user
			if credential.UserID != requestUserID {
				b.err = httperror.Forbidden("couldn't add the git credential of another user", httperrors.ErrUnauthorized)
				return nil, httperrors.ErrUnauthorized
			}

			repositoryUsername = credential.Username
			repositoryPassword = credential.Password
			repositoryGitCredentialID = payload.GitCredentialID
		}

		if payload.Password != "" {
			repositoryUsername = payload.Username
			repositoryPassword = payload.Password
			repositoryGitCredentialID = 0
		}

		return &gittypes.GitAuthentication{
			Username:        repositoryUsername,
			Password:        repositoryPassword,
			GitCredentialID: repositoryGitCredentialID,
		}, nil
	}
	return nil, nil
}

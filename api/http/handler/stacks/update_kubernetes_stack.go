package stacks

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/git/update"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type kubernetesFileStackUpdatePayload struct {
	StackFileContent string
	// RollbackTo specifies the stack file version to rollback to (only support to rollback to the last version currently)
	RollbackTo *int
}

type kubernetesGitStackUpdatePayload struct {
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	AutoUpdate               *portaineree.AutoUpdateSettings
	TLSSkipVerify            bool
}

func (payload *kubernetesFileStackUpdatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

func (payload *kubernetesGitStackUpdatePayload) Validate(r *http.Request) error {
	if err := update.ValidateAutoUpdateSettings(payload.AutoUpdate); err != nil {
		return err
	}
	return nil
}

func (handler *Handler) updateKubernetesStack(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint) *httperror.HandlerError {

	if stack.GitConfig != nil {
		//stop the autoupdate job if there is any
		if stack.AutoUpdate != nil {
			deployments.StopAutoupdate(stack.ID, stack.AutoUpdate.JobID, handler.Scheduler)
		}

		var payload kubernetesGitStackUpdatePayload

		if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
			return httperror.BadRequest("Invalid request payload", err)
		}

		stack.GitConfig.ReferenceName = payload.RepositoryReferenceName
		stack.GitConfig.TLSSkipVerify = payload.TLSSkipVerify
		stack.AutoUpdate = payload.AutoUpdate

		if payload.RepositoryAuthentication {
			password := payload.RepositoryPassword
			if password == "" && stack.GitConfig != nil && stack.GitConfig.Authentication != nil {
				password = stack.GitConfig.Authentication.Password
			}
			stack.GitConfig.Authentication = &gittypes.GitAuthentication{
				Username: payload.RepositoryUsername,
				Password: password,
			}
			_, err := handler.GitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, stack.GitConfig.Authentication.Username, stack.GitConfig.Authentication.Password, stack.GitConfig.TLSSkipVerify)
			if err != nil {
				return httperror.InternalServerError("Unable to fetch git repository", err)
			}
		} else {
			stack.GitConfig.Authentication = nil
		}

		if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
			jobID, e := deployments.StartAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService)
			if e != nil {
				return e
			}
			stack.AutoUpdate.JobID = jobID
		}

		return nil
	}

	var payload kubernetesFileStackUpdatePayload

	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.BadRequest("Failed to retrieve user token data", err)
	}

	tempFileDir, _ := os.MkdirTemp("", "kub_file_content")
	defer os.RemoveAll(tempFileDir)

	if err := filesystem.WriteToFile(filesystem.JoinPaths(tempFileDir, stack.EntryPoint), []byte(payload.StackFileContent)); err != nil {
		return httperror.InternalServerError("Failed to persist deployment file in a temp directory", err)
	}

	// Refresh ECR registry secret if needed
	// RefreshEcrSecret method checks if the namespace has any ECR registry
	// otherwise return nil
	cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err == nil {
		registryutils.RefreshEcrSecret(cli, endpoint, handler.DataStore, stack.Namespace)
	}

	//use temp dir as the stack project path for deployment
	//so if the deployment failed, the original file won't be over-written
	stack.ProjectPath = tempFileDir

	_, deployError := handler.deployKubernetesStack(tokenData, endpoint, stack, k.KubeAppLabels{
		StackID:   int(stack.ID),
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "content",
	})
	if deployError != nil {
		return deployError
	}

	// After deploying successfully, stack.ProjectPath should be set back
	stackFolder := strconv.Itoa(int(stack.ID))
	stack.ProjectPath = handler.FileService.GetStackProjectPath(stackFolder)

	// update or rollback stack file version
	err = handler.updateStackFileVersion(stack, payload.StackFileContent, payload.RollbackTo)
	if err != nil {
		return httperror.BadRequest("Unable to update or rollback kubernetes stack file version", err)
	}

	_, err = handler.FileService.StoreStackFileFromBytesByVersion(stackFolder,
		stack.EntryPoint,
		stack.StackFileVersion,
		[]byte(payload.StackFileContent))
	if err != nil {
		if rollbackErr := handler.FileService.RollbackStackFileByVersion(stackFolder, stack.StackFileVersion, stack.EntryPoint); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("rollback stack file error")
		}

		fileType := "Manifest"
		if stack.IsComposeFormat {
			fileType = "Compose"
		}
		errMsg := fmt.Sprintf("Unable to persist Kubernetes %s file on disk", fileType)
		return httperror.InternalServerError(errMsg, err)
	}

	handler.FileService.RemoveStackFileBackupByVersion(stackFolder, stack.StackFileVersion, stack.EntryPoint)

	return nil
}

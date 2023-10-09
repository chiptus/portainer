package stacks

import (
	"net/http"
	"strconv"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker/images"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	"github.com/portainer/portainer/api/filesystem"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type updateStackPayload struct {
	// New content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx"`
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// Prune services that are no longer referenced (only available for Swarm stacks)
	Prune bool `example:"true"`
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
	// Force a pulling to current image with the original tag though the image is already the latest
	PullImage bool `example:"false"`
	// RollbackTo specifies the stack file version to rollback to (only support to rollback to the last version currently)
	RollbackTo *int
}

func (payload *updateStackPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

// @id StackUpdate
// @summary Update a stack
// @description Update a stack, only for file based stacks.
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Stack identifier"
// @param endpointId query int true "Environment identifier"
// @param body body updateStackPayload true "Stack details"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id} [put]
func (handler *Handler) stackUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().Read(portaineree.StackID(stackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	}

	// TODO: this is a work-around for stacks created with Portainer version >= 1.17.1
	// The EndpointID property is not available for these stacks, this API endpoint
	// can use the optional EndpointID query parameter to associate a valid environment(endpoint) identifier to the stack.
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}
	if endpointID != int(stack.EndpointID) {
		stack.EndpointID = portaineree.EndpointID(endpointID)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the environment associated to the stack inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the environment associated to the stack inside the database", err)
	}
	middlewares.SetEndpoint(endpoint, r)

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	//only check resource control when it is a DockerSwarmStack or a DockerComposeStack
	if stack.Type == portaineree.DockerSwarmStack || stack.Type == portaineree.DockerComposeStack {
		resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve a resource control associated to the stack", err)
		}

		access, err := handler.userCanAccessStack(securityContext, endpoint.ID, resourceControl)
		if err != nil {
			return httperror.InternalServerError("Unable to verify user authorizations to validate stack access", err)
		}
		if !access {
			return httperror.Forbidden("Access denied to resource", httperrors.ErrResourceAccessDenied)
		}
	}

	canManage, err := handler.userCanManageStacks(securityContext, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to verify user authorizations to validate stack deletion", err)
	}
	if !canManage {
		errMsg := "Stack editing is disabled for non-admin users"
		return httperror.Forbidden(errMsg, errors.New(errMsg))
	}

	updateError := handler.updateAndDeployStack(r, stack, endpoint)
	if updateError != nil {
		return updateError
	}

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return httperror.BadRequest("Cannot find context user", errors.Wrap(err, "failed to fetch the user"))
	}
	stack.UpdatedBy = user.Username
	stack.UpdateDate = time.Now().Unix()
	stack.Status = portaineree.StackStatusActive

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil &&
		stack.GitConfig.Authentication.GitCredentialID != 0 {
		// prevent the username and password from saving into db if the git
		// credential is used
		stack.GitConfig.Authentication.Username = ""
		stack.GitConfig.Authentication.Password = ""
	}

	err = handler.DataStore.Stack().Update(stack.ID, stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

func (handler *Handler) updateAndDeployStack(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	if stack.Type == portaineree.DockerSwarmStack {
		stack.Name = handler.SwarmStackManager.NormalizeStackName(stack.Name)
		return handler.updateSwarmOrComposeStack(r, stack, endpoint)

	} else if stack.Type == portaineree.DockerComposeStack {
		stack.Name = handler.ComposeStackManager.NormalizeStackName(stack.Name)
		return handler.updateSwarmOrComposeStack(r, stack, endpoint)

	} else if stack.Type == portaineree.KubernetesStack {
		return handler.updateKubernetesStack(r, stack, endpoint)

	} else {
		return httperror.InternalServerError("Unsupported stack", errors.Errorf("unsupported stack type: %v", stack.Type))
	}
}

func (handler *Handler) updateSwarmOrComposeStack(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint) *httperror.HandlerError {

	// Must not be git based stack. stop the auto update job if there is any
	if stack.AutoUpdate != nil {
		deployments.StopAutoupdate(stack.ID, stack.AutoUpdate.JobID, handler.Scheduler)
		stack.AutoUpdate = nil
	}

	var payload updateStackPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	stack.Env = payload.Env

	if payload.Webhook != "" && stack.Webhook != payload.Webhook {
		isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
		if isUniqueError != nil {
			return isUniqueError
		}
	}
	stack.Webhook = payload.Webhook

	if stack.GitConfig != nil {
		httpErr := handler.updateStackDeployedByGitRepo(payload, stack)
		if httpErr != nil {
			return httpErr
		}
	} else {
		httpErr := handler.updateStackDeployedByFileContent(payload, stack)
		if httpErr != nil {
			return httpErr
		}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	// Create deployment config based on the stack type
	var stackDeploymentConfig deployments.StackDeploymentConfiger
	switch stack.Type {
	case portaineree.DockerSwarmStack:
		stackDeploymentConfig, err = deployments.CreateSwarmStackDeploymentConfig(securityContext,
			stack,
			endpoint,
			handler.DataStore,
			handler.FileService,
			handler.StackDeployer,
			payload.Prune,
			payload.PullImage)

	case portaineree.DockerComposeStack:
		stackDeploymentConfig, err = deployments.CreateComposeStackDeploymentConfig(securityContext,
			stack,
			endpoint,
			handler.DataStore,
			handler.FileService,
			handler.StackDeployer,
			payload.PullImage,
			false)

	default:
		return httperror.InternalServerError("Invalid stack type", errors.New("invalid stack type"))
	}
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	// Deploy the stack
	err = stackDeploymentConfig.Deploy()
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	go func() {
		images.EvictImageStatus(stack.Name)
		EvictSwarmStackImageStatusCache(r.Context(), endpoint, stack.Name, handler.DockerClientFactory)
	}()
	return nil
}

func isFromAppTemplateStackMode(stack *portaineree.Stack) bool {
	return stack.FromAppTemplate && stack.GitConfig != nil
}

// updateStackDeployedByGitRepo updates the stack deployed from git repo
// There are two situations that a stack deployed from git repo will be
// updated by this function:
// 1. the stack is deployed from app template stack mode
// 2. the git repo is detached from the stack
// After the first time update, the stack will be detached from git repo
// and treated as a regular stack deployed from file content
func (handler *Handler) updateStackDeployedByGitRepo(payload updateStackPayload, stack *portaineree.Stack) *httperror.HandlerError {
	if stack.GitConfig == nil {
		return httperror.BadRequest("Invalid stack", errors.New("stack is not deployed from git repo"))
	}

	stackFolder := strconv.Itoa(int(stack.ID))

	// folder structure before update: "/data/compose/1/f94735d43bc9a046bc0f6a794f588140db860742"
	oldProjectPath := handler.FileService.GetStackProjectPathByVersion(stackFolder, stack.StackFileVersion, stack.GitConfig.ConfigHash)

	// update or rollback stack file version
	err := handler.updateStackFileVersion(stack, payload.StackFileContent, payload.RollbackTo)
	if err != nil {
		return httperror.BadRequest("Unable to update or rollback stack file version", err)
	}

	// folder structure after update: "/data/compose/1/v1"
	newProjectPath := handler.FileService.GetStackProjectPathByVersion(stackFolder, stack.StackFileVersion, "")

	// rename "f94735d43bc9a046bc0f6a794f588140db860742" to "v1"
	err = handler.FileService.SafeMoveDirectory(oldProjectPath, newProjectPath)
	if err != nil {
		return httperror.InternalServerError("Failed to move the stack folder", err)
	}

	_, err = handler.FileService.UpdateStoreStackFileFromBytesByVersion(stackFolder,
		stack.EntryPoint,
		stack.StackFileVersion,
		"",
		[]byte(payload.StackFileContent))
	if err != nil {
		if rollbackErr := handler.FileService.RollbackStackFileByVersion(stackFolder, stack.StackFileVersion, stack.EntryPoint); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("rollback stack file error")
		}

		return httperror.InternalServerError("Unable to persist updated Compose file on disk", err)
	}

	stack.GitConfig = nil
	stack.PreviousDeploymentInfo = nil
	stack.IsDetachedFromGit = true

	handler.FileService.RemoveStackFileBackupByVersion(stackFolder, stack.StackFileVersion, stack.EntryPoint)

	return nil
}

// updateStackDeployedByFileContent updates the stack deployed from file
// content (non-git repo method)
func (handler *Handler) updateStackDeployedByFileContent(payload updateStackPayload, stack *portaineree.Stack) *httperror.HandlerError {
	if stack.GitConfig != nil {
		return httperror.BadRequest("Invalid stack", errors.New("stack is deployed from git repo"))
	}

	stackFolder := strconv.Itoa(int(stack.ID))

	folderToBeRemoved := ""
	if stack.PreviousDeploymentInfo != nil && payload.RollbackTo == nil && stack.IsDetachedFromGit {
		folderToBeRemoved = handler.FileService.GetStackProjectPathByVersion(stackFolder, stack.PreviousDeploymentInfo.FileVersion, "")
	}

	// 1. Record the current stack version folder
	// file structure before update: "/data/compose/1/v5"
	oldProjectPath := handler.FileService.GetStackProjectPathByVersion(stackFolder, stack.StackFileVersion, "")

	// 2. update stack file version
	err := handler.updateStackFileVersion(stack, payload.StackFileContent, payload.RollbackTo)
	if err != nil {
		return httperror.BadRequest("Unable to update or rollback stack file version", err)
	}

	// 3. Copy the existing folder to the new stack version folder
	// file structure after update: "/data/compose/1/v6"
	newProjectPath := handler.FileService.GetStackProjectPathByVersion(stackFolder, stack.StackFileVersion, "")
	err = filesystem.CopyDir(oldProjectPath, newProjectPath, false)
	if err != nil {
		return httperror.InternalServerError("Failed to copy the stack folder", err)
	}

	// 4. Update the stack file content in the new stack version folder
	_, err = handler.FileService.UpdateStoreStackFileFromBytesByVersion(stackFolder,
		stack.EntryPoint,
		stack.StackFileVersion,
		"",
		[]byte(payload.StackFileContent))
	if err != nil {
		if rollbackErr := handler.FileService.RollbackStackFileByVersion(stackFolder, stack.StackFileVersion, stack.EntryPoint); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("rollback stack file error")
		}

		return httperror.InternalServerError("Unable to persist updated Compose file on disk", err)
	}

	// 5. Remove the stack version folders except current and previous version
	if folderToBeRemoved != "" {
		err = handler.FileService.RemoveDirectory(folderToBeRemoved)
		if err != nil {
			log.Info().Err(err).Msg("failed to remove the stack version folder")
		}
	}

	handler.FileService.RemoveStackFileBackupByVersion(stackFolder, stack.StackFileVersion, stack.EntryPoint)

	return nil
}

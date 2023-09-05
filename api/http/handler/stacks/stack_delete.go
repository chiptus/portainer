package stacks

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	"github.com/portainer/portainer/api/filesystem"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// @id StackDelete
// @summary Remove a stack
// @description Remove a stack.
// @description **Access policy**: restricted
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Stack identifier"
// @param external query boolean false "Set to true to delete an external stack. Only external Swarm stacks are supported"
// @param endpointId query int true "Environment identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id} [delete]
func (handler *Handler) stackDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	externalStack, _ := request.RetrieveBooleanQueryParameter(r, "external", true)
	if externalStack {
		return handler.deleteExternalStack(r, w, stackID, securityContext)
	}

	id, err := strconv.Atoi(stackID)
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().Read(portaineree.StackID(id))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", true)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}
	isOrphaned := portaineree.EndpointID(endpointID) != stack.EndpointID

	if isOrphaned && !securityContext.IsAdmin {
		return httperror.Forbidden("Permission denied to remove orphaned stack", errors.New("Permission denied to remove orphaned stack"))
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the endpoint associated to the stack inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the endpoint associated to the stack inside the database", err)
	}
	middlewares.SetEndpoint(endpoint, r)

	var resourceControl *portaineree.ResourceControl
	if stack.Type == portaineree.DockerSwarmStack || stack.Type == portaineree.DockerComposeStack {
		resourceControl, err = handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve a resource control associated to the stack", err)
		}
	}

	if !isOrphaned {
		err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
		if err != nil {
			return httperror.Forbidden("Permission denied to access endpoint", err)
		}

		if stack.Type == portaineree.DockerSwarmStack || stack.Type == portaineree.DockerComposeStack {
			access, err := handler.userCanAccessStack(securityContext, endpoint.ID, resourceControl)
			if err != nil {
				return httperror.InternalServerError("Unable to verify user authorizations to validate stack access", err)
			}
			if !access {
				return httperror.Forbidden("Access denied to resource", httperrors.ErrResourceAccessDenied)
			}
		}
	}

	canManage, err := handler.userCanManageStacks(securityContext, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to verify user authorizations to validate stack deletion", err)
	}
	if !canManage {
		errMsg := "stack deletion is disabled for non-admin users"
		return httperror.Forbidden(errMsg, fmt.Errorf(errMsg))
	}

	// stop scheduler updates of the stack before removal
	if stack.AutoUpdate != nil {
		deployments.StopAutoupdate(stack.ID, stack.AutoUpdate.JobID, handler.Scheduler)
	}

	err = handler.deleteStack(securityContext.UserID, stack, endpoint)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	err = handler.DataStore.Stack().Delete(portaineree.StackID(id))
	if err != nil {
		return httperror.InternalServerError("Unable to remove the stack from the database", err)
	}

	if resourceControl != nil {
		err = handler.DataStore.ResourceControl().Delete(resourceControl.ID)
		if err != nil {
			return httperror.InternalServerError("Unable to remove the associated resource control from the database", err)
		}
	}

	err = handler.FileService.RemoveDirectory(stack.ProjectPath)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to remove stack files from disk")
	}

	return response.Empty(w)
}

func (handler *Handler) deleteExternalStack(r *http.Request, w http.ResponseWriter, stackName string, securityContext *security.RestrictedRequestContext) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return httperror.InternalServerError("Unable to load user information from the database", err)
	}

	_, endpointResourceAccess := user.EndpointAuthorizations[portaineree.EndpointID(endpointID)][portaineree.EndpointResourcesAccess]

	if !securityContext.IsAdmin && !endpointResourceAccess {
		return &httperror.HandlerError{StatusCode: http.StatusUnauthorized, Message: "Permission denied to delete the stack", Err: httperrors.ErrUnauthorized}
	}

	stack, err := handler.DataStore.Stack().StackByName(stackName)
	if err != nil && !handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.InternalServerError("Unable to check for stack existence inside the database", err)
	}
	if stack != nil {
		return httperror.BadRequest("A stack with this name exists inside the database. Cannot use external delete method", errors.New("A tag already exists with this name"))
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the endpoint associated to the stack inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the endpoint associated to the stack inside the database", err)
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access endpoint", err)
	}

	stack = &portaineree.Stack{
		Name: stackName,
		Type: portaineree.DockerSwarmStack,
	}

	err = handler.deleteStack(securityContext.UserID, stack, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to delete stack", err)
	}

	return response.Empty(w)
}

func (handler *Handler) deleteStack(userID portaineree.UserID, stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	if stack.Type == portaineree.DockerSwarmStack {
		stack.Name = handler.SwarmStackManager.NormalizeStackName(stack.Name)

		if stackutils.IsRelativePathStack(stack) {
			return handler.StackDeployer.UndeployRemoteSwarmStack(stack, endpoint)
		}

		return handler.SwarmStackManager.Remove(stack, endpoint)
	}

	if stack.Type == portaineree.DockerComposeStack {
		stack.Name = handler.ComposeStackManager.NormalizeStackName(stack.Name)

		if stackutils.IsRelativePathStack(stack) {
			return handler.StackDeployer.UndeployRemoteComposeStack(stack, endpoint)
		}

		return handler.ComposeStackManager.Down(context.TODO(), stack, endpoint)
	}

	if stack.Type == portaineree.KubernetesStack {
		var manifestFiles []string

		// use manifest files to remove kube resources;
		// if stack was created with compose files, convert them to kube manifests first
		if stack.IsComposeFormat {
			fileNames := stackutils.GetStackFilePaths(stack, false)
			tmpDir, err := os.MkdirTemp("", "kube_delete")
			if err != nil {
				return errors.Wrap(err, "failed to create temp directory for deleting kub stack")
			}

			defer os.RemoveAll(tmpDir)

			projectPath := stackutils.GetStackProjectPathByVersion(stack)
			for _, fileName := range fileNames {
				manifestFilePath := filesystem.JoinPaths(tmpDir, fileName)
				manifestContent, err := handler.FileService.GetFileContent(projectPath, fileName)
				if err != nil {
					return errors.Wrap(err, "failed to read manifest file")
				}

				manifestContent, err = handler.KubernetesDeployer.ConvertCompose(manifestContent)
				if err != nil {
					return errors.Wrap(err, "failed to convert docker compose file to a kube manifest")
				}

				err = filesystem.WriteToFile(manifestFilePath, []byte(manifestContent))
				if err != nil {
					return errors.Wrap(err, "failed to create temp manifest file")
				}
				manifestFiles = append(manifestFiles, manifestFilePath)
			}
		} else {
			manifestFiles = stackutils.GetStackFilePaths(stack, true)
		}

		out, err := handler.KubernetesDeployer.Remove(userID, endpoint, manifestFiles, stack.Namespace)

		return errors.WithMessagef(err, "failed to remove kubernetes resources: %q", out)
	}

	return fmt.Errorf("unsupported stack type: %v", stack.Type)
}

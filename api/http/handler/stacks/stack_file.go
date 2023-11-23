package stacks

import (
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
)

type stackFileResponse struct {
	// Content of the Stack file
	StackFileContent string `json:"StackFileContent" example:"version: 3\n services:\n web:\n image:nginx"`
}

// @id StackFileInspect
// @summary Retrieve the content of the Stack file for the specified stack
// @description Get Stack file content.
// @description **Access policy**: restricted
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Stack identifier"
// @param version query int false "Stack file version maintained by Portainer. If both version and commitHash are provided, the commitHash will be used"
// @param commitHash query string false "Git repository commit hash. If both version and commitHash are provided, the commitHash will be used"
// @success 200 {object} stackFileResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Stack not found"
// @failure 500 "Server error"
// @router /stacks/{id}/file [get]
func (handler *Handler) stackFile(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().Read(portainer.StackID(stackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		// ignore not found error for admins and edge admins to manage orphaned stacks
		if !security.IsAdminOrEdgeAdminContext(securityContext) {
			return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
		}
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	if endpoint != nil {
		err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
		if err != nil {
			return httperror.Forbidden("Permission denied to access environment", err)
		}

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
	}

	canManage, err := handler.userCanManageStacks(securityContext, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to verify user authorizations to validate stack deletion", err)
	}
	if !canManage {
		errMsg := "Stack management is disabled for non-admin users"
		return httperror.Forbidden(errMsg, errors.New(errMsg))
	}

	// Get project path
	var projectPath string
	if stack.GitConfig != nil {
		projectPath = handler.FileService.FormProjectPathByVersion(stack.ProjectPath, stack.StackFileVersion, stack.GitConfig.ConfigHash)

		// check if a commit hash is provided
		commitHash, _ := request.RetrieveQueryParameter(r, "commitHash", true)
		if commitHash != "" {
			if (stack.PreviousDeploymentInfo != nil && commitHash != stack.PreviousDeploymentInfo.ConfigHash) &&
				commitHash != stack.GitConfig.ConfigHash {
				return httperror.BadRequest("Only support latest two versions", fmt.Errorf("commit hash %s is not a valid commit hash for this stack", commitHash))
			}
			projectPath = handler.FileService.FormProjectPathByVersion(stack.ProjectPath, stack.StackFileVersion, commitHash)
		}
	} else {
		projectPath = handler.FileService.FormProjectPathByVersion(stack.ProjectPath, stack.StackFileVersion, "")

		// check if a version is provided
		version, _ := request.RetrieveNumericQueryParameter(r, "version", true)
		if version != 0 {
			if (stack.PreviousDeploymentInfo != nil && version != stack.PreviousDeploymentInfo.FileVersion) &&
				version != stack.StackFileVersion {
				return httperror.BadRequest("Only support latest two versions", fmt.Errorf("version %d is not a valid version for this stack", version))
			}
			projectPath = handler.FileService.FormProjectPathByVersion(stack.ProjectPath, version, "")
		}
	}

	stackFileContent, err := handler.FileService.GetFileContent(projectPath, stack.EntryPoint)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Compose file from disk", err)
	}

	return response.JSON(w, &stackFileResponse{StackFileContent: string(stackFileContent)})
}

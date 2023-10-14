package edgestacks

import (
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type stackFileResponse struct {
	StackFileContent string `json:"StackFileContent"`
}

// @id EdgeStackFile
// @summary Fetches the stack file for an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "EdgeStack Id"
// @param version query int false "Stack file version maintained by Portainer. If both version and commitHash are provided, the commitHash will be used"
// @param commitHash query string false "Git repository commit hash. If both version and commitHash are provided, the commitHash will be used"
// @success 200 {object} stackFileResponse
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/file [get]
func (handler *Handler) edgeStackFile(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	stack, err := handler.DataStore.EdgeStack().EdgeStack(portainer.EdgeStackID(stackID))
	if err != nil {
		return handler.handlerDBErr(err, "Unable to find an edge stack with the specified identifier inside the database")
	}

	fileName := stack.EntryPoint
	if stack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		fileName = stack.ManifestPath
	}

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

	stackFileContent, err := handler.FileService.GetFileContent(projectPath, fileName)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve stack file from disk", err)
	}

	return response.JSON(w, &stackFileResponse{StackFileContent: string(stackFileContent)})
}

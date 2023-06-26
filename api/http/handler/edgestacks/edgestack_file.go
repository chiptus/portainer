package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
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
// @param version query int false "Stack version"
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

	stack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(stackID))
	if err != nil {
		return handler.handlerDBErr(err, "Unable to find an edge stack with the specified identifier inside the database")
	}

	fileName := stack.EntryPoint
	if stack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		fileName = stack.ManifestPath
	}

	projectPath := stack.ProjectPath
	if stack.GitConfig == nil {
		projectPath = handler.FileService.FormProjectPathByVersion(stack.ProjectPath, stack.Version)
		// check if a version is provided
		version, _ := request.RetrieveNumericQueryParameter(r, "version", true)
		if version != 0 {
			if version != stack.PreviousDeploymentInfo.Version && version != stack.Version {
				return httperror.BadRequest("Invalid version", err)
			}
			projectPath = handler.FileService.FormProjectPathByVersion(stack.ProjectPath, version)
		}
	}

	stackFileContent, err := handler.FileService.GetFileContent(projectPath, fileName)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve stack file from disk", err)
	}

	return response.JSON(w, &stackFileResponse{StackFileContent: string(stackFileContent)})
}

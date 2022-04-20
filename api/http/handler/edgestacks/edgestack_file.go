package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
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
// @param id path string true "EdgeStack Id"
// @success 200 {object} stackFileResponse
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/file [get]
func (handler *Handler) edgeStackFile(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid edge stack identifier route variable", Err: err}
	}

	stack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(stackID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an edge stack with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an edge stack with the specified identifier inside the database", Err: err}
	}

	fileName := stack.EntryPoint
	if stack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		fileName = stack.ManifestPath
	}

	stackFileContent, err := handler.FileService.GetFileContent(stack.ProjectPath, fileName)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve stack file from disk", Err: err}
	}

	return response.JSON(w, &stackFileResponse{StackFileContent: string(stackFileContent)})
}

package endpointedge

import (
	"errors"
	"net/http"
	"os"
	"path"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/api/edge"
)

// @summary Inspect an Edge Stack for an Environment(Endpoint)
// @description **Access policy**: public
// @tags edge, endpoints, edge_stacks
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) Id"
// @param stackId path int true "EdgeStack Id"
// @success 200 {object} edge.StackPayload
// @failure 500
// @failure 400
// @failure 404
// @router /endpoints/{id}/edge/stacks/{stackId} [get]
func (handler *Handler) endpointEdgeStackInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "stackId")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(edgeStackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an edge stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an edge stack with the specified identifier inside the database", err)
	}

	fileName := edgeStack.EntryPoint
	if endpointutils.IsDockerEndpoint(endpoint) {
		if fileName == "" {
			return httperror.BadRequest("Docker is not supported by this stack", errors.New("Docker is not supported by this stack"))
		}
	}

	namespace := ""
	if !edgeStack.UseManifestNamespaces {
		namespace = kubernetes.DefaultNamespace
	}

	if endpointutils.IsKubernetesEndpoint(endpoint) {
		fileName = edgeStack.ManifestPath

		if fileName == "" {
			return httperror.BadRequest("Kubernetes is not supported by this stack", errors.New("Kubernetes is not supported by this stack"))
		}

	}

	stackFileContent, err := handler.FileService.GetFileContent(edgeStack.ProjectPath, fileName)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Compose file from disk", err)
	}

	var dotEnvFileContent []byte
	if _, err = os.Stat(path.Join(edgeStack.ProjectPath, ".env")); err == nil {
		dotEnvFileContent, err = handler.FileService.GetFileContent(edgeStack.ProjectPath, ".env")
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve .env file from disk", err)
		}
	}

	registryCredentials := registryutils.GetRegistryCredentialsForEdgeStack(handler.DataStore, edgeStack, endpoint)

	return response.JSON(w, edge.StackPayload{
		FileContent:         string(stackFileContent),
		DotEnvFileContent:   string(dotEnvFileContent),
		Name:                edgeStack.Name,
		RegistryCredentials: registryCredentials,
		Namespace:           namespace,
		PrePullImage:        edgeStack.PrePullImage,
		RePullImage:         edgeStack.RePullImage,
		RetryDeploy:         edgeStack.RetryDeploy,
		EdgeUpdateID:        edgeStack.EdgeUpdateID,
	})
}

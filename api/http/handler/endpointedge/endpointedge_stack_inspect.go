package endpointedge

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	internaledge "github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	"github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/api/edge"
	"github.com/portainer/portainer/api/filesystem"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

// @summary Inspect an Edge Stack for an Environment(Endpoint)
// @description **Access policy**: public
// @tags edge, endpoints, edge_stacks
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) Id"
// @param stackId path int true "EdgeStack Id"
// @param version query int false "Stack file version maintained by Portainer"
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

	var (
		projectVersionPath string
		deployedVersion    int
	)
	version, _ := request.RetrieveNumericQueryParameter(r, "version", true)

	rollbackTo := new(int)
	// Check if the requested version is the previous version
	if edgeStack.PreviousDeploymentInfo != nil && version == edgeStack.PreviousDeploymentInfo.FileVersion {
		projectVersionPath = handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, edgeStack.PreviousDeploymentInfo.FileVersion, edgeStack.PreviousDeploymentInfo.ConfigHash)
		*rollbackTo = edgeStack.PreviousDeploymentInfo.FileVersion
		deployedVersion = edgeStack.PreviousDeploymentInfo.FileVersion
	} else {
		if version != 0 && version != edgeStack.StackFileVersion {
			log.Debug().Msgf("[Stagger] Invalid stack file version %d being requested, fallback to the latest stack file version %d", version, edgeStack.StackFileVersion)
		}

		// If the requested version is not the previous version, use the latest stack file version
		commitHash := ""
		if edgeStack.GitConfig != nil {
			commitHash = edgeStack.GitConfig.ConfigHash
		}
		projectVersionPath = handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, edgeStack.StackFileVersion, commitHash)
		rollbackTo = nil
		deployedVersion = version
	}

	dirEntries, err := filesystem.LoadDir(projectVersionPath)
	if err != nil {
		return httperror.InternalServerError("Unable to load repository", err)
	}

	fileContent, err := filesystem.FilterDirForCompatibility(dirEntries, fileName, endpoint.Agent.Version)
	if err != nil {
		return httperror.InternalServerError("File not found", err)
	}

	dirEntries, err = internaledge.FilterEntriesForEdgeStack(
		handler.DataStore,
		edgeStack,
		endpoint,
		dirEntries,
		fileName,
	)
	if err != nil {
		return httperror.InternalServerError("Unable to filter dir entries for edge stack", err)
	}

	registryCredentials := registryutils.GetRegistryCredentialsForEdgeStack(handler.DataStore, edgeStack, endpoint)

	return response.JSON(w, edge.StackPayload{
		DirEntries:          dirEntries,
		EntryFileName:       fileName,
		StackFileContent:    fileContent,
		SupportRelativePath: edgeStack.SupportRelativePath,
		FilesystemPath:      edgeStack.FilesystemPath,
		Name:                edgeStack.Name,
		RegistryCredentials: registryCredentials,
		Namespace:           namespace,
		Version:             deployedVersion,
		RollbackTo:          rollbackTo,
		PrePullImage:        edgeStack.PrePullImage,
		RePullImage:         edgeStack.RePullImage,
		RetryDeploy:         edgeStack.RetryDeploy,
		EdgeUpdateID:        edgeStack.EdgeUpdateID,
		EnvVars:             edgeStack.EnvVars,
	})
}

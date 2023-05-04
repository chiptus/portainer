package edgestacks

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer/api/filesystem"
)

type edgeStackFromStringPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// List of identifiers of EdgeGroups
	EdgeGroups []portaineree.EdgeGroupID `example:"1"`
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'
	// for compose stacks will use kompose to convert to kubernetes manifest for kubernetes environments(endpoints)
	// kubernetes deploy type is enabled only for kubernetes environments(endpoints)
	// nomad deploy type is enabled only for nomad environments(endpoints)
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portaineree.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	// Pre Pull image
	PrePullImage bool `example:"false"`
	// Retry deploy
	RetryDeploy bool `example:"false"`
}

func (payload *edgeStackFromStringPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return httperrors.NewInvalidPayloadError("Invalid stack name")
	}
	if govalidator.IsNull(payload.StackFileContent) {
		return httperrors.NewInvalidPayloadError("Invalid stack file content")
	}
	if len(payload.EdgeGroups) == 0 {
		return httperrors.NewInvalidPayloadError("Edge Groups are mandatory for an Edge stack")
	}

	return nil
}

// @id EdgeStackCreateString
// @summary Create an EdgeStack from a text
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param body body edgeStackFromStringPayload true "stack config"
// @param dryrun query string false "if true, will not create an edge stack, but just will check the settings and return a non-persisted edge stack object"
// @success 200 {object} portaineree.EdgeStack
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/create/string [post]
func (handler *Handler) createEdgeStackFromFileContent(r *http.Request, dryrun bool) (*portaineree.EdgeStack, error) {
	var payload edgeStackFromStringPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	stack, err := handler.edgeStacksService.BuildEdgeStack(payload.Name, payload.DeploymentType, payload.EdgeGroups, payload.Registries, "", payload.UseManifestNamespaces, payload.PrePullImage, false, payload.RetryDeploy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Edge stack object")
	}

	if len(payload.Registries) == 0 && dryrun {
		err = handler.assignPrivateRegistriesToStack(stack, bytes.NewReader([]byte(payload.StackFileContent)))
		if err != nil {
			return nil, errors.Wrap(err, "failed to assign private registries to stack")
		}
	}

	if dryrun {
		return stack, nil
	}

	return handler.edgeStacksService.PersistEdgeStack(stack, func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (configPath string, manifestPath string, projectPath string, err error) {
		return handler.storeFileContent(stackFolder, payload.DeploymentType, relatedEndpointIds, []byte(payload.StackFileContent))
	})
}

func (handler *Handler) storeFileContent(stackFolder string, deploymentType portaineree.EdgeStackDeploymentType, relatedEndpointIds []portaineree.EndpointID, fileContent []byte) (composePath, manifestPath, projectPath string, err error) {
	if deploymentType == portaineree.EdgeStackDeploymentCompose {
		composePath = filesystem.ComposeFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, composePath, fileContent)
		if err != nil {
			return "", "", "", err
		}

		manifestPath, err = handler.convertAndStoreKubeManifestIfNeeded(stackFolder, projectPath, composePath, relatedEndpointIds)
		if err != nil {
			return "", "", "", fmt.Errorf("Failed creating and storing kube manifest: %w", err)
		}

		return composePath, manifestPath, projectPath, nil
	}

	hasDockerEndpoint, err := hasDockerEndpoint(handler.DataStore.Endpoint(), relatedEndpointIds)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to check for existence of docker environment: %w", err)
	}

	if hasDockerEndpoint {
		return "", "", "", errors.New("edge stack with docker environment cannot be deployed with kubernetes or nomad config")
	}

	if deploymentType == portaineree.EdgeStackDeploymentKubernetes {

		manifestPath = filesystem.ManifestFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, manifestPath, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return "", manifestPath, projectPath, nil

	}

	if deploymentType == portaineree.EdgeStackDeploymentNomad {

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, eefs.NomadJobFileDefaultName, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return eefs.NomadJobFileDefaultName, "", projectPath, nil
	}

	errMessage := fmt.Sprintf("invalid deployment type: %d", deploymentType)
	return "", "", "", httperrors.NewInvalidPayloadError(errMessage)
}

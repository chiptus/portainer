package edgestacks

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/edge/staggers"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
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
	// compose is enabled only for docker environments
	// kubernetes is enabled only for kubernetes environments
	// nomad is enabled only for nomad environments
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portaineree.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	// Pre Pull image
	PrePullImage bool `example:"false"`
	// Retry deploy
	RetryDeploy bool `example:"false"`
	// Optional webhook configuration
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
	// List of environment variables
	EnvVars []portainer.Pair
	// Configuration for stagger updates
	StaggerConfig *portaineree.EdgeStaggerConfig
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

	if payload.DeploymentType != portaineree.EdgeStackDeploymentCompose && payload.DeploymentType != portaineree.EdgeStackDeploymentKubernetes && payload.DeploymentType != portaineree.EdgeStackDeploymentNomad {
		return httperrors.NewInvalidPayloadError("Invalid deployment type")
	}

	return staggers.ValidateStaggerConfig(payload.StaggerConfig)
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
func (handler *Handler) createEdgeStackFromFileContent(r *http.Request, tx dataservices.DataStoreTx, dryrun bool) (*portaineree.EdgeStack, error) {
	var payload edgeStackFromStringPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	if payload.Webhook != "" {
		err := handler.checkUniqueWebhookID(payload.Webhook)
		if err != nil {
			return nil, err
		}
	}

	buildEdgeStackArgs := edgestacks.BuildEdgeStackArgs{
		Registries:            payload.Registries,
		ScheduledTime:         "",
		UseManifestNamespaces: payload.UseManifestNamespaces,
		PrePullImage:          payload.PrePullImage,
		RePullImage:           false,
		RetryDeploy:           payload.RetryDeploy,
		EnvVars:               payload.EnvVars,
	}

	stack, err := handler.edgeStacksService.BuildEdgeStack(tx, payload.Name, payload.DeploymentType, payload.EdgeGroups, buildEdgeStackArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Edge stack object")
	}

	if len(payload.Registries) == 0 && dryrun {
		err = handler.assignPrivateRegistriesToStack(tx, stack, bytes.NewReader([]byte(payload.StackFileContent)))
		if err != nil {
			return nil, errors.Wrap(err, "failed to assign private registries to stack")
		}
	}

	stack.Webhook = payload.Webhook
	stack.StaggerConfig = payload.StaggerConfig

	if dryrun {
		return stack, nil
	}

	return handler.edgeStacksService.PersistEdgeStack(tx, stack, func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (configPath string, manifestPath string, projectPath string, err error) {
		return handler.storeFileContent(tx, stackFolder, payload.DeploymentType, relatedEndpointIds, []byte(payload.StackFileContent))
	})
}

func (handler *Handler) storeFileContent(tx dataservices.DataStoreTx, stackFolder string, deploymentType portaineree.EdgeStackDeploymentType, relatedEndpointIds []portaineree.EndpointID, fileContent []byte) (composePath, manifestPath, projectPath string, err error) {
	hasWrongType, err := hasWrongEnvironmentType(tx.Endpoint(), relatedEndpointIds, deploymentType)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to check for existence of non fitting environments: %w", err)
	}
	if hasWrongType {
		return "", "", "", fmt.Errorf("edge stack with config do not match the environment type")
	}

	projectPath = handler.FileService.GetEdgeStackProjectPath(stackFolder)
	initialStackFileVersion := 1 // When creating a new stack, the version is always 1
	if deploymentType == portaineree.EdgeStackDeploymentCompose {
		composePath = filesystem.ComposeFileDefaultName

		_, err = handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, composePath, initialStackFileVersion, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return composePath, "", projectPath, nil
	}

	if deploymentType == portaineree.EdgeStackDeploymentKubernetes {
		manifestPath = filesystem.ManifestFileDefaultName

		_, err = handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, manifestPath, initialStackFileVersion, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return "", manifestPath, projectPath, nil
	}

	if deploymentType == portaineree.EdgeStackDeploymentNomad {
		nomadConfigPath := eefs.NomadJobFileDefaultName

		_, err = handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, nomadConfigPath, initialStackFileVersion, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return nomadConfigPath, "", projectPath, nil
	}

	errMessage := fmt.Sprintf("invalid deployment type: %d", deploymentType)
	return "", "", "", httperrors.NewInvalidPayloadError(errMessage)
}

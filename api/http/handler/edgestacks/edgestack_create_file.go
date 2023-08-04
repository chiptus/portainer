package edgestacks

import (
	"bytes"
	"net/http"

	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/edge/staggers"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
)

type edgeStackFromFileUploadPayload struct {
	Name             string
	StackFileContent []byte
	EdgeGroups       []portaineree.EdgeGroupID
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'
	// compose is enabled only for docker environments
	// kubernetes is enabled only for kubernetes environments
	// nomad is enabled only for nomad environments
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	Registries     []portaineree.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	// Pre Pull image
	PrePullImage bool `example:"false"`
	// Retry deploy
	RetryDeploy bool `example:"false"`

	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
	// List of environment variables
	EnvVars []portainer.Pair
	// Configuration for stagger updates
	StaggerConfig *portaineree.EdgeStaggerConfig
}

func (payload *edgeStackFromFileUploadPayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return httperrors.NewInvalidPayloadError("Invalid stack name")
	}
	payload.Name = name

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return httperrors.NewInvalidPayloadError("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var edgeGroups []portaineree.EdgeGroupID
	err = request.RetrieveMultiPartFormJSONValue(r, "EdgeGroups", &edgeGroups, false)
	if err != nil || len(edgeGroups) == 0 {
		return httperrors.NewInvalidPayloadError("Edge Groups are mandatory for an Edge stack")
	}
	payload.EdgeGroups = edgeGroups

	deploymentType, err := request.RetrieveNumericMultiPartFormValue(r, "DeploymentType", false)
	if err != nil {
		return httperrors.NewInvalidPayloadError("Invalid deployment type")
	}
	payload.DeploymentType = portaineree.EdgeStackDeploymentType(deploymentType)
	if payload.DeploymentType != portaineree.EdgeStackDeploymentCompose && payload.DeploymentType != portaineree.EdgeStackDeploymentKubernetes && payload.DeploymentType != portaineree.EdgeStackDeploymentNomad {
		return httperrors.NewInvalidPayloadError("Invalid deployment type")
	}

	var registries []portaineree.RegistryID
	err = request.RetrieveMultiPartFormJSONValue(r, "Registries", &registries, true)
	if err != nil {
		return httperrors.NewInvalidPayloadError("Invalid registry type")
	}
	payload.Registries = registries

	useManifestNamespaces, _ := request.RetrieveBooleanMultiPartFormValue(r, "UseManifestNamespaces", true)
	payload.UseManifestNamespaces = useManifestNamespaces

	prePullImage, _ := request.RetrieveBooleanMultiPartFormValue(r, "PrePullImage", true)
	payload.PrePullImage = prePullImage

	retryDeploy, _ := request.RetrieveBooleanMultiPartFormValue(r, "RetryDeploy", true)
	payload.RetryDeploy = retryDeploy

	envVars := make([]portainer.Pair, 0)
	err = request.RetrieveMultiPartFormJSONValue(r, "EnvVars", &envVars, true)
	if err != nil {
		return httperrors.NewInvalidPayloadError("Invalid environment variables")
	}
	payload.EnvVars = envVars

	return staggers.ValidateStaggerConfig(payload.StaggerConfig)
}

// @id EdgeStackCreateFile
// @summary Create an EdgeStack from file
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @accept multipart/form-data
// @produce json
// @param Name formData string true "Name of the stack"
// @param file formData file true "Content of the Stack file"
// @param EdgeGroups formData string true "JSON stringified array of Edge Groups ids"
// @param DeploymentType formData int true "deploy type 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'"
// @param Registries formData string false "JSON stringified array of Registry ids to use for this stack"
// @param UseManifestNamespaces formData bool false "Uses the manifest's namespaces instead of the default one, relevant only for kube environments"
// @param PrePullImage formData bool false "Pre Pull image"
// @param RetryDeploy formData bool false "Retry deploy"
// @param dryrun query string false "if true, will not create an edge stack, but just will check the settings and return a non-persisted edge stack object"
// @param EnvVars formData string false "JSON stringified array of environment variables {name, value}"
// @success 200 {object} portaineree.EdgeStack
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/create/file [post]
func (handler *Handler) createEdgeStackFromFileUpload(r *http.Request, tx dataservices.DataStoreTx, dryrun bool) (*portaineree.EdgeStack, error) {
	payload := &edgeStackFromFileUploadPayload{}
	err := payload.Validate(r)
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
		return nil, errors.Wrap(err, "failed to create edge stack object")
	}

	if len(payload.Registries) == 0 && dryrun {
		err = handler.assignPrivateRegistriesToStack(tx, stack, bytes.NewReader(payload.StackFileContent))
		if err != nil {
			return nil, errors.Wrap(err, "failed to assign private registries to stack")
		}
	}

	stack.Webhook = payload.Webhook
	stack.StaggerConfig = payload.StaggerConfig

	if dryrun {
		return stack, nil
	}

	return handler.edgeStacksService.PersistEdgeStack(
		tx,
		stack,
		func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (configPath string, manifestPath string, projectPath string, err error) {
			return handler.storeFileContent(tx, stackFolder, payload.DeploymentType, relatedEndpointIds, payload.StackFileContent)
		},
	)
}

package edgestacks

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer/api/filesystem"
)

const nomadJobFileDefaultName = "nomad-job.hcl"

// @id EdgeStackCreate
// @summary Create an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param method query string true "Creation Method" Enums(file,string,repository)
// @param body_string body swarmStackFromFileContentPayload true "Required when using method=string"
// @param body_file body swarmStackFromFileUploadPayload true "Required when using method=file"
// @param body_repository body swarmStackFromGitRepositoryPayload true "Required when using method=repository"
// @success 200 {object} portaineree.EdgeStack
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks [post]
func (handler *Handler) edgeStackCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: method", err)
	}
	dryrun, _ := request.RetrieveBooleanQueryParameter(r, "dryrun", true)

	edgeStack, err := handler.createSwarmStack(method, dryrun, r)
	if err != nil {
		return httperror.InternalServerError("Unable to create Edge stack", err)
	}

	return response.JSON(w, edgeStack)
}

func (handler *Handler) createSwarmStack(method string, dryrun bool, r *http.Request) (*portaineree.EdgeStack, error) {

	switch method {
	case "string":
		return handler.createSwarmStackFromFileContent(r, dryrun)
	case "repository":
		return handler.createSwarmStackFromGitRepository(r, dryrun)
	case "file":
		return handler.createSwarmStackFromFileUpload(r, dryrun)
	}
	return nil, errors.New("Invalid value for query parameter: method. Value must be one of: string, repository or file")
}

type swarmStackFromFileContentPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// List of identifiers of EdgeGroups
	EdgeGroups []portaineree.EdgeGroupID `example:"1"`
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'
	// for compose stacks will use kompose to convert to kubernetes manifest for kubernetes environments(endpoints)
	// kubernetes deploytype is enabled only for kubernetes environments(endpoints)
	// nomad deploytype is enabled only for nomad environments(endpoints)
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portaineree.RegistryID
}

func (payload *swarmStackFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	if payload.EdgeGroups == nil || len(payload.EdgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	return nil
}

func (handler *Handler) createSwarmStackFromFileContent(r *http.Request, dryrun bool) (*portaineree.EdgeStack, error) {
	var payload swarmStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	err = handler.validateUniqueName(payload.Name)
	if err != nil {
		return nil, err
	}

	stack := &portaineree.EdgeStack{
		Name:           payload.Name,
		DeploymentType: payload.DeploymentType,
		CreationDate:   time.Now().Unix(),
		EdgeGroups:     payload.EdgeGroups,
		Status:         make(map[portaineree.EndpointID]portaineree.EdgeStackStatus),
		Version:        1,
		Registries:     payload.Registries,
	}

	if len(payload.Registries) == 0 && dryrun {
		err = handler.assignPrivateRegistriesToStack(stack, strings.NewReader(payload.StackFileContent))
		if err != nil {
			return nil, errors.Wrap(err, "failed to assign private registries to stack")
		}
	}

	if dryrun {
		return stack, nil
	}

	stack.ID = portaineree.EdgeStackID(handler.DataStore.EdgeStack().GetNextIdentifier())

	relationConfig, err := fetchEndpointRelationsConfig(handler.DataStore)
	if err != nil {
		return nil, fmt.Errorf("unable to find environment relations in database: %w", err)
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(stack.EdgeGroups, relationConfig.endpoints, relationConfig.endpointGroups, relationConfig.edgeGroups)
	if err != nil {
		return nil, fmt.Errorf("unable to persist environment relation in database: %w", err)
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	if stack.DeploymentType == portaineree.EdgeStackDeploymentCompose {
		stack.EntryPoint = filesystem.ComposeFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return nil, err
		}
		stack.ProjectPath = projectPath

		err = handler.convertAndStoreKubeManifestIfNeeded(stack, relatedEndpointIds)
		if err != nil {
			return nil, fmt.Errorf("Failed creating and storing kube manifest: %w", err)
		}

	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		hasDockerEndpoint, err := hasDockerEndpoint(handler.DataStore.Endpoint(), relatedEndpointIds)
		if err != nil {
			return nil, fmt.Errorf("unable to check for existence of docker environment: %w", err)
		}

		if hasDockerEndpoint {
			return nil, fmt.Errorf("edge stack with docker environment cannot be deployed with kubernetes config")
		}

		stack.ManifestPath = filesystem.ManifestFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.ManifestPath, []byte(payload.StackFileContent))
		if err != nil {
			return nil, err
		}

		stack.ProjectPath = projectPath
	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentNomad {
		stack.EntryPoint = nomadJobFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return nil, err
		}

		stack.ProjectPath = projectPath
	}

	err = handler.updateEndpointRelations(stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update environment relations: %w", err)
	}

	err = handler.DataStore.EdgeStack().Create(stack.ID, stack)
	if err != nil {
		return nil, err
	}

	err = handler.createEdgeCommands(stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update environment relations: %w", err)
	}

	return stack, nil
}

type swarmStackFromGitRepositoryPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// URL of a Git repository hosting the Stack file
	RepositoryURL string `example:"https://github.com/openfaas/faas" validate:"required"`
	// Reference name of a Git repository hosting the Stack file
	RepositoryReferenceName string `example:"refs/heads/master"`
	// Use basic authentication to clone the Git repository
	RepositoryAuthentication bool `example:"true"`
	// Username used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryPassword string `example:"myGitPassword"`
	// Path to the Stack file inside the Git repository
	FilePathInRepository string `example:"docker-compose.yml" default:"docker-compose.yml"`
	// List of identifiers of EdgeGroups
	EdgeGroups []portaineree.EdgeGroupID `example:"1"`
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'
	// for compose stacks will use kompose to convert to kubernetes manifest for kubernetes environments(endpoints)
	// kubernetes deploytype is enabled only for kubernetes environments(endpoints)
	// nomad deploytype is enabled only for nomad environments(endpoints)
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portaineree.RegistryID
}

func (payload *swarmStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && (govalidator.IsNull(payload.RepositoryUsername) || govalidator.IsNull(payload.RepositoryPassword)) {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.FilePathInRepository) {
		switch payload.DeploymentType {
		case portaineree.EdgeStackDeploymentCompose:
			payload.FilePathInRepository = filesystem.ComposeFileDefaultName
		case portaineree.EdgeStackDeploymentKubernetes:
			payload.FilePathInRepository = filesystem.ManifestFileDefaultName
		case portaineree.EdgeStackDeploymentNomad:
			payload.FilePathInRepository = nomadJobFileDefaultName
		}
	}
	if payload.EdgeGroups == nil || len(payload.EdgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	return nil
}

func (handler *Handler) createSwarmStackFromGitRepository(r *http.Request, dryrun bool) (*portaineree.EdgeStack, error) {
	var payload swarmStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	err = handler.validateUniqueName(payload.Name)
	if err != nil {
		return nil, err
	}

	stackID := handler.DataStore.EdgeStack().GetNextIdentifier()
	stack := &portaineree.EdgeStack{
		ID:             portaineree.EdgeStackID(stackID),
		Name:           payload.Name,
		CreationDate:   time.Now().Unix(),
		EdgeGroups:     payload.EdgeGroups,
		Status:         make(map[portaineree.EndpointID]portaineree.EdgeStackStatus),
		DeploymentType: payload.DeploymentType,
		Version:        1,
		Registries:     payload.Registries,
	}

	if dryrun {
		return stack, nil
	}

	stack.ID = portaineree.EdgeStackID(handler.DataStore.EdgeStack().GetNextIdentifier())
	projectPath := handler.FileService.GetEdgeStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	repositoryUsername := payload.RepositoryUsername
	repositoryPassword := payload.RepositoryPassword
	if !payload.RepositoryAuthentication {
		repositoryUsername = ""
		repositoryPassword = ""
	}

	relationConfig, err := fetchEndpointRelationsConfig(handler.DataStore)
	if err != nil {
		return nil, fmt.Errorf("failed fetching relations config: %w", err)
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(stack.EdgeGroups, relationConfig.endpoints, relationConfig.endpointGroups, relationConfig.edgeGroups)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve related environment: %w", err)
	}

	err = handler.GitService.CloneRepository(projectPath, payload.RepositoryURL, payload.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		return nil, err
	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentCompose {
		stack.EntryPoint = payload.FilePathInRepository

		err = handler.convertAndStoreKubeManifestIfNeeded(stack, relatedEndpointIds)
		if err != nil {
			return nil, fmt.Errorf("Failed creating and storing kube manifest: %w", err)
		}
	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		stack.ManifestPath = payload.FilePathInRepository
	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentNomad {
		stack.EntryPoint = payload.FilePathInRepository
	}

	err = handler.updateEndpointRelations(stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update environment relations: %w", err)
	}

	err = handler.DataStore.EdgeStack().Create(stack.ID, stack)
	if err != nil {
		return nil, err
	}

	err = handler.createEdgeCommands(stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update environment relations: %w", err)
	}

	return stack, nil
}

type swarmStackFromFileUploadPayload struct {
	Name             string
	StackFileContent []byte
	EdgeGroups       []portaineree.EdgeGroupID
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'
	// for compose stacks will use kompose to convert to kubernetes manifest for kubernetes environments(endpoints)
	// kubernetes deploytype is enabled only for kubernetes environments(endpoints)
	// nomad deploytype is enabled only for nomad environments(endpoints)
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	Registries     []portaineree.RegistryID
}

func (payload *swarmStackFromFileUploadPayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("Invalid stack name")
	}
	payload.Name = name

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var edgeGroups []portaineree.EdgeGroupID
	err = request.RetrieveMultiPartFormJSONValue(r, "EdgeGroups", &edgeGroups, false)
	if err != nil || len(edgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	payload.EdgeGroups = edgeGroups

	deploymentType, err := request.RetrieveNumericMultiPartFormValue(r, "DeploymentType", true)
	if err != nil {
		return errors.New("Invalid deployment type")
	}
	payload.DeploymentType = portaineree.EdgeStackDeploymentType(deploymentType)

	var registries []portaineree.RegistryID
	request.RetrieveMultiPartFormJSONValue(r, "Registries", &registries, false)
	if err != nil {
		return errors.New("Invalid registry type")
	}
	payload.Registries = registries

	return nil
}

func (handler *Handler) createSwarmStackFromFileUpload(r *http.Request, dryrun bool) (*portaineree.EdgeStack, error) {
	payload := &swarmStackFromFileUploadPayload{}
	err := payload.Validate(r)
	if err != nil {
		return nil, err
	}

	err = handler.validateUniqueName(payload.Name)
	if err != nil {
		return nil, err
	}

	stack := &portaineree.EdgeStack{
		Name:           payload.Name,
		DeploymentType: payload.DeploymentType,
		CreationDate:   time.Now().Unix(),
		EdgeGroups:     payload.EdgeGroups,
		Status:         make(map[portaineree.EndpointID]portaineree.EdgeStackStatus),
		Version:        1,
		Registries:     payload.Registries,
	}

	if len(payload.Registries) == 0 && dryrun {
		err = handler.assignPrivateRegistriesToStack(stack, bytes.NewReader(payload.StackFileContent))
		if err != nil {
			return nil, errors.Wrap(err, "failed to assign private registries to stack")
		}
	}

	if dryrun {
		return stack, nil
	}

	stack.ID = portaineree.EdgeStackID(handler.DataStore.EdgeStack().GetNextIdentifier())

	relationConfig, err := fetchEndpointRelationsConfig(handler.DataStore)
	if err != nil {
		return nil, fmt.Errorf("failed fetching relations config: %w", err)
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(stack.EdgeGroups, relationConfig.endpoints, relationConfig.endpointGroups, relationConfig.edgeGroups)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve related environment: %w", err)
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	if stack.DeploymentType == portaineree.EdgeStackDeploymentCompose {
		stack.EntryPoint = filesystem.ComposeFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return nil, err
		}
		stack.ProjectPath = projectPath

		err = handler.convertAndStoreKubeManifestIfNeeded(stack, relatedEndpointIds)
		if err != nil {
			return nil, fmt.Errorf("Failed creating and storing kube manifest: %w", err)
		}

	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		stack.ManifestPath = filesystem.ManifestFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.ManifestPath, []byte(payload.StackFileContent))
		if err != nil {
			return nil, err
		}
		stack.ProjectPath = projectPath
	}

	if stack.DeploymentType == portaineree.EdgeStackDeploymentNomad {
		stack.EntryPoint = nomadJobFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return nil, err
		}

		stack.ProjectPath = projectPath
	}

	err = handler.updateEndpointRelations(stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update environment relations: %w", err)
	}

	err = handler.DataStore.EdgeStack().Create(stack.ID, stack)
	if err != nil {
		return nil, err
	}

	err = handler.createEdgeCommands(stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update environment relations: %w", err)
	}

	return stack, nil
}

func (handler *Handler) validateUniqueName(name string) error {
	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return err
	}

	for _, stack := range edgeStacks {
		if strings.EqualFold(stack.Name, name) {
			return errors.New("Edge stack name must be unique")
		}
	}
	return nil
}

// updateEndpointRelations adds a relation between the Edge Stack to the related environments(endpoints)
func (handler *Handler) updateEndpointRelations(edgeStackID portaineree.EdgeStackID, relatedEndpointIds []portaineree.EndpointID) error {
	for _, endpointID := range relatedEndpointIds {
		relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return fmt.Errorf("unable to find environment relation in database: %w", err)
		}

		relation.EdgeStacks[edgeStackID] = true

		err = handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return fmt.Errorf("unable to persist environment relation in database: %w", err)
		}
	}

	return nil
}

func (handler *Handler) createEdgeCommands(edgeStackID portaineree.EdgeStackID, relatedEndpointIds []portaineree.EndpointID) error {
	for _, endpointID := range relatedEndpointIds {
		endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
		if err != nil {
			return err
		}

		err = handler.edgeService.AddStackCommand(endpoint, edgeStackID)
		if err != nil {
			return err
		}
	}

	return nil
}

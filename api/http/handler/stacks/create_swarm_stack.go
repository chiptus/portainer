package stacks

import (
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
)

type swarmStackFromFileContentPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Swarm cluster identifier
	SwarmID string `example:"jpofkc0i9uo9wtx1zesuk649w" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// Whether the stack is from a app template
	FromAppTemplate bool `example:"false"`
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
}

func (payload *swarmStackFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.SwarmID) {
		return errors.New("Invalid Swarm ID")
	}
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

func (handler *Handler) createSwarmStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	var payload swarmStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, true)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to check for name collision", err}
	}
	if !isUnique {
		return stackExistsError(payload.Name)
	}

	isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
	if isUniqueError != nil {
		return isUniqueError
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Name:            payload.Name,
		Type:            portaineree.DockerSwarmStack,
		SwarmID:         payload.SwarmID,
		EndpointID:      endpoint.ID,
		EntryPoint:      filesystem.ComposeFileDefaultName,
		Env:             payload.Env,
		Status:          portaineree.StackStatusActive,
		CreationDate:    time.Now().Unix(),
		FromAppTemplate: payload.FromAppTemplate,
		Webhook:         payload.Webhook,
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Compose file on disk", err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, false, true)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, err.Error(), err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the stack inside the database", err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type swarmStackFromGitRepositoryPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Swarm cluster identifier
	SwarmID string `example:"jpofkc0i9uo9wtx1zesuk649w" validate:"required"`
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair

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
	ComposeFile string `example:"docker-compose.yml" default:"docker-compose.yml"`
	// Applicable when deploying with multiple stack files
	AdditionalFiles []string `example:"[nz.compose.yml, uat.compose.yml]"`
	// Optional auto update configuration
	AutoUpdate *portaineree.StackAutoUpdate
	// Whether the stack is from a app template
	FromAppTemplate bool `example:"false"`
}

func (payload *swarmStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.SwarmID) {
		return errors.New("Invalid Swarm ID")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if govalidator.IsNull(payload.RepositoryReferenceName) {
		payload.RepositoryReferenceName = defaultGitReferenceName
	}
	if payload.RepositoryAuthentication && govalidator.IsNull(payload.RepositoryPassword) {
		return errors.New("Invalid repository credentials. Password must be specified when authentication is enabled")
	}
	if err := validateStackAutoUpdate(payload.AutoUpdate); err != nil {
		return err
	}
	return nil
}

func (handler *Handler) createSwarmStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	var payload swarmStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, true)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for name collision", Err: err}
	}
	if !isUnique {
		return stackExistsError(payload.Name)
	}

	//make sure the webhook ID is unique
	if payload.AutoUpdate != nil {
		isUniqueError := handler.checkUniqueWebhookID(payload.AutoUpdate.Webhook)
		if isUniqueError != nil {
			return isUniqueError
		}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Name:            payload.Name,
		Type:            portaineree.DockerSwarmStack,
		SwarmID:         payload.SwarmID,
		EndpointID:      endpoint.ID,
		EntryPoint:      payload.ComposeFile,
		AdditionalFiles: payload.AdditionalFiles,
		AutoUpdate:      payload.AutoUpdate,
		GitConfig: &gittypes.RepoConfig{
			URL:            payload.RepositoryURL,
			ReferenceName:  payload.RepositoryReferenceName,
			ConfigFilePath: payload.ComposeFile,
		},
		FromAppTemplate: payload.FromAppTemplate,
		Env:             payload.Env,
		Status:          portaineree.StackStatusActive,
		CreationDate:    time.Now().Unix(),
	}

	if payload.RepositoryAuthentication {
		stack.GitConfig.Authentication = &gittypes.GitAuthentication{
			Username: payload.RepositoryUsername,
			Password: payload.RepositoryPassword,
		}
	}

	projectPath := handler.FileService.GetStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	err = handler.clone(projectPath, payload.RepositoryURL, payload.RepositoryReferenceName, payload.RepositoryAuthentication, payload.RepositoryUsername, payload.RepositoryPassword)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to clone git repository", Err: err}
	}

	commitID, err := handler.latestCommitID(payload.RepositoryURL, payload.RepositoryReferenceName, payload.RepositoryAuthentication, payload.RepositoryUsername, payload.RepositoryPassword)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch git repository id", Err: err}
	}
	stack.GitConfig.ConfigHash = commitID

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, false, true)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
		jobID, e := startAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService)
		if e != nil {
			return e
		}

		stack.AutoUpdate.JobID = jobID
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack inside the database", Err: err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type swarmStackFromFileUploadPayload struct {
	Name             string
	SwarmID          string
	StackFileContent []byte
	Env              []portaineree.Pair
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
}

func (payload *swarmStackFromFileUploadPayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("Invalid stack name")
	}
	payload.Name = name

	swarmID, err := request.RetrieveMultiPartFormValue(r, "SwarmID", false)
	if err != nil {
		return errors.New("Invalid Swarm ID")
	}
	payload.SwarmID = swarmID

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var env []portaineree.Pair
	err = request.RetrieveMultiPartFormJSONValue(r, "Env", &env, true)
	if err != nil {
		return errors.New("Invalid Env parameter")
	}
	payload.Env = env
	webhook, err := request.RetrieveMultiPartFormValue(r, "Webhook", true)
	if err == nil {
		payload.Webhook = webhook
	}
	return nil
}

func (handler *Handler) createSwarmStackFromFileUpload(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	payload := &swarmStackFromFileUploadPayload{}
	err := payload.Validate(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, true)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to check for name collision", err}
	}
	if !isUnique {
		return stackExistsError(payload.Name)
	}

	isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
	if isUniqueError != nil {
		return isUniqueError
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:           portaineree.StackID(stackID),
		Name:         payload.Name,
		Type:         portaineree.DockerSwarmStack,
		SwarmID:      payload.SwarmID,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		Env:          payload.Env,
		Status:       portaineree.StackStatusActive,
		CreationDate: time.Now().Unix(),
		Webhook:      payload.Webhook,
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Compose file on disk", err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, false, true)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, err.Error(), err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the stack inside the database", err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type swarmStackDeploymentConfig struct {
	stack      *portaineree.Stack
	endpoint   *portaineree.Endpoint
	registries []portaineree.Registry
	prune      bool
	isAdmin    bool
	user       *portaineree.User
	pullImage  bool
}

func (handler *Handler) createSwarmDeployConfig(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint, prune bool, pullImage bool) (*swarmStackDeploymentConfig, *httperror.HandlerError) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to load user information from the database", err}
	}

	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve registries from the database", err}
	}

	filteredRegistries := security.FilterRegistries(registries, user, securityContext.UserMemberships, endpoint.ID)

	config := &swarmStackDeploymentConfig{
		stack:      stack,
		endpoint:   endpoint,
		registries: filteredRegistries,
		prune:      prune,
		isAdmin:    securityContext.IsAdmin,
		user:       user,
		pullImage:  pullImage,
	}

	return config, nil
}

func (handler *Handler) deploySwarmStack(config *swarmStackDeploymentConfig) error {
	isAdminOrEndpointAdmin, err := handler.userIsAdminOrEndpointAdmin(config.user, config.endpoint.ID)
	if err != nil {
		return errors.Wrap(err, "failed to validate user admin privileges")
	}

	settings := &config.endpoint.SecuritySettings

	if !settings.AllowBindMountsForRegularUsers && !isAdminOrEndpointAdmin {
		for _, file := range append([]string{config.stack.EntryPoint}, config.stack.AdditionalFiles...) {
			stackContent, err := handler.FileService.GetFileContent(config.stack.ProjectPath, file)
			if err != nil {
				return errors.WithMessage(err, "failed to get stack file content")
			}

			err = handler.isValidStackFile(stackContent, settings)
			if err != nil {
				return errors.WithMessage(err, "swarm stack file content validation failed")
			}
		}
	}

	return handler.StackDeployer.DeploySwarmStack(config.stack, config.endpoint, config.registries, config.prune, config.pullImage)
}

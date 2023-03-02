package stacks

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/git/update"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/stackbuilders"
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

func createStackPayloadFromSwarmFileContentPayload(name string, swarmID string, fileContent string, env []portaineree.Pair, fromAppTemplate bool, webhook string) stackbuilders.StackPayload {
	return stackbuilders.StackPayload{
		Name:             name,
		SwarmID:          swarmID,
		StackFileContent: fileContent,
		Env:              env,
		FromAppTemplate:  fromAppTemplate,
		Webhook:          webhook,
	}
}

func (handler *Handler) createSwarmStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	var payload swarmStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, true)

	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}
	if !isUnique {
		return stackExistsError(payload.Name)
	}

	isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
	if isUniqueError != nil {
		return isUniqueError
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stackPayload := createStackPayloadFromSwarmFileContentPayload(payload.Name, payload.SwarmID, payload.StackFileContent, payload.Env, payload.FromAppTemplate, payload.Webhook)

	swarmStackBuilder := stackbuilders.CreateSwarmStackFileContentBuilder(securityContext,
		handler.DataStore,
		handler.FileService,
		handler.StackDeployer)

	stackBuilderDirector := stackbuilders.NewStackBuilderDirector(swarmStackBuilder)
	stack, httpErr := stackBuilderDirector.Build(&stackPayload, endpoint, userID)
	if httpErr != nil {
		return httpErr
	}

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
	// Username used in basic authentication. Required when RepositoryAuthentication is true
	// and RepositoryGitCredentialID is 0
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true
	// and RepositoryGitCredentialID is 0
	RepositoryPassword string `example:"myGitPassword"`
	// GitCredentialID used to identify the bound git credential. Required when RepositoryAuthentication
	// is true and RepositoryUsername/RepositoryPassword are not provided
	RepositoryGitCredentialID int `example:"0"`
	// Path to the Stack file inside the Git repository
	ComposeFile string `example:"docker-compose.yml" default:"docker-compose.yml"`
	// Applicable when deploying with multiple stack files
	AdditionalFiles []string `example:"[nz.compose.yml, uat.compose.yml]"`
	// Optional auto update configuration
	AutoUpdate *portaineree.AutoUpdateSettings
	// Whether the stack is from a app template
	FromAppTemplate bool `example:"false"`
	// Whether the stack suppors relative path volume
	SupportRelativePath bool `example:"false"`
	// Network filesystem path
	FilesystemPath string `example:"/tmp"`
}

func createStackPayloadFromSwarmGitPayload(name, swarmID, repoUrl, repoReference, repoUsername, repoPassword string, repoGitCredentialID int, repoAuthentication bool, composeFile string, additionalFiles []string, autoUpdate *portaineree.AutoUpdateSettings, env []portaineree.Pair, fromAppTemplate, supportRelativePath bool, filesystemPath string) stackbuilders.StackPayload {
	return stackbuilders.StackPayload{
		Name:    name,
		SwarmID: swarmID,
		RepositoryConfigPayload: stackbuilders.RepositoryConfigPayload{
			URL:             repoUrl,
			ReferenceName:   repoReference,
			Authentication:  repoAuthentication,
			Username:        repoUsername,
			Password:        repoPassword,
			GitCredentialID: repoGitCredentialID,
		},
		ComposeFile:         composeFile,
		AdditionalFiles:     additionalFiles,
		AutoUpdate:          autoUpdate,
		Env:                 env,
		FromAppTemplate:     fromAppTemplate,
		SupportRelativePath: supportRelativePath,
		FilesystemPath:      filesystemPath,
	}
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
	if payload.RepositoryAuthentication && govalidator.IsNull(payload.RepositoryPassword) && payload.RepositoryGitCredentialID == 0 {
		return errors.New("Invalid repository credentials. Password must be specified when authentication is enabled")
	}
	if err := update.ValidateAutoUpdateSettings(payload.AutoUpdate); err != nil {
		return err
	}
	return nil
}

func (handler *Handler) createSwarmStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	var payload swarmStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, true)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
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

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stackPayload := createStackPayloadFromSwarmGitPayload(payload.Name,
		payload.SwarmID,
		payload.RepositoryURL,
		payload.RepositoryReferenceName,
		payload.RepositoryUsername,
		payload.RepositoryPassword,
		payload.RepositoryGitCredentialID,
		payload.RepositoryAuthentication,
		payload.ComposeFile,
		payload.AdditionalFiles,
		payload.AutoUpdate,
		payload.Env,
		payload.FromAppTemplate,
		payload.SupportRelativePath,
		payload.FilesystemPath)

	swarmStackBuilder := stackbuilders.CreateSwarmStackGitBuilder(securityContext,
		handler.userActivityService,
		handler.DataStore,
		handler.FileService,
		handler.GitService,
		handler.Scheduler,
		handler.StackDeployer)

	stackBuilderDirector := stackbuilders.NewStackBuilderDirector(swarmStackBuilder)
	stack, httpErr := stackBuilderDirector.Build(&stackPayload, endpoint, userID)
	if httpErr != nil {
		return httpErr
	}

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

func createStackPayloadFromSwarmFileUploadPayload(name, swarmID string, fileContentBytes []byte, env []portaineree.Pair, webhook string) stackbuilders.StackPayload {
	return stackbuilders.StackPayload{
		Name:                  name,
		SwarmID:               swarmID,
		StackFileContentBytes: fileContentBytes,
		Env:                   env,
		Webhook:               webhook,
	}
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
	var payload swarmStackFromFileUploadPayload
	err := payload.Validate(r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, true)

	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}
	if !isUnique {
		return stackExistsError(payload.Name)
	}

	isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
	if isUniqueError != nil {
		return isUniqueError
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stackPayload := createStackPayloadFromSwarmFileUploadPayload(payload.Name, payload.SwarmID, payload.StackFileContent, payload.Env, payload.Webhook)

	swarmStackBuilder := stackbuilders.CreateSwarmStackFileUploadBuilder(securityContext,
		handler.DataStore,
		handler.FileService,
		handler.StackDeployer)

	stackBuilderDirector := stackbuilders.NewStackBuilderDirector(swarmStackBuilder)
	stack, httpErr := stackBuilderDirector.Build(&stackPayload, endpoint, userID)
	if httpErr != nil {
		return httpErr
	}

	return handler.decorateStackResponse(w, stack, userID)
}

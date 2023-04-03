package stacks

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/git/update"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackbuilders"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type composeStackFromFileContentPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// Whether the stack is from a app template
	FromAppTemplate bool `example:"false"`
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
}

func (payload *composeStackFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}

	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

func createStackPayloadFromComposeFileContentPayload(name string, fileContent string, env []portaineree.Pair, fromAppTemplate bool, webhook string) stackbuilders.StackPayload {
	return stackbuilders.StackPayload{
		Name:             name,
		StackFileContent: fileContent,
		Env:              env,
		FromAppTemplate:  fromAppTemplate,
		Webhook:          webhook,
	}
}

func (handler *Handler) checkAndCleanStackDupFromSwarm(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID, stack *portaineree.Stack) error {
	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
	if err != nil {
		return err
	}

	// stop scheduler updates of the stack before removal
	if stack.AutoUpdate != nil {
		deployments.StopAutoupdate(stack.ID, stack.AutoUpdate.JobID, handler.Scheduler)
	}

	err = handler.DataStore.Stack().DeleteStack(stack.ID)
	if err != nil {
		return err
	}

	if resourceControl != nil {
		err = handler.DataStore.ResourceControl().DeleteResourceControl(resourceControl.ID)
		if err != nil {
			log.Error().
				Str("stack", fmt.Sprintf("%+v", stack)).
				Msg("unable to remove the associated resource control from the database for stack")
		}
	}

	if exists, _ := handler.FileService.FileExists(stack.ProjectPath); exists {
		err = handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Warn().
				Str("stack", fmt.Sprintf("%+v", stack)).
				Msg("unable to remove stack files from disk for stack")
		}
	}

	return nil
}

func (handler *Handler) createComposeStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	var payload composeStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	payload.Name = handler.ComposeStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, false)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}

	if !isUnique {
		stacks, err := handler.DataStore.Stack().StacksByName(payload.Name)
		if err != nil {
			return stackExistsError(payload.Name)
		}
		for _, stack := range stacks {
			if stack.Type != portaineree.DockerComposeStack && stack.EndpointID == endpoint.ID {
				err := handler.checkAndCleanStackDupFromSwarm(w, r, endpoint, userID, &stack)
				if err != nil {
					return httperror.BadRequest("Invalid request payload", err)
				}
			} else {
				return stackExistsError(payload.Name)
			}
		}
	}

	isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
	if isUniqueError != nil {
		return isUniqueError
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stackPayload := createStackPayloadFromComposeFileContentPayload(payload.Name, payload.StackFileContent, payload.Env, payload.FromAppTemplate, payload.Webhook)

	composeStackBuilder := stackbuilders.CreateComposeStackFileContentBuilder(securityContext,
		handler.DataStore,
		handler.FileService,
		handler.StackDeployer)

	stackBuilderDirector := stackbuilders.NewStackBuilderDirector(composeStackBuilder)
	stack, httpErr := stackBuilderDirector.Build(&stackPayload, endpoint, userID)
	if httpErr != nil {
		return httpErr
	}

	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackFromGitRepositoryPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
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
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// Whether the stack is from a app template
	FromAppTemplate bool `example:"false"`
	// Whether the stack supports relative path volume
	SupportRelativePath bool `example:"false"`
	// Local filesystem path
	FilesystemPath string `example:"/tmp"`
	// TLSSkipVerify skips SSL verification when cloning the Git repository
	TLSSkipVerify bool `example:"false"`
}

func createStackPayloadFromComposeGitPayload(name, repoUrl, repoReference, repoUsername, repoPassword string, repoGitCredentialID int, repoAuthentication bool, composeFile string, additionalFiles []string, autoUpdate *portaineree.AutoUpdateSettings, env []portaineree.Pair, fromAppTemplate, supportRelativePath bool, filesystemPath string, repoTLSSkipVerify bool) stackbuilders.StackPayload {
	return stackbuilders.StackPayload{
		Name: name,
		RepositoryConfigPayload: stackbuilders.RepositoryConfigPayload{
			URL:             repoUrl,
			ReferenceName:   repoReference,
			Authentication:  repoAuthentication,
			Username:        repoUsername,
			Password:        repoPassword,
			GitCredentialID: repoGitCredentialID,
			TLSSkipVerify:   repoTLSSkipVerify,
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

func (payload *composeStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
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

func (handler *Handler) createComposeStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	var payload composeStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	payload.Name = handler.ComposeStackManager.NormalizeStackName(payload.Name)
	if payload.ComposeFile == "" {
		payload.ComposeFile = filesystem.ComposeFileDefaultName
	}

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, false)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}

	if !isUnique {
		stacks, err := handler.DataStore.Stack().StacksByName(payload.Name)
		if err != nil {
			return stackExistsError(payload.Name)
		}
		for _, stack := range stacks {
			if stack.Type != portaineree.DockerComposeStack && stack.EndpointID == endpoint.ID {
				err := handler.checkAndCleanStackDupFromSwarm(w, r, endpoint, userID, &stack)
				if err != nil {
					return httperror.BadRequest("Invalid request payload", err)
				}
			} else {
				return stackExistsError(payload.Name)
			}
		}
	}

	//make sure the webhook ID is unique
	if payload.AutoUpdate != nil {
		err := handler.checkUniqueWebhookID(payload.AutoUpdate.Webhook)
		if err != nil {
			return err
		}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stackPayload := createStackPayloadFromComposeGitPayload(payload.Name,
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
		payload.FilesystemPath,
		payload.TLSSkipVerify,
	)

	composeStackBuilder := stackbuilders.CreateComposeStackGitBuilder(securityContext,
		handler.userActivityService,
		handler.DataStore,
		handler.FileService,
		handler.GitService,
		handler.Scheduler,
		handler.StackDeployer)

	stackBuilderDirector := stackbuilders.NewStackBuilderDirector(composeStackBuilder)
	stack, httpErr := stackBuilderDirector.Build(&stackPayload, endpoint, userID)
	if httpErr != nil {
		return httpErr
	}

	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackFromFileUploadPayload struct {
	Name             string
	StackFileContent []byte
	Env              []portaineree.Pair
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
}

func createStackPayloadFromComposeFileUploadPayload(name string, fileContentBytes []byte, env []portaineree.Pair, webhook string) stackbuilders.StackPayload {
	return stackbuilders.StackPayload{
		Name:                  name,
		StackFileContentBytes: fileContentBytes,
		Env:                   env,
		Webhook:               webhook,
	}
}

func decodeRequestForm(r *http.Request) (*composeStackFromFileUploadPayload, error) {
	payload := &composeStackFromFileUploadPayload{}
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return nil, errors.New("Invalid stack name")
	}
	payload.Name = name

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return nil, errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var env []portaineree.Pair
	err = request.RetrieveMultiPartFormJSONValue(r, "Env", &env, true)
	if err != nil {
		return nil, errors.New("Invalid Env parameter")
	}
	payload.Env = env
	webhook, err := request.RetrieveMultiPartFormValue(r, "Webhook", true)
	if err == nil {
		payload.Webhook = webhook
	}
	return payload, nil
}

func (handler *Handler) createComposeStackFromFileUpload(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID) *httperror.HandlerError {
	payload, err := decodeRequestForm(r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	payload.Name = handler.ComposeStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, payload.Name, 0, false)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}

	if !isUnique {
		stacks, err := handler.DataStore.Stack().StacksByName(payload.Name)
		if err != nil {
			return stackExistsError(payload.Name)
		}
		for _, stack := range stacks {
			if stack.Type != portaineree.DockerComposeStack && stack.EndpointID == endpoint.ID {
				err := handler.checkAndCleanStackDupFromSwarm(w, r, endpoint, userID, &stack)
				if err != nil {
					return httperror.BadRequest("Invalid request payload", err)
				}
			} else {
				return stackExistsError(payload.Name)
			}
		}
	}

	isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
	if isUniqueError != nil {
		return isUniqueError
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stackPayload := createStackPayloadFromComposeFileUploadPayload(payload.Name, payload.StackFileContent, payload.Env, payload.Webhook)

	composeStackBuilder := stackbuilders.CreateComposeStackFileUploadBuilder(securityContext,
		handler.DataStore,
		handler.FileService,
		handler.StackDeployer)

	stackBuilderDirector := stackbuilders.NewStackBuilderDirector(composeStackBuilder)
	stack, httpErr := stackBuilderDirector.Build(&stackPayload, endpoint, userID)
	if httpErr != nil {
		return httpErr
	}

	return handler.decorateStackResponse(w, stack, userID)
}

package stacks

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
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
func (handler *Handler) checkAndCleanStackDupFromSwarm(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, userID portaineree.UserID, stack *portaineree.Stack) error {
	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
	if err != nil {
		return err
	}
	// stop scheduler updates of the stack before removal
	if stack.AutoUpdate != nil {
		stopAutoupdate(stack.ID, stack.AutoUpdate.JobID, *handler.Scheduler)
	}

	err = handler.DataStore.Stack().DeleteStack(stack.ID)
	if err != nil {
		return err
	}

	if resourceControl != nil {
		err = handler.DataStore.ResourceControl().DeleteResourceControl(resourceControl.ID)
		if err != nil {
			log.Printf("[ERROR] [Stack] Unable to remove the associated resource control from the database for stack: [%+v].", stack)
		}
	}

	if exists, _ := handler.FileService.FileExists(stack.ProjectPath); exists {
		err = handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Printf("Unable to remove stack files from disk for stack: [%+v].", stack)
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

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Name:            payload.Name,
		Type:            portaineree.DockerComposeStack,
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
		return httperror.InternalServerError("Unable to persist Compose file on disk", err)
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint, false)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config, false)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().Create(stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack inside the database", err)
	}

	doCleanUp = false
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
	AutoUpdate *portaineree.StackAutoUpdate
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// Whether the stack is from a app template
	FromAppTemplate bool `example:"false"`
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
	if err := validateStackAutoUpdate(payload.AutoUpdate); err != nil {
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

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Name:            payload.Name,
		Type:            portaineree.DockerComposeStack,
		EndpointID:      endpoint.ID,
		EntryPoint:      payload.ComposeFile,
		AdditionalFiles: payload.AdditionalFiles,
		AutoUpdate:      payload.AutoUpdate,
		Env:             payload.Env,
		FromAppTemplate: payload.FromAppTemplate,
		GitConfig: &gittypes.RepoConfig{
			URL:            payload.RepositoryURL,
			ReferenceName:  payload.RepositoryReferenceName,
			ConfigFilePath: payload.ComposeFile,
		},
		Status:       portaineree.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	repositoryUsername := ""
	repositoryPassword := ""
	repositoryGitCredentialID := 0
	if payload.RepositoryAuthentication {
		if payload.RepositoryGitCredentialID != 0 {
			credential, err := handler.DataStore.GitCredential().GetGitCredential(portaineree.GitCredentialID(payload.RepositoryGitCredentialID))
			if err != nil {
				return httperror.InternalServerError("git credential not found", err)
			}

			// When creating the stack with an existing git credential, the git credential must be owned by the calling user
			if credential.UserID != userID {
				return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "couldn't add the git credential of another user", Err: httperrors.ErrUnauthorized}
			}

			repositoryUsername = credential.Username
			repositoryPassword = credential.Password
			repositoryGitCredentialID = payload.RepositoryGitCredentialID
		}

		if payload.RepositoryPassword != "" {
			repositoryUsername = payload.RepositoryUsername
			repositoryPassword = payload.RepositoryPassword
			repositoryGitCredentialID = 0
		}

		stack.GitConfig.Authentication = &gittypes.GitAuthentication{
			GitCredentialID: repositoryGitCredentialID,
			Username:        repositoryUsername,
			Password:        repositoryPassword,
		}
	}

	projectPath := handler.FileService.GetStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	err = handler.clone(projectPath, payload.RepositoryURL, payload.RepositoryReferenceName, payload.RepositoryAuthentication, repositoryUsername, repositoryPassword)
	if err != nil {
		return httperror.InternalServerError("Unable to clone git repository", err)
	}

	commitID, err := handler.latestCommitID(payload.RepositoryURL, payload.RepositoryReferenceName, payload.RepositoryAuthentication, repositoryUsername, repositoryPassword)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch git repository id", err)
	}
	stack.GitConfig.ConfigHash = commitID

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint, false)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config, false)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
		jobID, e := startAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService)
		if e != nil {
			return e
		}

		stack.AutoUpdate.JobID = jobID
	}

	stack.CreatedBy = config.user.Username
	err = handler.DataStore.Stack().Create(stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack inside the database", err)
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackFromFileUploadPayload struct {
	Name             string
	StackFileContent []byte
	Env              []portaineree.Pair
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
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

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:           portaineree.StackID(stackID),
		Name:         payload.Name,
		Type:         portaineree.DockerComposeStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		Env:          payload.Env,
		Status:       portaineree.StackStatusActive,
		CreationDate: time.Now().Unix(),
		Webhook:      payload.Webhook,
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, payload.StackFileContent)
	if err != nil {
		return httperror.InternalServerError("Unable to persist Compose file on disk", err)
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint, false)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config, false)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().Create(stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack inside the database", err)
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackDeploymentConfig struct {
	stack          *portaineree.Stack
	endpoint       *portaineree.Endpoint
	registries     []portaineree.Registry
	isAdmin        bool
	user           *portaineree.User
	forcePullImage bool
}

func (handler *Handler) createComposeDeployConfig(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint, forcePullImage bool) (*composeStackDeploymentConfig, *httperror.HandlerError) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to load user information from the database", err}
	}

	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve registries from the database", err)
	}

	filteredRegistries := security.FilterRegistries(registries, user, securityContext.UserMemberships, endpoint.ID)

	config := &composeStackDeploymentConfig{
		stack:          stack,
		endpoint:       endpoint,
		registries:     filteredRegistries,
		isAdmin:        securityContext.IsAdmin,
		user:           user,
		forcePullImage: forcePullImage,
	}

	return config, nil
}

// TODO: libcompose uses credentials store into a config.json file to pull images from
// private registries. Right now the only solution is to re-use the embedded Docker binary
// to login/logout, which will generate the required data in the config.json file and then
// clean it. Hence the use of the mutex.
// We should contribute to libcompose to support authentication without using the config.json file.
func (handler *Handler) deployComposeStack(config *composeStackDeploymentConfig, forceCreate bool) error {
	isAdminOrEndpointAdmin, err := handler.userIsAdminOrEndpointAdmin(config.user, config.endpoint.ID)
	if err != nil {
		return errors.Wrap(err, "failed to check user priviliges deploying a stack")
	}

	securitySettings := &config.endpoint.SecuritySettings

	if (!securitySettings.AllowBindMountsForRegularUsers ||
		!securitySettings.AllowPrivilegedModeForRegularUsers ||
		!securitySettings.AllowHostNamespaceForRegularUsers ||
		!securitySettings.AllowDeviceMappingForRegularUsers ||
		!securitySettings.AllowSysctlSettingForRegularUsers ||
		!securitySettings.AllowContainerCapabilitiesForRegularUsers) &&
		!isAdminOrEndpointAdmin {

		for _, file := range append([]string{config.stack.EntryPoint}, config.stack.AdditionalFiles...) {
			stackContent, err := handler.FileService.GetFileContent(config.stack.ProjectPath, file)
			if err != nil {
				return errors.Wrapf(err, "failed to get stack file content `%q`", file)
			}

			err = handler.isValidStackFile(stackContent, securitySettings)
			if err != nil {
				return errors.Wrap(err, "compose file is invalid")
			}
		}
	}

	return handler.StackDeployer.DeployComposeStack(config.stack, config.endpoint, config.registries, config.forcePullImage, forceCreate)
}

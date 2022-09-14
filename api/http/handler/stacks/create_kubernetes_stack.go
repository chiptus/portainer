package stacks

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
)

type kubernetesStringDeploymentPayload struct {
	StackName        string
	ComposeFormat    bool
	Namespace        string
	StackFileContent string
}

type kubernetesGitDeploymentPayload struct {
	StackName                string
	ComposeFormat            bool
	Namespace                string
	RepositoryURL            string
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	ManifestFile             string
	AdditionalFiles          []string
	AutoUpdate               *portaineree.StackAutoUpdate
}

type kubernetesManifestURLDeploymentPayload struct {
	StackName     string
	Namespace     string
	ComposeFormat bool
	ManifestURL   string
}

func (payload *kubernetesStringDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	if govalidator.IsNull(payload.StackName) {
		return errors.New("Invalid stack name")
	}
	return nil
}

func (payload *kubernetesGitDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && govalidator.IsNull(payload.RepositoryPassword) {
		return errors.New("Invalid repository credentials. Password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.ManifestFile) {
		return errors.New("Invalid manifest file in repository")
	}
	if err := validateStackAutoUpdate(payload.AutoUpdate); err != nil {
		return err
	}
	if govalidator.IsNull(payload.StackName) {
		return errors.New("Invalid stack name")
	}
	return nil
}

func (payload *kubernetesManifestURLDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.ManifestURL) || !govalidator.IsURL(payload.ManifestURL) {
		return errors.New("Invalid manifest URL")
	}
	if govalidator.IsNull(payload.StackName) {
		return errors.New("Invalid stack name")
	}
	return nil
}

type createKubernetesStackResponse struct {
	Output string `json:"Output"`
}

func (handler *Handler) createKubernetesStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return httperror.BadRequest("Environment type does not match", errors.New("Environment type does not match"))
	}

	var payload kubernetesStringDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	isUnique, err := handler.checkUniqueStackNameInKubernetes(endpoint, payload.StackName, 0, payload.Namespace)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}
	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' already exists", payload.StackName), Err: errStackAlreadyExists}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("unable to retrieve user details from authentication token", err)
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Type:            portaineree.KubernetesStack,
		EndpointID:      endpoint.ID,
		EntryPoint:      filesystem.ManifestFileDefaultName,
		Name:            payload.StackName,
		Namespace:       payload.Namespace,
		Status:          portaineree.StackStatusActive,
		CreationDate:    time.Now().Unix(),
		CreatedBy:       tokenData.Username,
		IsComposeFormat: payload.ComposeFormat,
	}

	stackFolder := strconv.Itoa(int(stack.ID))

	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		fileType := "Manifest"
		if stack.IsComposeFormat {
			fileType = "Compose"
		}
		errMsg := fmt.Sprintf("Unable to persist Kubernetes %s file on disk", fileType)
		return httperror.InternalServerError(errMsg, err)
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	output, deployError := handler.deployKubernetesStack(tokenData, endpoint, stack, k.KubeAppLabels{
		StackID:   stackID,
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "content",
	})
	if deployError != nil {
		return deployError
	}

	err = handler.DataStore.Stack().Create(stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the Kubernetes stack inside the database", err)
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}

	doCleanUp = false
	return response.JSON(w, resp)
}

func (handler *Handler) createKubernetesStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return httperror.BadRequest("Environment type does not match", errors.New("Environment type does not match"))
	}

	var payload kubernetesGitDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	isUnique, err := handler.checkUniqueStackNameInKubernetes(endpoint, payload.StackName, 0, payload.Namespace)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}
	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' already exists", payload.StackName), Err: errStackAlreadyExists}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("unable to retrieve user details from authentication token", err)
	}

	//make sure the webhook ID is unique
	if payload.AutoUpdate != nil && payload.AutoUpdate.Webhook != "" {
		isUnique, err := handler.isUniqueWebhookID(payload.AutoUpdate.Webhook)
		if err != nil {
			return httperror.InternalServerError("Unable to check for webhook ID collision", err)
		}
		if !isUnique {
			return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("Webhook ID: %s already exists", payload.AutoUpdate.Webhook), Err: errWebhookIDAlreadyExists}
		}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:         portaineree.StackID(stackID),
		Type:       portaineree.KubernetesStack,
		EndpointID: endpoint.ID,
		EntryPoint: payload.ManifestFile,
		GitConfig: &gittypes.RepoConfig{
			URL:            payload.RepositoryURL,
			ReferenceName:  payload.RepositoryReferenceName,
			ConfigFilePath: payload.ManifestFile,
		},
		Namespace:       payload.Namespace,
		Name:            payload.StackName,
		Status:          portaineree.StackStatusActive,
		CreationDate:    time.Now().Unix(),
		CreatedBy:       tokenData.Username,
		IsComposeFormat: payload.ComposeFormat,
		AutoUpdate:      payload.AutoUpdate,
		AdditionalFiles: payload.AdditionalFiles,
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

	commitID, err := handler.latestCommitID(payload.RepositoryURL, payload.RepositoryReferenceName, payload.RepositoryAuthentication, payload.RepositoryUsername, payload.RepositoryPassword)
	if err != nil {
		return httperror.InternalServerError("Unable to fetch git repository id", err)
	}
	stack.GitConfig.ConfigHash = commitID

	repositoryUsername := payload.RepositoryUsername
	repositoryPassword := payload.RepositoryPassword
	if !payload.RepositoryAuthentication {
		repositoryUsername = ""
		repositoryPassword = ""
	}

	err = handler.GitService.CloneRepository(projectPath, payload.RepositoryURL, payload.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		return httperror.InternalServerError("Failed to clone git repository", err)
	}

	output, deployError := handler.deployKubernetesStack(tokenData, endpoint, stack, k.KubeAppLabels{
		StackID:   stackID,
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "git",
	})
	if deployError != nil {
		return deployError
	}

	if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
		jobID, e := startAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService)
		if e != nil {
			return e
		}

		stack.AutoUpdate.JobID = jobID
	}

	err = handler.DataStore.Stack().Create(stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the Kubernetes stack inside the database", err)
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}

	doCleanUp = false
	return response.JSON(w, resp)
}

func (handler *Handler) createKubernetesStackFromManifestURL(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	var payload kubernetesManifestURLDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	isUnique, err := handler.checkUniqueStackNameInKubernetes(endpoint, payload.StackName, 0, payload.Namespace)
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}
	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' already exists", payload.StackName), Err: errStackAlreadyExists}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("unable to retrieve user details from authentication token", err)
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Type:            portaineree.KubernetesStack,
		EndpointID:      endpoint.ID,
		EntryPoint:      filesystem.ManifestFileDefaultName,
		Namespace:       payload.Namespace,
		Name:            payload.StackName,
		Status:          portaineree.StackStatusActive,
		CreationDate:    time.Now().Unix(),
		CreatedBy:       tokenData.Username,
		IsComposeFormat: payload.ComposeFormat,
	}

	var manifestContent []byte
	manifestContent, err = client.Get(payload.ManifestURL, 30)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve manifest from URL", err)
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, manifestContent)
	if err != nil {
		return httperror.InternalServerError("Unable to persist Kubernetes manifest file on disk", err)
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	output, deployError := handler.deployKubernetesStack(tokenData, endpoint, stack, k.KubeAppLabels{
		StackID:   stackID,
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "url",
	})
	if deployError != nil {
		return deployError
	}

	err = handler.DataStore.Stack().Create(stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the Kubernetes stack inside the database", err)
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}

	doCleanUp = false
	return response.JSON(w, resp)
}

func (handler *Handler) deployKubernetesStack(tokenData *portaineree.TokenData, endpoint *portaineree.Endpoint, stack *portaineree.Stack, appLabels k.KubeAppLabels) (string, *httperror.HandlerError) {
	handler.stackCreationMutex.Lock()
	defer handler.stackCreationMutex.Unlock()

	deploymentFilesInfo, cleanup, err := stackutils.CreateTempK8SDeploymentFiles(stack, handler.KubernetesDeployer, appLabels)
	if err != nil {
		return "", httperror.InternalServerError("failed to create temp kub deployment files", err)
	}

	defer cleanup()

	err = handler.checkEndpointPermission(tokenData, deploymentFilesInfo.Namespaces, endpoint)
	if err != nil {
		return "", httperror.Forbidden("user does not have permission to deploy stack", err)
	}

	output, err := handler.KubernetesDeployer.Deploy(tokenData.ID, endpoint, deploymentFilesInfo.FilePaths, stack.Namespace)
	if err != nil {
		return "", httperror.InternalServerError("failed to deploy stack", err)
	}

	return output, nil
}

func (handler *Handler) checkEndpointPermission(tokenData *portaineree.TokenData, namespaces []string, endpoint *portaineree.Endpoint) error {
	permissionDeniedErr := errors.New("Permission denied to access environment")

	if tokenData.Role == portaineree.AdministratorRole {
		return nil
	}

	// check if the user has OperationK8sApplicationsAdvancedDeploymentRW access in the environment(endpoint)
	endpointRole, err := handler.AuthorizationService.GetUserEndpointRole(int(tokenData.ID), int(endpoint.ID))
	if err != nil {
		return errors.Wrap(err, "failed to retrieve user endpoint role")
	}
	if !endpointRole.Authorizations[portaineree.OperationK8sApplicationsAdvancedDeploymentRW] {
		return permissionDeniedErr
	}

	// will skip if user can access all namespaces
	if endpointRole.Authorizations[portaineree.OperationK8sAccessAllNamespaces] {
		return nil
	}

	cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return errors.Wrap(err, "unable to create Kubernetes client")
	}

	// check if the user has RW access to the namespace
	namespaceAuthorizations, err := handler.AuthorizationService.GetNamespaceAuthorizations(int(tokenData.ID), *endpoint, cli)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve user namespace authorizations")
	}

	// if no namespace provided, either by form or by manifest, use the default namespace
	if len(namespaces) == 0 {
		namespaces = []string{"default"}
	}

	for _, namespace := range namespaces {
		if auth, ok := namespaceAuthorizations[namespace]; !ok || !auth[portaineree.OperationK8sAccessNamespaceWrite] {
			return errors.Wrap(permissionDeniedErr, "user does not have permission to access namespace")
		}
	}

	return nil
}

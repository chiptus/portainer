package stacks

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/http/client"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/internal/endpointutils"
	"github.com/portainer/portainer/api/internal/stackutils"
	k "github.com/portainer/portainer/api/kubernetes"
	consts "github.com/portainer/portainer/api/useractivity"
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
	AutoUpdate               *portainer.StackAutoUpdate
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
	if govalidator.IsNull(payload.Namespace) {
		return errors.New("Invalid namespace")
	}
	if govalidator.IsNull(payload.StackName) {
		return errors.New("Invalid stack name")
	}
	return nil
}

func (payload *kubernetesGitDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Namespace) {
		return errors.New("Invalid namespace")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && govalidator.IsNull(payload.RepositoryPassword) {
		return errors.New("Invalid repository credentials. Password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.ManifestFile) {
		return errors.New("Invalid manifest file in repository")
	}
	if govalidator.IsNull(payload.RepositoryReferenceName) {
		payload.RepositoryReferenceName = defaultGitReferenceName
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

func (handler *Handler) createKubernetesStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Environment type does not match", Err: errors.New("Environment type does not match")}
	}

	var payload kubernetesStringDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	isUnique, err := handler.checkUniqueStackName(endpoint, payload.StackName, 0)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for name collision", Err: err}
	}
	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' already exists", payload.StackName), Err: errStackAlreadyExists}
	}

	tokenData, permissionErr := handler.checkEndpointPermission(r, payload.Namespace, endpoint)
	if permissionErr != nil {
		return permissionErr
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:              portainer.StackID(stackID),
		Type:            portainer.KubernetesStack,
		EndpointID:      endpoint.ID,
		EntryPoint:      filesystem.ManifestFileDefaultName,
		Name:            payload.StackName,
		Namespace:       payload.Namespace,
		Status:          portainer.StackStatusActive,
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
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: errMsg, Err: err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	output, err := handler.deployKubernetesStack(tokenData.ID, endpoint, stack, k.KubeAppLabels{
		StackID:   stackID,
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "content",
	})
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to deploy Kubernetes stack", Err: err}
	}

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the Kubernetes stack inside the database", Err: err}
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}

	useractivity.LogHttpActivity(handler.UserActivityStore, endpoint.Name, r, payload)

	doCleanUp = false
	return response.JSON(w, resp)
}

func (handler *Handler) createKubernetesStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Environment type does not match", Err: errors.New("Environment type does not match")}
	}

	var payload kubernetesGitDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	isUnique, err := handler.checkUniqueStackName(endpoint, payload.StackName, 0)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for name collision", Err: err}
	}
	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' already exists", payload.StackName), Err: errStackAlreadyExists}
	}

	tokenData, permissionErr := handler.checkEndpointPermission(r, payload.Namespace, endpoint)
	if permissionErr != nil {
		return permissionErr
	}

	//make sure the webhook ID is unique
	if payload.AutoUpdate != nil && payload.AutoUpdate.Webhook != "" {
		isUnique, err := handler.checkUniqueWebhookID(payload.AutoUpdate.Webhook)
		if err != nil {
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for webhook ID collision", Err: err}
		}
		if !isUnique {
			return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("Webhook ID: %s already exists", payload.AutoUpdate.Webhook), Err: errWebhookIDAlreadyExists}
		}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:         portainer.StackID(stackID),
		Type:       portainer.KubernetesStack,
		EndpointID: endpoint.ID,
		EntryPoint: payload.ManifestFile,
		GitConfig: &gittypes.RepoConfig{
			URL:            payload.RepositoryURL,
			ReferenceName:  payload.RepositoryReferenceName,
			ConfigFilePath: payload.ManifestFile,
		},
		Namespace:       payload.Namespace,
		Name:            payload.StackName,
		Status:          portainer.StackStatusActive,
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
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch git repository id", Err: err}
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
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Failed to clone git repository", Err: err}
	}

	output, err := handler.deployKubernetesStack(tokenData.ID, endpoint, stack, k.KubeAppLabels{
		StackID:   stackID,
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "git",
	})
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to deploy Kubernetes stack", Err: err}
	}

	if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
		jobID, e := startAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.UserActivityStore)
		if e != nil {
			return e
		}

		stack.AutoUpdate.JobID = jobID
	}

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the Kubernetes stack inside the database", Err: err}
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}
	if payload.RepositoryPassword != "" {
		payload.RepositoryPassword = consts.RedactedValue
	}

	useractivity.LogHttpActivity(handler.UserActivityStore, endpoint.Name, r, payload)

	doCleanUp = false
	return response.JSON(w, resp)
}

func (handler *Handler) createKubernetesStackFromManifestURL(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	var payload kubernetesManifestURLDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	isUnique, err := handler.checkUniqueStackName(endpoint, payload.StackName, 0)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for name collision", Err: err}
	}
	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' already exists", payload.StackName), Err: errStackAlreadyExists}
	}

	tokenData, permissionErr := handler.checkEndpointPermission(r, payload.Namespace, endpoint)
	if permissionErr != nil {
		return permissionErr
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:              portainer.StackID(stackID),
		Type:            portainer.KubernetesStack,
		EndpointID:      endpoint.ID,
		EntryPoint:      filesystem.ManifestFileDefaultName,
		Namespace:       payload.Namespace,
		Name:            payload.StackName,
		Status:          portainer.StackStatusActive,
		CreationDate:    time.Now().Unix(),
		CreatedBy:       tokenData.Username,
		IsComposeFormat: payload.ComposeFormat,
	}

	var manifestContent []byte
	manifestContent, err = client.Get(payload.ManifestURL, 30)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve manifest from URL", Err: err}
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, manifestContent)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist Kubernetes manifest file on disk", Err: err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	output, err := handler.deployKubernetesStack(tokenData.ID, endpoint, stack, k.KubeAppLabels{
		StackID:   stackID,
		StackName: stack.Name,
		Owner:     stack.CreatedBy,
		Kind:      "url",
	})
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to deploy Kubernetes stack", Err: err}
	}

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the Kubernetes stack inside the database", Err: err}
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}

	useractivity.LogHttpActivity(handler.UserActivityStore, endpoint.Name, r, payload)

	doCleanUp = false
	return response.JSON(w, resp)
}

func (handler *Handler) deployKubernetesStack(userID portainer.UserID, endpoint *portainer.Endpoint, stack *portainer.Stack, appLabels k.KubeAppLabels) (string, error) {
	handler.stackCreationMutex.Lock()
	defer handler.stackCreationMutex.Unlock()

	manifestFilePaths, tempDir, err := stackutils.CreateTempK8SDeploymentFiles(stack, handler.KubernetesDeployer, appLabels)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp kub deployment files")
	}
	defer os.RemoveAll(tempDir)
	return handler.KubernetesDeployer.Deploy(userID, endpoint, manifestFilePaths, stack.Namespace)
}

func (handler *Handler) checkEndpointPermission(r *http.Request, namespace string, endpoint *portainer.Endpoint) (*portainer.TokenData, *httperror.HandlerError) {
	permissionDeniedErr := errors.New("Permission denied to access environment")
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: err}
	}

	if tokenData.Role == portainer.AdministratorRole {
		return tokenData, nil
	}

	// check if the user has OperationK8sApplicationsAdvancedDeploymentRW access in the environment(endpoint)
	endpointRole, err := handler.AuthorizationService.GetUserEndpointRole(int(tokenData.ID), int(endpoint.ID))
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: err}
	}
	if !endpointRole.Authorizations[portainer.OperationK8sApplicationsAdvancedDeploymentRW] {
		return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: permissionDeniedErr}
	}

	// will skip if user can access all namespaces
	if !endpointRole.Authorizations[portainer.OperationK8sAccessAllNamespaces] {
		cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
		if err != nil {
			return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Unable to create Kubernetes client", Err: err}
		}
		// check if the user has RW access to the namespace
		namespaceAuthorizations, err := handler.AuthorizationService.GetNamespaceAuthorizations(int(tokenData.ID), *endpoint, cli)
		if err != nil {
			return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: err}
		}
		if auth, ok := namespaceAuthorizations[namespace]; !ok || !auth[portainer.OperationK8sAccessNamespaceWrite] {
			return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: permissionDeniedErr}
		}
	}

	return tokenData, nil
}

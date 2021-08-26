package stacks

import (
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/internal/endpointutils"
	consts "github.com/portainer/portainer/api/useractivity"
)

const defaultReferenceName = "refs/heads/master"

type kubernetesStringDeploymentPayload struct {
	ComposeFormat    bool
	Namespace        string
	StackFileContent string
}

type kubernetesGitDeploymentPayload struct {
	ComposeFormat            bool
	Namespace                string
	RepositoryURL            string
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	FilePathInRepository     string
}

func (payload *kubernetesStringDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	if govalidator.IsNull(payload.Namespace) {
		return errors.New("Invalid namespace")
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
	if govalidator.IsNull(payload.FilePathInRepository) {
		return errors.New("Invalid file path in repository")
	}
	if govalidator.IsNull(payload.RepositoryReferenceName) {
		payload.RepositoryReferenceName = defaultReferenceName
	}
	return nil
}

type createKubernetesStackResponse struct {
	Output string `json:"Output"`
}

func (handler *Handler) createKubernetesStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Endpoint type does not match", Err: errors.New("Endpoint type does not match")}
	}
	var payload kubernetesStringDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}
	username, permissionErr := handler.checkEndpointPermission(r, payload.Namespace, endpoint)
	if permissionErr != nil {
		return permissionErr
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Type:         portainer.KubernetesStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ManifestFileDefaultName,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
		CreatedBy:    username,
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist Kubernetes manifest file on disk", Err: err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	output, err := handler.deployKubernetesStack(r, endpoint, payload.StackFileContent, payload.ComposeFormat, payload.Namespace)
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

	return response.JSON(w, resp)
}

func (handler *Handler) createKubernetesStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Endpoint type does not match", Err: errors.New("Endpoint type does not match")}
	}
	var payload kubernetesGitDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}
	username, permissionErr := handler.checkEndpointPermission(r, payload.Namespace, endpoint)
	if permissionErr != nil {
		return permissionErr
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Type:         portainer.KubernetesStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   payload.FilePathInRepository,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
		CreatedBy:    username,
	}

	projectPath := handler.FileService.GetStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	stackFileContent, err := handler.cloneManifestContentFromGitRepo(&payload, stack.ProjectPath)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Failed to process Kubernetes manifest from Git repository", Err: err}
	}

	output, err := handler.deployKubernetesStack(r, endpoint, stackFileContent, payload.ComposeFormat, payload.Namespace)
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
	if payload.RepositoryPassword != "" {
		payload.RepositoryPassword = consts.RedactedValue
	}
	useractivity.LogHttpActivity(handler.UserActivityStore, endpoint.Name, r, payload)
	return response.JSON(w, resp)
}

func (handler *Handler) deployKubernetesStack(request *http.Request, endpoint *portainer.Endpoint, stackConfig string, composeFormat bool, namespace string) (string, error) {
	handler.stackCreationMutex.Lock()
	defer handler.stackCreationMutex.Unlock()

	if composeFormat {
		convertedConfig, err := handler.KubernetesDeployer.ConvertCompose(stackConfig)
		if err != nil {
			return "", err
		}
		stackConfig = string(convertedConfig)
	}

	return handler.KubernetesDeployer.Deploy(request, endpoint, stackConfig, namespace)

}
func (handler *Handler) checkEndpointPermission(r *http.Request, namespace string, endpoint *portainer.Endpoint) (string, *httperror.HandlerError) {
	permissionDeniedErr := errors.New("Permission denied to access endpoint")
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return "", &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: err}
	}

	if tokenData.Role == portainer.AdministratorRole {
		return tokenData.Username, nil
	}

	// check if the user has OperationK8sApplicationsAdvancedDeploymentRW access in the endpoint
	endpointRole, err := handler.AuthorizationService.GetUserEndpointRole(int(tokenData.ID), int(endpoint.ID))
	if err != nil {
		return "", &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: err}
	}
	if !endpointRole.Authorizations[portainer.OperationK8sApplicationsAdvancedDeploymentRW] {
		return "", &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: permissionDeniedErr}
	}

	// will skip if user can access all namespaces
	if !endpointRole.Authorizations[portainer.OperationK8sAccessAllNamespaces] {
		cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
		if err != nil {
			return "", &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Unable to create Kubernetes client", Err: err}
		}
		// check if the user has RW access to the namespace
		namespaceAuthorizations, err := handler.AuthorizationService.GetNamespaceAuthorizations(int(tokenData.ID), *endpoint, cli)
		if err != nil {
			return "", &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: err}
		}
		if auth, ok := namespaceAuthorizations[namespace]; !ok || !auth[portainer.OperationK8sAccessNamespaceWrite] {
			return "", &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: permissionDeniedErr.Error(), Err: permissionDeniedErr}
		}
	}

	return tokenData.Username, nil
}

//read the cloned manifest yaml file from git repo and convert it to string
func (handler *Handler) cloneManifestContentFromGitRepo(gitInfo *kubernetesGitDeploymentPayload, projectPath string) (string, error) {
	repositoryUsername := gitInfo.RepositoryUsername
	repositoryPassword := gitInfo.RepositoryPassword
	if !gitInfo.RepositoryAuthentication {
		repositoryUsername = ""
		repositoryPassword = ""
	}

	err := handler.GitService.CloneRepository(projectPath, gitInfo.RepositoryURL, gitInfo.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(filepath.Join(projectPath, gitInfo.FilePathInRepository))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

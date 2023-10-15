package stacks

import (
	"net/http"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/git"
	gittypes "github.com/portainer/portainer/api/git/types"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type stackGitRedployPayload struct {
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	Env                      []portainer.Pair
	Prune                    bool `example:"false"`
	// Force a pulling to current image with the original tag though the image is already the latest
	PullImage                 bool `example:"false"`
	RepositoryGitCredentialID int
}

func (payload *stackGitRedployPayload) Validate(r *http.Request) error {
	return nil
}

// @id StackGitRedeploy
// @summary Redeploy a stack
// @description Pull and redeploy a stack via Git
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Stack identifier"
// @param endpointId query int false "Stacks created before version 1.18.0 might not have an associated environment(endpoint) identifier. Use this optional parameter to set the environment(endpoint) identifier used by the stack."
// @param body body stackGitRedployPayload true "Git configs for pull and redeploy a stack"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id}/git/redeploy [put]
func (handler *Handler) stackGitRedeploy(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().Read(portainer.StackID(stackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	}

	if stack.GitConfig == nil {
		return httperror.BadRequest("Stack is not created from git", err)
	}

	// TODO: this is a work-around for stacks created with Portainer version >= 1.17.1
	// The EndpointID property is not available for these stacks, this API environment(endpoint)
	// can use the optional EndpointID query parameter to associate a valid environment(endpoint) identifier to the stack.
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", true)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}
	if endpointID != int(stack.EndpointID) {
		stack.EndpointID = portainer.EndpointID(endpointID)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the environment associated to the stack inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the environment associated to the stack inside the database", err)
	}
	middlewares.SetEndpoint(endpoint, r)

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	//only check resource control when it is a DockerSwarmStack or a DockerComposeStack
	if stack.Type == portaineree.DockerSwarmStack || stack.Type == portaineree.DockerComposeStack {

		resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve a resource control associated to the stack", err)
		}

		access, err := handler.userCanAccessStack(securityContext, endpoint.ID, resourceControl)
		if err != nil {
			return httperror.InternalServerError("Unable to verify user authorizations to validate stack access", err)
		}
		if !access {
			return httperror.Forbidden("Access denied to resource", httperrors.ErrResourceAccessDenied)
		}
	}

	canManage, err := handler.userCanManageStacks(securityContext, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to verify user authorizations to validate stack deletion", err)
	}
	if !canManage {
		errMsg := "Stack management is disabled for non-admin users"
		return httperror.Forbidden(errMsg, errors.New(errMsg))
	}

	var payload stackGitRedployPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	stack.GitConfig.ReferenceName = payload.RepositoryReferenceName
	stack.Env = payload.Env
	if stack.Type == portaineree.DockerSwarmStack {
		stack.Option = &portainer.StackOption{
			Prune: payload.Prune,
		}
	}

	repositoryUsername := ""
	repositoryPassword := ""
	repositoryGitCredentialID := 0
	if payload.RepositoryAuthentication {
		if payload.RepositoryGitCredentialID != 0 {
			credential, err := handler.DataStore.GitCredential().Read(portaineree.GitCredentialID(payload.RepositoryGitCredentialID))
			if err != nil {
				return httperror.InternalServerError("Git credential not found", err)
			}

			// Only check the ownership of git credential when it is updated
			if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && payload.RepositoryGitCredentialID != stack.GitConfig.Authentication.GitCredentialID && credential.UserID != securityContext.UserID {
				return httperror.Forbidden("Couldn't update the git credential for another user", httperrors.ErrUnauthorized)
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

		// When the existing stack is using the custom username/password and the password is not updated,
		// the stack should keep using the saved username/password
		if payload.RepositoryPassword == "" && payload.RepositoryGitCredentialID == 0 &&
			stack.GitConfig != nil && stack.GitConfig.Authentication != nil {
			repositoryUsername = stack.GitConfig.Authentication.Username
			repositoryPassword = stack.GitConfig.Authentication.Password
			repositoryGitCredentialID = stack.GitConfig.Authentication.GitCredentialID
		}

		stack.GitConfig.Authentication = &gittypes.GitAuthentication{
			Username:        repositoryUsername,
			Password:        repositoryPassword,
			GitCredentialID: repositoryGitCredentialID,
		}
	}

	newHash, err := handler.GitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, repositoryUsername, repositoryPassword, stack.GitConfig.TLSSkipVerify)
	if err != nil {
		return httperror.InternalServerError("Unable get latest commit id", errors.WithMessagef(err, "failed to fetch latest commit id of the stack %v", stack.ID))
	}

	if stack.GitConfig.ConfigHash != newHash {
		folderToBeRemoved := stackutils.GetStackVersionFoldersToRemove(true, stack.ProjectPath, stack.GitConfig, stack.PreviousDeploymentInfo, true)

		stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
			ConfigHash:  stack.GitConfig.ConfigHash,
			FileVersion: stack.StackFileVersion,
		}
		stack.GitConfig.ConfigHash = newHash
		// When the commit hash is different, we consume the stack file different
		stack.StackFileVersion++

		// Although the git clone operation will be executed in portainer-unpacker if relative path feature is
		// enabled, it is still necessasry to clone the repository in our data volume to keep the stack file
		// consistency, especially after introducing the new feature of stack file versioning.
		projectVersionPath := handler.FileService.FormProjectPathByVersion(stack.ProjectPath, 0, stack.GitConfig.ConfigHash)
		err = handler.GitService.CloneRepository(projectVersionPath,
			stack.GitConfig.URL,
			stack.GitConfig.ReferenceName,
			repositoryUsername,
			repositoryPassword,
			stack.GitConfig.TLSSkipVerify)
		if err != nil {
			if errors.Is(err, gittypes.ErrAuthenticationFailure) {
				return httperror.InternalServerError("Unable to clone git repository directory", git.ErrInvalidGitCredential)
			}

			return httperror.InternalServerError("Unable to clone git repository directory", err)
		}

		// only keep the latest version of the stack folder
		stackutils.RemoveStackVersionFolders(folderToBeRemoved, func() {
			log.Info().Err(err).Msg("failed to remove the old stack version folder")
		})
	}

	log.Debug().Bool("pull_image_flag", payload.PullImage).Msg("")

	httpErr := handler.deployStack(r, stack, payload.PullImage, endpoint)
	if httpErr != nil {
		return httpErr
	}

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return httperror.BadRequest("Cannot find context user", errors.Wrap(err, "failed to fetch the user"))
	}
	stack.UpdatedBy = user.Username
	stack.UpdateDate = time.Now().Unix()
	stack.Status = portaineree.StackStatusActive

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil &&
		stack.GitConfig.Authentication.GitCredentialID != 0 {
		// prevent the username and password from saving into db if the git
		// credential is used
		stack.GitConfig.Authentication.Username = ""
		stack.GitConfig.Authentication.Password = ""
	}

	err = handler.DataStore.Stack().Update(stack.ID, stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", errors.Wrap(err, "failed to update the stack"))
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil &&
		stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

func (handler *Handler) deployStack(r *http.Request, stack *portaineree.Stack, pullImage bool, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	var (
		deploymentConfiger deployments.StackDeploymentConfiger
		err                error
	)

	switch stack.Type {
	case portaineree.DockerSwarmStack:
		prune := false
		if stack.Option != nil {
			prune = stack.Option.Prune
		}

		// Create swarm deployment config
		securityContext, err := security.RetrieveRestrictedRequestContext(r)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve info from request context", err)
		}

		// When the relative path feature is enabled in the docker swarm environment,
		// redeployment can be skipped if both of conditions below are met:
		// 1. If the pull image is not enforced
		// 2. If git repository has no changes since last deployment
		// The reason is that Docker swarm needs to forcibly recreate a docker container in
		// every redeployment in order to mount the relative path, this check will avoid
		// unnecessary docker container recreation.
		if isRelativePathEnabled(stack) && !pullImage {
			repositoryName := ""
			repositoryPwd := ""
			if stack.GitConfig.Authentication != nil {
				repositoryName = stack.GitConfig.Authentication.Username
				repositoryPwd = stack.GitConfig.Authentication.Password
			}
			newHash, err := handler.GitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, repositoryName, repositoryPwd, stack.GitConfig.TLSSkipVerify)
			if err != nil {
				return httperror.InternalServerError("Unable get latest commit id", errors.WithMessagef(err, "failed to fetch latest commit id of the remote swarm stack %v", stack.ID))
			}
			if stack.GitConfig.ConfigHash == newHash {
				// Skip the swarm stack redeployment
				return nil
			}
		}

		deploymentConfiger, err = deployments.CreateSwarmStackDeploymentConfig(securityContext, stack, endpoint, handler.DataStore, handler.FileService, handler.StackDeployer, prune, pullImage)
		if err != nil {
			return httperror.InternalServerError(err.Error(), err)
		}

	case portaineree.DockerComposeStack:
		// Create compose deployment config
		securityContext, err := security.RetrieveRestrictedRequestContext(r)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve info from request context", err)
		}

		deploymentConfiger, err = deployments.CreateComposeStackDeploymentConfig(securityContext, stack, endpoint, handler.DataStore, handler.FileService, handler.StackDeployer, pullImage, true)
		if err != nil {
			return httperror.InternalServerError(err.Error(), err)
		}

	case portaineree.KubernetesStack:
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			return httperror.BadRequest("Failed to retrieve user token data", err)
		}

		appLabels := k.KubeAppLabels{
			StackID:   int(stack.ID),
			StackName: stack.Name,
			Owner:     tokenData.Username,
			Kind:      "git",
		}

		deploymentConfiger, err = deployments.CreateKubernetesStackDeploymentConfig(stack, handler.KubernetesDeployer, appLabels, tokenData, endpoint, handler.AuthorizationService, handler.KubernetesClientFactory)
		if err != nil {
			return httperror.InternalServerError(err.Error(), err)
		}

	default:
		return httperror.InternalServerError("Unsupported stack", errors.Errorf("unsupported stack type: %v", stack.Type))
	}

	err = deploymentConfiger.Deploy()
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}
	return nil
}

func isRelativePathEnabled(stack *portaineree.Stack) bool {
	return stack.SupportRelativePath && stack.FilesystemPath != ""
}

package stacks

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/git"
	gittypes "github.com/portainer/portainer/api/git/types"
	logger "github.com/sirupsen/logrus"
)

type stackGitRedployPayload struct {
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	Env                      []portaineree.Pair
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

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if err == bolterrors.ErrObjectNotFound {
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
		stack.EndpointID = portaineree.EndpointID(endpointID)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if err == bolterrors.ErrObjectNotFound {
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
		stack.Option = &portaineree.StackOption{
			Prune: payload.Prune,
		}
	}

	backupProjectPath := fmt.Sprintf("%s-old", stack.ProjectPath)
	err = filesystem.MoveDirectory(stack.ProjectPath, backupProjectPath)
	if err != nil {
		return httperror.InternalServerError("Unable to move git repository directory", err)
	}

	repositoryUsername := ""
	repositoryPassword := ""
	repositoryGitCredentialID := 0
	if payload.RepositoryAuthentication {
		if payload.RepositoryGitCredentialID != 0 {
			credential, err := handler.DataStore.GitCredential().GetGitCredential(portaineree.GitCredentialID(payload.RepositoryGitCredentialID))
			if err != nil {
				return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Git credential not found", Err: err}
			}

			// Only check the ownership of git credential when it is updated
			if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && payload.RepositoryGitCredentialID != stack.GitConfig.Authentication.GitCredentialID && credential.UserID != securityContext.UserID {
				return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Couldn't update the git credential for another user", Err: httperrors.ErrUnauthorized}
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

	err = handler.GitService.CloneRepository(stack.ProjectPath, stack.GitConfig.URL, payload.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		restoreError := filesystem.MoveDirectory(backupProjectPath, stack.ProjectPath)
		if restoreError != nil {
			log.Printf("[WARN] [http,stacks,git] [error: %s] [message: failed restoring backup folder]", restoreError)
		}

		if err == git.ErrAuthenticationFailure {
			return httperror.InternalServerError(errInvalidGitCredential.Error(), err)
		}
		return httperror.InternalServerError("Unable to clone git repository", err)
	}

	defer func() {
		err = handler.FileService.RemoveDirectory(backupProjectPath)
		if err != nil {
			log.Printf("[WARN] [http,stacks,git] [error: %s] [message: unable to remove git repository directory]", err)
		}
	}()

	logger.Debugf("Pull image flag is %t", payload.PullImage)
	httpErr := handler.deployStack(r, stack, payload.PullImage, endpoint)
	if httpErr != nil {
		return httpErr
	}

	newHash, err := handler.GitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		return httperror.InternalServerError("Unable get latest commit id", errors.WithMessagef(err, "failed to fetch latest commit id of the stack %v", stack.ID))
	}
	stack.GitConfig.ConfigHash = newHash

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return httperror.BadRequest("Cannot find context user", errors.Wrap(err, "failed to fetch the user"))
	}
	stack.UpdatedBy = user.Username
	stack.UpdateDate = time.Now().Unix()
	stack.Status = portaineree.StackStatusActive

	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", errors.Wrap(err, "failed to update the stack"))
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

func (handler *Handler) deployStack(r *http.Request, stack *portaineree.Stack, pullImage bool, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	switch stack.Type {
	case portaineree.DockerSwarmStack:
		prune := false
		if stack.Option != nil {
			prune = stack.Option.Prune
		}
		config, httpErr := handler.createSwarmDeployConfig(r, stack, endpoint, prune, pullImage)
		if httpErr != nil {
			return httpErr
		}

		if err := handler.deploySwarmStack(config); err != nil {
			return httperror.InternalServerError(err.Error(), err)
		}

	case portaineree.DockerComposeStack:
		config, httpErr := handler.createComposeDeployConfig(r, stack, endpoint, pullImage)
		if httpErr != nil {
			return httpErr
		}

		if err := handler.deployComposeStack(config, true); err != nil {
			return httperror.InternalServerError(err.Error(), err)
		}

	case portaineree.KubernetesStack:
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			return httperror.BadRequest("Failed to retrieve user token data", err)
		}
		_, deployError := handler.deployKubernetesStack(tokenData, endpoint, stack, k.KubeAppLabels{
			StackID:   int(stack.ID),
			StackName: stack.Name,
			Owner:     tokenData.Username,
			Kind:      "git",
		})
		if deployError != nil {
			return deployError
		}

	default:
		return httperror.InternalServerError("Unsupported stack", errors.Errorf("unsupported stack type: %v", stack.Type))
	}

	return nil
}

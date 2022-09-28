package stacks

import (
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
	gittypes "github.com/portainer/portainer/api/git/types"
)

type stackGitUpdatePayload struct {
	AutoUpdate                *portaineree.StackAutoUpdate
	Env                       []portaineree.Pair
	Prune                     bool
	RepositoryReferenceName   string
	RepositoryAuthentication  bool
	RepositoryUsername        string
	RepositoryPassword        string
	RepositoryGitCredentialID int
}

func (payload *stackGitUpdatePayload) Validate(r *http.Request) error {
	if err := validateStackAutoUpdate(payload.AutoUpdate); err != nil {
		return err
	}
	return nil
}

// @id StackUpdateGit
// @summary Update a stack's Git configs
// @description Update the Git settings in a stack, e.g., RepositoryReferenceName and AutoUpdate
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Stack identifier"
// @param endpointId query int false "Stacks created before version 1.18.0 might not have an associated environment(endpoint) identifier. Use this optional parameter to set the environment(endpoint) identifier used by the stack."
// @param body body stackGitUpdatePayload true "Git configs for pull and redeploy a stack"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id}/git [post]
func (handler *Handler) stackUpdateGit(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	var payload stackGitUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	} else if stack.GitConfig == nil {
		msg := "No Git config in the found stack"
		return httperror.InternalServerError(msg, errors.New(msg))
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

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return httperror.BadRequest("Cannot find context user", errors.Wrap(err, "failed to fetch the user"))
	}

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
		errMsg := "Stack editing is disabled for non-admin users"
		return httperror.Forbidden(errMsg, errors.New(errMsg))
	}

	//stop the autoupdate job if there is any
	if stack.AutoUpdate != nil {
		stopAutoupdate(stack.ID, stack.AutoUpdate.JobID, *handler.Scheduler)
	}

	//update retrieved stack data based on the payload
	stack.GitConfig.ReferenceName = payload.RepositoryReferenceName
	stack.AutoUpdate = payload.AutoUpdate
	stack.Env = payload.Env
	stack.UpdatedBy = user.Username
	stack.UpdateDate = time.Now().Unix()

	if stack.Type == portaineree.DockerSwarmStack {
		stack.Option = &portaineree.StackOption{
			Prune: payload.Prune,
		}
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
			if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && payload.RepositoryGitCredentialID != stack.GitConfig.Authentication.GitCredentialID && credential.UserID != user.ID {
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
		_, err = handler.GitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, repositoryUsername, repositoryPassword)
		if err != nil {
			return httperror.InternalServerError("Unable to fetch git repository", err)
		}
	} else {
		stack.GitConfig.Authentication = nil
	}

	if payload.AutoUpdate != nil && payload.AutoUpdate.Interval != "" {
		jobID, e := startAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService)
		if e != nil {
			return e
		}

		stack.AutoUpdate.JobID = jobID
	}

	//save the updated stack to DB
	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

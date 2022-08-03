package stacks

import (
	"fmt"
	"net/http"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id StackAssociate
// @summary Associate an orphaned stack to a new environment(endpoint)
// @description **Access policy**: administrator
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Stack identifier"
// @param endpointId query int true "Stacks created before version 1.18.0 might not have an associated environment(endpoint) identifier. Use this optional parameter to set the environment(endpoint) identifier used by the stack."
// @param swarmId query int true "Swarm identifier"
// @param orphanedRunning query boolean true "Indicates whether the stack is orphaned"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Stack not found"
// @failure 500 "Server error"
// @router /stacks/{id}/associate [put]
func (handler *Handler) stackAssociate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid stack identifier route variable", err}
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: endpointId", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}
	middlewares.SetEndpoint(endpoint, r)

	swarmId, err := request.RetrieveQueryParameter(r, "swarmId", true)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: swarmId", err}
	}

	orphanedRunning, err := request.RetrieveBooleanQueryParameter(r, "orphanedRunning", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: orphanedRunning", err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to load user information from the database", err}
	}

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a stack with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a stack with the specified identifier inside the database", err}
	}

	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve a resource control associated to the stack", err}
	}

	if resourceControl != nil {
		resourceControl.ResourceID = fmt.Sprintf("%d_%s", endpointID, stack.Name)

		err = handler.DataStore.ResourceControl().UpdateResourceControl(resourceControl.ID, resourceControl)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist resource control changes inside the database", err}
		}
	}

	canManage, err := handler.userCanManageStacks(securityContext, endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to verify user authorizations to validate stack deletion", Err: err}
	}
	if !canManage {
		errMsg := "Stack management is disabled for non-admin users"
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: errMsg, Err: fmt.Errorf(errMsg)}
	}

	stack.EndpointID = portaineree.EndpointID(endpointID)
	stack.SwarmID = swarmId

	if orphanedRunning {
		stack.Status = portaineree.StackStatusActive
	} else {
		stack.Status = portaineree.StackStatusInactive
	}

	stack.CreationDate = time.Now().Unix()
	stack.CreatedBy = user.Username
	stack.UpdateDate = 0
	stack.UpdatedBy = ""

	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the stack changes inside the database", err}
	}

	stack.ResourceControl = resourceControl

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

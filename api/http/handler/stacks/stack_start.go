package stacks

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/stackutils"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	bolterrors "github.com/portainer/portainer-ee/api/bolt/errors"
)

// @id StackStart
// @summary Starts a stopped Stack
// @description Starts a stopped Stack.
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Stack identifier"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id}/start [post]
func (handler *Handler) stackStart(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid stack identifier route variable", Err: err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve info from request context", Err: err}
	}

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	}

	if stack.Type == portaineree.KubernetesStack {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Starting a kubernetes stack is not supported", Err: err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an endpoint with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an endpoint with the specified identifier inside the database", Err: err}
	}
	middlewares.SetEndpoint(endpoint, r)

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Permission denied to access endpoint", Err: err}
	}

	isUnique, err := handler.checkUniqueStackNameInDocker(endpoint, stack.Name, stack.ID, stack.SwarmID != "")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for name collision", Err: err}
	}
	if !isUnique {
		errorMessage := fmt.Sprintf("A stack with the name '%s' is already running", stack.Name)
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: errorMessage, Err: errors.New(errorMessage)}
	}

	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portaineree.StackResourceControl)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve a resource control associated to the stack", Err: err}
	}

	access, err := handler.userCanAccessStack(securityContext, endpoint.ID, resourceControl)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to verify user authorizations to validate stack access", Err: err}
	}
	if !access {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Access denied to resource", Err: httperrors.ErrResourceAccessDenied}
	}

	if stack.Status == portaineree.StackStatusActive {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Stack is already active", Err: errors.New("Stack is already active")}
	}

	if stack.AutoUpdate != nil && stack.AutoUpdate.Interval != "" {
		stopAutoupdate(stack.ID, stack.AutoUpdate.JobID, *handler.Scheduler)

		jobID, e := startAutoupdate(stack.ID, stack.AutoUpdate.Interval, handler.Scheduler, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService)
		if e != nil {
			return e
		}

		stack.AutoUpdate.JobID = jobID
	}

	err = handler.startStack(stack, endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to start stack", Err: err}
	}

	stack.Status = portaineree.StackStatusActive
	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to update stack status", Err: err}
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

func (handler *Handler) startStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	switch stack.Type {
	case portaineree.DockerComposeStack:
		return handler.ComposeStackManager.Up(context.TODO(), stack, endpoint, false)
	case portaineree.DockerSwarmStack:
		return handler.SwarmStackManager.Deploy(stack, true, true, endpoint)
	}
	return nil
}

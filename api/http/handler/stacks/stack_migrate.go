package stacks

import (
	"errors"
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

type stackMigratePayload struct {
	EndpointID int    `json:",omitempty" example:"2" validate:"required"`
	SwarmID    string `json:",omitempty" example:"jpofkc0i9uo9wtx1zesuk649w"`
	Name       string `json:",omitempty" example:"new-stack"`
}

func (payload *stackMigratePayload) Validate(r *http.Request) error {
	if payload.EndpointID == 0 {
		return errors.New("Invalid environment identifier. Must be a positive number")
	}
	return nil
}

// @id StackMigrate
// @summary Migrate a stack to another environment(endpoint)
// @description  Migrate a stack from an environment(endpoint) to another environment(endpoint). It will re-create the stack inside the target environment(endpoint) before removing the original stack.
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Stack identifier"
// @param endpointId query int false "Stacks created before version 1.18.0 might not have an associated environment(endpoint) identifier. Use this optional parameter to set the environment(endpoint) identifier used by the stack."
// @param body body stackMigratePayload true "Stack migration details"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Stack not found"
// @failure 500 "Server error"
// @router /stacks/{id}/migrate [post]
func (handler *Handler) stackMigrate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	var payload stackMigratePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	}

	if stack.Type == portaineree.KubernetesStack {
		return httperror.BadRequest("Migrating a kubernetes stack is not supported", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find an endpoint with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an endpoint with the specified identifier inside the database", err)
	}
	middlewares.SetEndpoint(endpoint, r)

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access endpoint", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

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

	canManage, err := handler.userCanManageStacks(securityContext, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to verify user authorizations to validate stack deletion", err)
	}
	if !canManage {
		errMsg := "Stack migration is disabled for non-admin users"
		return httperror.Forbidden(errMsg, errors.New(errMsg))
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

	targetEndpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(payload.EndpointID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find an endpoint with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an endpoint with the specified identifier inside the database", err)
	}

	stack.EndpointID = portaineree.EndpointID(payload.EndpointID)
	if payload.SwarmID != "" {
		stack.SwarmID = payload.SwarmID
	}

	oldName := stack.Name
	if payload.Name != "" {
		stack.Name = payload.Name
	}

	isUnique, err := handler.checkUniqueStackNameInDocker(targetEndpoint, stack.Name, stack.ID, stack.SwarmID != "")
	if err != nil {
		return httperror.InternalServerError("Unable to check for name collision", err)
	}

	if !isUnique {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("A stack with the name '%s' is already running on endpoint '%s'", stack.Name, targetEndpoint.Name), Err: errStackAlreadyExists}
	}

	migrationError := handler.migrateStack(r, stack, targetEndpoint)
	if migrationError != nil {
		return migrationError
	}

	newName := stack.Name
	stack.Name = oldName
	err = handler.deleteStack(securityContext.UserID, stack, endpoint)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	stack.Name = newName
	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	if resourceControl != nil {
		resourceControl.ResourceID = stackutils.ResourceControlID(stack.EndpointID, stack.Name)
		err := handler.DataStore.ResourceControl().UpdateResourceControl(resourceControl.ID, resourceControl)
		if err != nil {
			return httperror.InternalServerError("Unable to persist the resource control changes", err)
		}
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

func (handler *Handler) migrateStack(r *http.Request, stack *portaineree.Stack, next *portaineree.Endpoint) *httperror.HandlerError {
	if stack.Type == portaineree.DockerSwarmStack {
		return handler.migrateSwarmStack(r, stack, next)
	}
	return handler.migrateComposeStack(r, stack, next)
}

func (handler *Handler) migrateComposeStack(r *http.Request, stack *portaineree.Stack, next *portaineree.Endpoint) *httperror.HandlerError {
	config, configErr := handler.createComposeDeployConfig(r, stack, next, false)
	if configErr != nil {
		return configErr
	}

	err := handler.deployComposeStack(config, false)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	return nil
}

func (handler *Handler) migrateSwarmStack(r *http.Request, stack *portaineree.Stack, next *portaineree.Endpoint) *httperror.HandlerError {
	config, configErr := handler.createSwarmDeployConfig(r, stack, next, true, true)
	if configErr != nil {
		return configErr
	}

	err := handler.deploySwarmStack(config)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	return nil
}

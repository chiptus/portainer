package stacks

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	bolterrors "github.com/portainer/portainer-ee/api/bolt/errors"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
)

type updateComposeStackPayload struct {
	// New content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx"`
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
	// Force a pulling to current image with the original tag though the image is already the latest
	PullImage bool `example:"false"`
}

func (payload *updateComposeStackPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

type updateSwarmStackPayload struct {
	// New content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx"`
	// A list of environment(endpoint) variables used during stack deployment
	Env []portaineree.Pair
	// Prune services that are no longer referenced (only available for Swarm stacks)
	Prune bool `example:"true"`
	// A UUID to identify a webhook. The stack will be force updated and pull the latest image when the webhook was invoked.
	Webhook string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
	// Force a pulling to current image with the original tag though the image is already the latest
	PullImage bool `example:"false"`
}

func (payload *updateSwarmStackPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

// @id StackUpdate
// @summary Update a stack
// @description Update a stack, only for file based stacks.
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Stack identifier"
// @param endpointId query int false "Stacks created before version 1.18.0 might not have an associated environment(endpoint) identifier. Use this optional parameter to set the environment(endpoint) identifier used by the stack."
// @param body body updateSwarmStackPayload true "Stack details"
// @success 200 {object} portaineree.Stack "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id} [put]
func (handler *Handler) stackUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid stack identifier route variable", Err: err}
	}

	stack, err := handler.DataStore.Stack().Stack(portaineree.StackID(stackID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	}

	// TODO: this is a work-around for stacks created with Portainer version >= 1.17.1
	// The EndpointID property is not available for these stacks, this API environment(endpoint)
	// can use the optional EndpointID query parameter to associate a valid environment(endpoint) identifier to the stack.
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", true)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid query parameter: endpointId", Err: err}
	}
	if endpointID != int(stack.EndpointID) {
		stack.EndpointID = portaineree.EndpointID(endpointID)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find the environment associated to the stack inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find the environment associated to the stack inside the database", Err: err}
	}
	middlewares.SetEndpoint(endpoint, r)

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Permission denied to access environment", Err: err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve info from request context", Err: err}
	}

	//only check resource control when it is a DockerSwarmStack or a DockerComposeStack
	if stack.Type == portaineree.DockerSwarmStack || stack.Type == portaineree.DockerComposeStack {
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
	}

	// Must not be git based stack. stop the auto update job if there is any
	if stack.AutoUpdate != nil {
		stopAutoupdate(stack.ID, stack.AutoUpdate.JobID, *handler.Scheduler)
		stack.AutoUpdate = nil
	}
	if stack.GitConfig != nil {
		stack.FromAppTemplate = true
	}

	updateError := handler.updateAndDeployStack(r, stack, endpoint)
	if updateError != nil {
		return updateError
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Cannot find context user", Err: errors.Wrap(err, "failed to fetch the user")}
	}
	stack.UpdatedBy = user.Username
	stack.UpdateDate = time.Now().Unix()
	stack.Status = portaineree.StackStatusActive

	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack changes inside the database", Err: err}
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
		// sanitize password in the http response to minimise possible security leaks
		stack.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, stack)
}

func (handler *Handler) updateAndDeployStack(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	if stack.Type == portaineree.DockerSwarmStack {
		return handler.updateSwarmStack(r, stack, endpoint)
	} else if stack.Type == portaineree.DockerComposeStack {
		return handler.updateComposeStack(r, stack, endpoint)
	} else if stack.Type == portaineree.KubernetesStack {
		return handler.updateKubernetesStack(r, stack, endpoint)
	} else {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unsupported stack", Err: errors.Errorf("unsupported stack type: %v", stack.Type)}
	}
}

func (handler *Handler) updateComposeStack(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	var payload updateComposeStackPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	stack.Env = payload.Env
	stack.Webhook = payload.Webhook

	if payload.Webhook != "" && stack.Webhook != payload.Webhook {
		isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
		if isUniqueError != nil {
			return isUniqueError
		}
	}
	stack.Webhook = payload.Webhook

	stackFolder := strconv.Itoa(int(stack.ID))
	_, err = handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist updated Compose file on disk", Err: err}
	}

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint, payload.PullImage)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config, false)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	return nil
}

func (handler *Handler) updateSwarmStack(r *http.Request, stack *portaineree.Stack, endpoint *portaineree.Endpoint) *httperror.HandlerError {
	var payload updateSwarmStackPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	stack.Env = payload.Env

	if payload.Webhook != "" && stack.Webhook != payload.Webhook {
		isUniqueError := handler.checkUniqueWebhookID(payload.Webhook)
		if isUniqueError != nil {
			return isUniqueError
		}
	}
	stack.Webhook = payload.Webhook

	stackFolder := strconv.Itoa(int(stack.ID))
	_, err = handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist updated Compose file on disk", Err: err}
	}

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, payload.Prune, payload.PullImage)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	return nil
}

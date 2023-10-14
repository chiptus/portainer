package endpointedge

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
)

// EdgeAsyncNormalStackCommandCreateRequest is the command request used to operate the normal stack
type EdgeAsyncNormalStackCommandCreateRequest struct {
	StackOperation portaineree.EdgeAsyncNormalStackOperation
}

func (payload *EdgeAsyncNormalStackCommandCreateRequest) Validate(r *http.Request) error {
	if len(payload.StackOperation) == 0 {
		return errors.New("stack operation is mandatory")
	}

	return nil
}

func (handler *Handler) createNormalStackCommand(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	stackId, err := request.RetrieveNumericRouteVariableValue(r, "stackId")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	var payload EdgeAsyncNormalStackCommandCreateRequest
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	stack, err := handler.DataStore.Stack().Read(portainer.StackID(stackId))
	if err != nil {
		httpErr := httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
		if handler.DataStore.IsErrObjectNotFound(err) {
			httpErr.StatusCode = http.StatusNotFound
		}
		return httpErr
	}

	switch payload.StackOperation {
	case portaineree.EdgeAsyncNormalStackOperationRemove:
		err = handler.EdgeService.RemoveNormalStackCommand(endpoint.ID, stack.ID)
		if err == nil {
			err = handler.DataStore.Stack().Delete(portainer.StackID(stackId))
		}
	}

	if err != nil {
		return httperror.InternalServerError("Unable to create edge async stack command", nil)
	}

	return response.Empty(w)
}

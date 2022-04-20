package edgestacks

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

type updateStatusPayload struct {
	Error      string
	Status     *portaineree.EdgeStackStatusType
	EndpointID *portaineree.EndpointID
}

func (payload *updateStatusPayload) Validate(r *http.Request) error {
	if payload.Status == nil {
		return errors.New("Invalid status")
	}
	if payload.EndpointID == nil {
		return errors.New("Invalid EnvironmentID")
	}
	if *payload.Status == portaineree.StatusError && govalidator.IsNull(payload.Error) {
		return errors.New("Error message is mandatory when status is error")
	}
	return nil
}

// @id EdgeStackStatusUpdate
// @summary Update an EdgeStack status
// @description Authorized only if the request is done by an Edge Environment(Endpoint)
// @tags edge_stacks
// @accept json
// @produce json
// @param id path string true "EdgeStack Id"
// @success 200 {object} portaineree.EdgeStack
// @failure 500
// @failure 400
// @failure 404
// @failure 403
// @router /edge_stacks/{id}/status [put]
func (handler *Handler) edgeStackStatusUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid stack identifier route variable", Err: err}
	}

	stack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(stackID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	}

	var payload updateStatusPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(*payload.EndpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Permission denied to access environment", Err: err}
	}

	stack.Status[*payload.EndpointID] = portaineree.EdgeStackStatus{
		Type:       *payload.Status,
		Error:      payload.Error,
		EndpointID: *payload.EndpointID,
	}

	err = handler.DataStore.EdgeStack().UpdateEdgeStack(stack.ID, stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack changes inside the database", Err: err}
	}

	return response.JSON(w, stack)

}

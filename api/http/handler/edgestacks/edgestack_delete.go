package edgestacks

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeStackDelete
// @summary Delete an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @param id path int true "EdgeStack Id"
// @success 204
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id} [delete]
func (handler *Handler) edgeStackDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		return handler.deleteEdgeStack(tx, portainer.EdgeStackID(edgeStackID))
	})
	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

func (handler *Handler) deleteEdgeStack(tx dataservices.DataStoreTx, edgeStackID portainer.EdgeStackID) error {
	edgeStack, err := tx.EdgeStack().EdgeStack(portainer.EdgeStackID(edgeStackID))
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an edge stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an edge stack with the specified identifier inside the database", err)
	}

	if edgeStack.EdgeUpdateID != 0 {
		return httperror.BadRequest("Unable to delete edge stack that is used by an edge update schedule", err)
	}

	if edgeStack.AutoUpdate != nil && edgeStack.AutoUpdate.JobID != "" {
		err := handler.scheduler.StopJob(edgeStack.AutoUpdate.JobID)
		if err != nil {
			return httperror.InternalServerError("Unable to stop auto update job", err)
		}
	}

	if handler.staggerService != nil &&
		edgeStack.StaggerConfig != nil &&
		edgeStack.StaggerConfig.StaggerOption == portaineree.EdgeStaggerOptionParallel {
		go handler.staggerService.StopAndRemoveStaggerScheduleOperation(edgeStack.ID)
	}

	err = handler.edgeStacksService.DeleteEdgeStack(tx, edgeStack.ID, edgeStack.EdgeGroups)
	if err != nil {
		return httperror.InternalServerError("Unable to delete edge stack", err)
	}

	return nil
}

package edgegroups

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/slices"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeGroupDelete
// @summary Deletes an EdgeGroup
// @description **Access policy**: administrator
// @tags edge_groups
// @security ApiKeyAuth
// @security jwt
// @param id path int true "EdgeGroup Id"
// @success 204
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups/{id} [delete]
func (handler *Handler) edgeGroupDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge group identifier route variable", err)
	}

	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		return deleteEdgeGroup(tx, portaineree.EdgeGroupID(edgeGroupID))
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

func deleteEdgeGroup(tx dataservices.DataStoreTx, ID portaineree.EdgeGroupID) error {
	_, err := tx.EdgeGroup().Read(ID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an Edge group with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an Edge group with the specified identifier inside the database", err)
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge stacks from the database", err)
	}

	for _, edgeStack := range edgeStacks {
		for _, groupID := range edgeStack.EdgeGroups {
			if groupID == ID {
				return httperror.NewError(http.StatusConflict, "Edge group is used by an Edge stack", errors.New("edge group is used by an Edge stack"))
			}
		}
	}

	edgeJobs, err := tx.EdgeJob().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge jobs from the database", err)
	}

	for _, edgeJob := range edgeJobs {
		for _, groupID := range edgeJob.EdgeGroups {
			if groupID == ID {
				return httperror.NewError(http.StatusConflict, "Edge group is used by an Edge job", errors.New("edge group is used by an Edge job"))
			}
		}
	}

	edgeConfigs, err := tx.EdgeConfig().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge configs from the database", err)
	}

	for _, edgeConfig := range edgeConfigs {
		if slices.Contains(edgeConfig.EdgeGroupIDs, ID) {
			return httperror.NewError(http.StatusConflict, "Edge group is used by an Edge config", errors.New("edge group is used by an Edge config"))
		}
	}

	err = tx.EdgeGroup().Delete(ID)
	if err != nil {
		return httperror.InternalServerError("Unable to remove the Edge group from the database", err)
	}

	return nil
}

package edgestacks

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeStackList
// @summary Fetches the list of EdgeStacks
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} portaineree.EdgeStack
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks [get]
func (handler *Handler) edgeStackList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge stacks from the database", err)
	}

	filteredEdgeStacks := []portaineree.EdgeStack{}
	for _, edgeStack := range edgeStacks {
		if edgeStack.EdgeUpdateID == 0 {
			filteredEdgeStacks = append(filteredEdgeStacks, edgeStack)
		}
	}

	return response.JSON(w, filteredEdgeStacks)
}

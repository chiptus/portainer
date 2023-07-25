package edgeconfigs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id EdgeConfigList
// @summary List available Edge Configurations
// @description **Access policy**: authenticated
// @tags edge_configs
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body edgeConfigCreatePayload true "body"
// @success 200 {array} portaineree.EdgeConfig "Success"
// @failure 500 "Server error"
// @router /edge_configurations [get]
func (h *Handler) edgeConfigList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// params := extractListModifiersQueryParams(r)

	edgeConfigurations, err := h.dataStore.EdgeConfig().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the edge configurations", err)
	}

	// TODO(LP): implement more fields to search on
	// searchGetters := SearchFieldGetters{
	// 	func(config portaineree.EdgeConfig) string { return config.Name },
	// }

	// filterResult := searchOrderAndPaginate(edgeConfigurations, params, searchGetters)
	// applyFilterResultsHeaders(&w, filterResult)

	// return response.JSON(w, filterResult.configs)
	return response.JSON(w, edgeConfigurations)
}

package edgeupdateschedules

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/utils"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeUpdateScheduleList
// @summary Fetches the list of Edge Update Schedules
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param includeEdgeStacks query boolean false "Include Edge Stacks in the response"
// @success 200 {array} decoratedUpdateSchedule
// @failure 500
// @router /edge_update_schedules [get]
func (handler *Handler) list(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	includeEdgeStacks, _ := request.RetrieveBooleanQueryParameter(r, "includeEdgeStacks", true)

	if !includeEdgeStacks {
		list, err := handler.updateService.Schedules(handler.dataStore)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve the edge update schedules list", err)
		}

		return response.JSON(w, list)
	}

	var decoratedList []decoratedUpdateSchedule
	err := handler.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		list, err := handler.updateService.Schedules(tx)
		if err != nil {
			return err
		}

		decoratedList = make([]decoratedUpdateSchedule, len(list))
		for idx, item := range list {
			decoratedItem, err := decorateSchedule(tx, item)
			if err != nil {
				return httperror.InternalServerError("Unable to decorate the edge update schedule", err)
			}

			decoratedList[idx] = *decoratedItem
		}

		return nil
	})
	return utils.TxResponse(w, decoratedList, err)
}

package webhooks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

type webhookListOperationFilters struct {
	ResourceID string `json:"ResourceID"`
	EndpointID int    `json:"EndpointID"`
}

// @summary List webhooks
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags webhooks
// @accept json
// @produce json
// @param filters query webhookListOperationFilters false "Filters"
// @success 200 {array} portaineree.Webhook
// @failure 400
// @failure 500
// @router /webhooks [get]
func (handler *Handler) webhookList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var filters webhookListOperationFilters
	err := request.RetrieveJSONQueryParameter(r, "filters", &filters, true)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: filters", err)
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(filters.EndpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}
	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	authorizations := []portaineree.Authorization{portaineree.OperationPortainerWebhookList}

	isAuthorized, handlerErr := handler.checkAuthorization(r, endpoint, authorizations)
	if handlerErr != nil || !isAuthorized {
		return response.JSON(w, []portaineree.Webhook{})
	}
	webhooks, err := handler.dataStore.Webhook().Webhooks()
	webhooks = filterWebhooks(webhooks, &filters)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve webhooks from the database", err)
	}

	return response.JSON(w, webhooks)
}

func filterWebhooks(webhooks []portaineree.Webhook, filters *webhookListOperationFilters) []portaineree.Webhook {
	if filters.EndpointID == 0 && filters.ResourceID == "" {
		return webhooks
	}

	filteredWebhooks := make([]portaineree.Webhook, 0, len(webhooks))
	for _, webhook := range webhooks {
		if webhook.EndpointID == portaineree.EndpointID(filters.EndpointID) && webhook.ResourceID == string(filters.ResourceID) {
			filteredWebhooks = append(filteredWebhooks, webhook)
		}
	}

	return filteredWebhooks
}

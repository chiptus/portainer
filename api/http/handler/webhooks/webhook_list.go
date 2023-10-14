package webhooks

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type webhookListOperationFilters struct {
	ResourceID string `json:"ResourceID" validate:"required"`
	EndpointID int    `json:"EndpointID" validate:"required"`
}

// @summary List webhooks
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags webhooks
// @accept json
// @produce json
// @param filters query webhookListOperationFilters false "Filters"
// @success 200 {array} portainer.Webhook
// @failure 400
// @failure 500
// @router /webhooks [get]
func (handler *Handler) webhookList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var filters webhookListOperationFilters
	err := request.RetrieveJSONQueryParameter(r, "filters", &filters, true)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: filters", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(filters.EndpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}
	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	authorizations := []portainer.Authorization{portaineree.OperationPortainerWebhookList}

	isAuthorized, handlerErr := handler.checkAuthorization(r, endpoint, authorizations)
	if handlerErr != nil || !isAuthorized {
		return response.JSON(w, []portainer.Webhook{})
	}

	webhooks, err := handler.DataStore.Webhook().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve webhooks from the database", err)
	}

	webhooks = filterWebhooks(webhooks, &filters)

	return response.JSON(w, webhooks)
}

func filterWebhooks(webhooks []portainer.Webhook, filters *webhookListOperationFilters) []portainer.Webhook {
	if filters.EndpointID == 0 && filters.ResourceID == "" {
		return webhooks
	}

	filteredWebhooks := make([]portainer.Webhook, 0, len(webhooks))
	for _, webhook := range webhooks {
		if webhook.EndpointID == portainer.EndpointID(filters.EndpointID) && webhook.ResourceID == string(filters.ResourceID) {
			filteredWebhooks = append(filteredWebhooks, webhook)
		}
	}

	return filteredWebhooks
}

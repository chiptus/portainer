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

// @summary Delete a webhook
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags webhooks
// @param id path int true "Webhook id"
// @success 202 "Webhook deleted"
// @failure 400
// @failure 500
// @router /webhooks/{id} [delete]
func (handler *Handler) webhookDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid webhook id", err)
	}

	webhook, err := handler.DataStore.Webhook().Read(portainer.WebhookID(id))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a webhook with this token", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to retrieve webhook from the database", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(webhook.EndpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	authorizations := []portainer.Authorization{portaineree.OperationPortainerWebhookDelete}

	_, handlerErr := handler.checkAuthorization(r, endpoint, authorizations)
	if handlerErr != nil {
		return handlerErr
	}

	err = handler.DataStore.Webhook().Delete(portainer.WebhookID(id))
	if err != nil {
		return httperror.InternalServerError("Unable to remove the webhook from the database", err)
	}

	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	return response.Empty(w)
}

package webhooks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
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

	webhook, err := handler.dataStore.Webhook().Webhook(portaineree.WebhookID(id))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a webhook with this token", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to retrieve webhook from the database", err)
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(webhook.EndpointID)
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	authorizations := []portaineree.Authorization{portaineree.OperationPortainerWebhookDelete}

	_, handlerErr := handler.checkAuthorization(r, endpoint, authorizations)
	if handlerErr != nil {
		return handlerErr
	}

	err = handler.dataStore.Webhook().DeleteWebhook(portaineree.WebhookID(id))
	if err != nil {
		return httperror.InternalServerError("Unable to remove the webhook from the database", err)
	}

	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	return response.Empty(w)
}

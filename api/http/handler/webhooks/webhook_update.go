package webhooks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils/access"
)

type webhookUpdatePayload struct {
	RegistryID portaineree.RegistryID
}

func (payload *webhookUpdatePayload) Validate(r *http.Request) error {
	return nil
}

// @summary Update a webhook
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags webhooks
// @accept json
// @produce json
// @param id path int true "Webhook id"
// @param body body webhookUpdatePayload true "Webhook data"
// @success 200 {object} portaineree.Webhook
// @failure 400
// @failure 409
// @failure 500
// @router /webhooks/{id} [put]
func (handler *Handler) webhookUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid webhook id", err)
	}
	webhookID := portaineree.WebhookID(id)

	var payload webhookUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	webhook, err := handler.DataStore.Webhook().Webhook(webhookID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a webhooks with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a webhooks with the specified identifier inside the database", err)
	}

	endpointID := webhook.EndpointID
	endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}
	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	authorizations := []portaineree.Authorization{portaineree.OperationPortainerWebhookCreate}

	_, handlerErr := handler.checkAuthorization(r, endpoint, authorizations)
	if handlerErr != nil {
		return handlerErr
	}

	if payload.RegistryID != 0 {
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve user authentication token", err)
		}

		_, err = access.GetAccessibleRegistry(handler.DataStore, tokenData.ID, endpointID, payload.RegistryID)
		if err != nil {
			return httperror.Forbidden("Permission deny to access registry", err)
		}
	}

	webhook.RegistryID = payload.RegistryID

	err = handler.DataStore.Webhook().UpdateWebhook(portaineree.WebhookID(id), webhook)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the webhook inside the database", err)
	}

	return response.JSON(w, webhook)
}

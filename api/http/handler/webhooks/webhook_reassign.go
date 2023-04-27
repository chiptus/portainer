package webhooks

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type webhookReassignPayload struct {
	ResourceID  string
	WebhookType portaineree.WebhookType
}

func (payload *webhookReassignPayload) Validate(r *http.Request) error {
	if payload.WebhookType != portaineree.ContainerWebhook && payload.WebhookType != portaineree.ServiceWebhook {
		return errors.New("Invalid WebhookType")
	}

	return nil
}

// @summary Reassign a webhook to another resource
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags webhooks
// @accept json
// @produce json
// @param id path int true "Webhook id"
// @param body body webhookReassignPayload true "Webhook data"
// @success 200 {object} portaineree.Webhook
// @success 204
// @failure 400
// @failure 404
// @failure 500
// @router /webhooks/{id}/reassign [put]
func (handler *Handler) webhookReassign(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid webhook id", err)
	}
	webhookID := portaineree.WebhookID(id)

	var payload webhookReassignPayload
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

	if payload.WebhookType != portaineree.ContainerWebhook {
		return response.Empty(w)
	}

	webhook.ResourceID = payload.ResourceID

	err = handler.DataStore.Webhook().UpdateWebhook(portaineree.WebhookID(id), webhook)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the webhook inside the database", err)
	}

	return response.JSON(w, webhook)
}

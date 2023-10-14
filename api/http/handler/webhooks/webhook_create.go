package webhooks

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils/access"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
	"github.com/gofrs/uuid"
)

type webhookCreatePayload struct {
	ResourceID string
	EndpointID portainer.EndpointID
	RegistryID portainer.RegistryID
	// Type of webhook (1 - service, 2 - container)
	WebhookType portainer.WebhookType
}

func (payload *webhookCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.ResourceID) {
		return errors.New("Invalid ResourceID")
	}
	if payload.EndpointID == 0 {
		return errors.New("Invalid EndpointID")
	}
	if payload.WebhookType != portaineree.ContainerWebhook && payload.WebhookType != portaineree.ServiceWebhook {
		return errors.New("Invalid WebhookType")
	}
	return nil
}

// @summary Create a webhook
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags webhooks
// @accept json
// @produce json
// @param body body webhookCreatePayload true "Webhook data"
// @success 200 {object} portainer.Webhook
// @failure 400
// @failure 409
// @failure 500
// @router /webhooks [post]
func (handler *Handler) webhookCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload webhookCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	webhook, err := handler.DataStore.Webhook().WebhookByResourceID(payload.ResourceID)
	if err != nil && !dataservices.IsErrObjectNotFound(err) {
		return httperror.InternalServerError("An error occurred retrieving webhooks from the database", err)
	}
	if webhook != nil {
		return httperror.Conflict("A webhook for this resource already exists", errors.New("A webhook for this resource already exists"))
	}

	endpointID := payload.EndpointID

	endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}
	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	authorizations := []portainer.Authorization{portaineree.OperationPortainerWebhookCreate}

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

	token, err := uuid.NewV4()
	if err != nil {
		return httperror.InternalServerError("Error creating unique token", err)
	}

	webhook = &portainer.Webhook{
		Token:       token.String(),
		ResourceID:  payload.ResourceID,
		EndpointID:  endpointID,
		RegistryID:  payload.RegistryID,
		WebhookType: payload.WebhookType,
	}

	err = handler.DataStore.Webhook().Create(webhook)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the webhook inside the database", err)
	}

	return response.JSON(w, webhook)
}

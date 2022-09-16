package webhooks

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils/access"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/asaskevich/govalidator"
	"github.com/gofrs/uuid"
)

type webhookCreatePayload struct {
	ResourceID  string
	EndpointID  int
	RegistryID  portaineree.RegistryID
	WebhookType int
}

func (payload *webhookCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.ResourceID) {
		return errors.New("Invalid ResourceID")
	}
	if payload.EndpointID == 0 {
		return errors.New("Invalid EndpointID")
	}
	if payload.WebhookType != 1 && payload.WebhookType != 2 {
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
// @success 200 {object} portaineree.Webhook
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

	webhook, err := handler.dataStore.Webhook().WebhookByResourceID(payload.ResourceID)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return httperror.InternalServerError("An error occurred retrieving webhooks from the database", err)
	}
	if webhook != nil {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: "A webhook for this resource already exists", Err: errors.New("A webhook for this resource already exists")}
	}

	endpointID := portaineree.EndpointID(payload.EndpointID)

	endpoint, err := handler.dataStore.Endpoint().Endpoint(endpointID)
	if err == bolterrors.ErrObjectNotFound {
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

		_, err = access.GetAccessibleRegistry(handler.dataStore, tokenData.ID, endpointID, payload.RegistryID)
		if err != nil {
			return httperror.Forbidden("Permission deny to access registry", err)
		}
	}

	token, err := uuid.NewV4()
	if err != nil {
		return httperror.InternalServerError("Error creating unique token", err)
	}

	webhook = &portaineree.Webhook{
		Token:       token.String(),
		ResourceID:  payload.ResourceID,
		EndpointID:  endpointID,
		RegistryID:  payload.RegistryID,
		WebhookType: portaineree.WebhookType(payload.WebhookType),
	}

	err = handler.dataStore.Webhook().Create(webhook)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the webhook inside the database", err)
	}

	return response.JSON(w, webhook)
}

package webhooks

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gofrs/uuid"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	bolterrors "github.com/portainer/portainer-ee/api/bolt/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/registryutils/access"
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
	if payload.WebhookType != 1 {
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
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	webhook, err := handler.dataStore.Webhook().WebhookByResourceID(payload.ResourceID)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusInternalServerError, "An error occurred retrieving webhooks from the database", err}
	}
	if webhook != nil {
		return &httperror.HandlerError{http.StatusConflict, "A webhook for this resource already exists", errors.New("A webhook for this resource already exists")}
	}

	endpointID := portaineree.EndpointID(payload.EndpointID)

	endpoint, err := handler.dataStore.Endpoint().Endpoint(endpointID)
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}
	// endpoint will be used in the user activity logging middleware
	middlewares.SetEndpoint(endpoint, r)

	if payload.RegistryID != 0 {
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user authentication token", err}
		}

		_, err = access.GetAccessibleRegistry(handler.dataStore, tokenData.ID, endpointID, payload.RegistryID)
		if err != nil {
			return &httperror.HandlerError{http.StatusForbidden, "Permission deny to access registry", err}
		}
	}

	token, err := uuid.NewV4()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error creating unique token", err}
	}

	webhook = &portaineree.Webhook{
		Token:       token.String(),
		ResourceID:  payload.ResourceID,
		EndpointID:  endpointID,
		RegistryID:  payload.RegistryID,
		WebhookType: portaineree.WebhookType(payload.WebhookType),
	}

	err = handler.dataStore.Webhook().CreateWebhook(webhook)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the webhook inside the database", err}
	}

	return response.JSON(w, webhook)
}

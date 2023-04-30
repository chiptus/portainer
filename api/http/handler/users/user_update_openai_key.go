package users

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
)

type userUpdateOpenAIConfigPayload struct {
	// ApiKey is the OpenAI API key that will be used to interact with the OpenAI API.
	ApiKey string `validate:"required" example:"sk-1234567890"`
}

func (payload *userUpdateOpenAIConfigPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.ApiKey) {
		return errors.New("missing mandatory payload parameter: apiKey")
	}
	return nil
}

// @id UserUpdateOpenAIConfig
// @summary Update the OpenAI API configuration associated to a user.
// @description Update the OpenAI API key and OpenAI model associated to a user. Requires the OpenAI experimental feature setting to be enabled.
// @description This configuration will be used when interacting with the OpenAI chat.
// @description Only an administrator user or the user itself can update the OpenAI API key.
// @description **Access policy**: restricted
// @tags users
// @security jwt
// @accept json
// @produce json
// @param id path int true "User identifier"
// @param body body userUpdateOpenAIConfigPayload true "payload"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/openai [put]
func (handler *Handler) userUpdateOpenAIConfig(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload userUpdateOpenAIConfigPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	if !settings.ExperimentalFeatures.OpenAIIntegration {
		return httperror.Forbidden("OpenAI integration is not enabled", httperrors.ErrNotAvailable)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.Role != portaineree.AdministratorRole && tokenData.ID != portaineree.UserID(userID) {
		return httperror.Forbidden("Permission denied to update OpenAI API token", httperrors.ErrUnauthorized)
	}

	user, err := handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		return httperror.NotFound("Unable to find a user", err)
	}

	user.OpenAIApiKey = payload.ApiKey

	err = handler.DataStore.User().UpdateUser(user.ID, user)
	if err != nil {
		return httperror.InternalServerError("Unable to update user", err)
	}

	return response.Empty(w)
}

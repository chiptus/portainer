package users

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	bolterrors "github.com/portainer/portainer-ee/api/bolt/errors"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
)

// @id UserGetAPIKeys
// @summary Get all API keys for a user
// @description Gets all API keys for a user.
// @description Only the calling user or admin can retrieve api-keys.
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 200 {array} portaineree.APIKey "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/tokens [get]
func (handler *Handler) userGetAccessTokens(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid user identifier route variable", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user authentication token", err}
	}

	if tokenData.Role != portaineree.AdministratorRole && tokenData.ID != portaineree.UserID(userID) {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to get user access tokens", httperrors.ErrUnauthorized}
	}

	_, err = handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		if err == bolterrors.ErrObjectNotFound {
			return &httperror.HandlerError{http.StatusNotFound, "Unable to find a user with the specified identifier inside the database", err}
		}
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a user with the specified identifier inside the database", err}
	}

	apiKeys, err := handler.apiKeyService.GetAPIKeys(portaineree.UserID(userID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Internal Server Error", err}
	}

	for idx := range apiKeys {
		hideAPIKeyFields(&apiKeys[idx])
	}

	return response.JSON(w, apiKeys)
}

// hideAPIKeyFields remove the digest from the API key (it is not needed in the response)
func hideAPIKeyFields(apiKey *portaineree.APIKey) {
	apiKey.Digest = nil
}

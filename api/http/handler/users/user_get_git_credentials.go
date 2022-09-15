package users

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id UserGetGitCredentials
// @summary Get all saved git credentials for a user
// @description Gets all saved git credentials for a user.
// @description Only the calling user can retrieve git credentials
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 200 {array} portaineree.GitCredential "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/gitcredentials [get]
func (handler *Handler) userGetGitCredentials(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.ID != portaineree.UserID(userID) {
		return httperror.Forbidden("Couldn't retrieve git credential of another user", httperrors.ErrUnauthorized)
	}

	_, err = handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		if err == bolterrors.ErrObjectNotFound {
			return httperror.NotFound("Unable to find a user with the specified identifier inside the database", err)
		}
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	credentials, err := handler.DataStore.GitCredential().GetGitCredentialsByUserID(portaineree.UserID(userID))
	if err != nil {
		return httperror.InternalServerError("Couldn't retrieve git credential", err)
	}

	for idx := range credentials {
		hidePasswordFields(&credentials[idx])
	}

	return response.JSON(w, credentials)
}

// @id UserGetGitCredential
// @summary Get the specific saved git credential for a user
// @description Gets the specific saved git credential for a user.
// @description Only the calling user can retrieve git credential
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @param credentialID path int true "Git Credential identifier"
// @success 200 {object} portaineree.GitCredential "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/gitcredentials/{credentialID} [get]
func (handler *Handler) userGetGitCredential(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	credID, err := request.RetrieveNumericRouteVariableValue(r, "credentialID")
	if err != nil {
		return httperror.BadRequest("Invalid api-key identifier route variable", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.ID != portaineree.UserID(userID) {
		return httperror.Forbidden("Couldn't retrieve git credential of another user", httperrors.ErrUnauthorized)
	}

	_, err = handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		if err == bolterrors.ErrObjectNotFound {
			return httperror.Forbidden("Unable to find a user with the specified identifier inside the database", err)
		}
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	cred, err := handler.DataStore.GitCredential().GetGitCredential(portaineree.GitCredentialID(credID))
	if err != nil {
		return httperror.InternalServerError("Couldn't retrieve git credential", err)
	}

	if cred.UserID != portaineree.UserID(userID) {
		return httperror.Forbidden("Permission denied to get git-credential", httperrors.ErrUnauthorized)
	}

	hidePasswordFields(cred)

	return response.JSON(w, cred)
}

//  hidePasswordFields remove the password from the Git Credential (it is not needed in the response)
func hidePasswordFields(cred *portaineree.GitCredential) {
	cred.Password = ""
}

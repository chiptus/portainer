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

// @id UserRemoveGitCredential
// @summary Remove a git-credential associated to a user
// @description Remove a git-credential associated to a user..
// @description Only the calling user can remove git-credential
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @param id path int true "User identifier"
// @param credentialID path int true "Git Credential identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /users/{id}/gitcredentials/{credentialID} [delete]
func (handler *Handler) userRemoveGitCredential(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
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
		return httperror.Forbidden("Couldn't remove git credential for another user", httperrors.ErrUnauthorized)
	}

	_, err = handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		if err == bolterrors.ErrObjectNotFound {
			return httperror.Forbidden("Unable to find a user with the specified identifier inside the database", err)
		}
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	// check if the credential exists and the credential belongs to the user
	cred, err := handler.DataStore.GitCredential().GetGitCredential(portaineree.GitCredentialID(credID))
	if err != nil {
		return httperror.InternalServerError("Git credential not found", err)
	}
	if cred.UserID != portaineree.UserID(userID) {
		return httperror.Forbidden("Permission denied to remove git-credential", httperrors.ErrUnauthorized)
	}

	err = handler.DataStore.GitCredential().DeleteGitCredential(portaineree.GitCredentialID(credID))
	if err != nil {
		return httperror.InternalServerError("Unable to remove the git-credential from the user", err)
	}

	return response.Empty(w)
}

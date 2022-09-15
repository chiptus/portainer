package users

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
	dberrors "github.com/portainer/portainer/api/dataservices/errors"
)

type userGitCredentialUpdatePayload struct {
	Name     string `validate:"required" example:"my-git-credential" json:"name"`
	Username string `validate:"required" json:"username"`
	Password string `validate:"required" json:"password"`
}

func (payload *userGitCredentialUpdatePayload) Validate(r *http.Request) error {
	match, _ := regexp.MatchString("^[-_a-z0-9]+$", payload.Name)
	if !match {
		return errors.New("credential name must consist of lower case alphanumeric characters, '_' or '-'.")
	}
	if govalidator.HasWhitespaceOnly(payload.Username) {
		return errors.New("invalid username: cannot contain only whitespaces")
	}
	if govalidator.HasWhitespaceOnly(payload.Password) {
		return errors.New("invalid password: cannot contain only whitespaces")
	}

	return nil
}

// @id UserUpdateGitCredential
// @summary Update a git-credential associated to a user
// @description Update a git-credential associated to a user..
// @description Only the calling user can update git-credential
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
// @router /users/{id}/gitcredentials/{credentialID} [put]
func (handler *Handler) userUpdateGitCredential(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload userGitCredentialUpdatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

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
		return httperror.Forbidden("Couldn't update git credential for another user", httperrors.ErrUnauthorized)
	}

	_, err = handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		if err == bolterrors.ErrObjectNotFound {
			return httperror.NotFound("Unable to find a user with the specified identifier inside the database", err)
		}
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	// check if the credential exists and the credential belongs to the user
	cred, err := handler.DataStore.GitCredential().GetGitCredential(portaineree.GitCredentialID(credID))
	if err != nil {
		return httperror.InternalServerError("Git credential not found", err)
	}
	if cred.UserID != portaineree.UserID(userID) {
		return httperror.Forbidden("Couldn't update git credential for another user", httperrors.ErrUnauthorized)
	}

	// check if the credential name has been used
	credByName, err := handler.DataStore.GitCredential().GetGitCredentialByName(portaineree.UserID(userID), payload.Name)
	if err != nil && err != dberrors.ErrObjectNotFound {
		return httperror.InternalServerError("Unable to verify the git credential with name", err)
	}

	if credByName != nil && cred.ID != credByName.ID {
		return httperror.BadRequest("Git credential name already exists", err)
	}

	cred.Name = payload.Name
	cred.Username = payload.Username
	if payload.Password != "" {
		cred.Password = payload.Password
	}

	err = handler.DataStore.GitCredential().UpdateGitCredential(portaineree.GitCredentialID(credID), cred)
	if err != nil {
		return httperror.InternalServerError("Unable to update the git-credential from the user", err)
	}

	cred.Password = ""

	return response.JSON(w, gitCredentialResponse{*cred})
}

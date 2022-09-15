package users

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	dberrors "github.com/portainer/portainer/api/dataservices/errors"
)

type userGitCredentialCreatePayload struct {
	Name     string `validate:"required" example:"my-credential" json:"name"`
	Username string `validate:"required" json:"username"`
	Password string `validate:"required" json:"password"`
}

func (payload *userGitCredentialCreatePayload) Validate(r *http.Request) error {
	match, _ := regexp.MatchString("^[-_a-z0-9]+$", payload.Name)
	if !match {
		return errors.New("credential name must consist of lower case alphanumeric characters, '_' or '-'.")
	}

	if govalidator.HasWhitespaceOnly(payload.Username) {
		return errors.New("invalid username: cannot contain only whitespaces")
	}
	if govalidator.IsNull(payload.Password) {
		return errors.New("invalid password: cannot be empty")
	}
	if govalidator.HasWhitespaceOnly(payload.Password) {
		return errors.New("invalid password: cannot contain only whitespaces")
	}

	return nil
}

type gitCredentialResponse struct {
	GitCredential portaineree.GitCredential `json:"gitCredential"`
}

// @id UserCreateGitCredential
// @summary Store a Git Credential for a user
// @description Store a Git Credential for a user.
// @description Only the calling user can store a git credential for themselves.
// @description **Access policy**: restricted
// @tags users
// @security jwt
// @accept json
// @produce json
// @param id path int true "User identifier"
// @param body body userGitCredentialCreatePayload true "details"
// @success 201 {object} gitCredentialResponse "Created"
// @failure 400 "Invalid request"
// @failure 401 "Unauthorized"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @router /users/{id}/gitcredentials [post]
func (handler *Handler) userCreateGitCredential(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload userGitCredentialCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.ID != portaineree.UserID(userID) {
		return httperror.Forbidden("Couldn't create git credential for another user", httperrors.ErrUnauthorized)
	}

	_, err = handler.DataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		return httperror.BadRequest("Unable to find a user", err)
	}

	cred, err := handler.DataStore.GitCredential().GetGitCredentialByName(portaineree.UserID(userID), payload.Name)
	if err != nil && err != dberrors.ErrObjectNotFound {
		return httperror.InternalServerError("Unable to verify the git credential with name", err)
	}

	if cred != nil {
		return httperror.BadRequest("Git credential name already exists", err)
	}

	newCred := &portaineree.GitCredential{
		UserID:       portaineree.UserID(userID),
		Name:         payload.Name,
		Username:     payload.Username,
		Password:     payload.Password,
		CreationDate: time.Now().Unix(),
	}

	err = handler.DataStore.GitCredential().Create(newCred)
	if err != nil {
		return httperror.InternalServerError("Couldn't create a git credential", err)
	}

	newCred.Password = ""

	w.WriteHeader(http.StatusCreated)
	return response.JSON(w, gitCredentialResponse{*newCred})
}

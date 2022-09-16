package users

import (
	"errors"
	"net/http"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/asaskevich/govalidator"
)

type userUpdatePasswordPayload struct {
	// Current Password
	Password string `example:"passwd" validate:"required"`
	// New Password
	NewPassword string `example:"new_passwd" validate:"required"`
}

func (payload *userUpdatePasswordPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Password) {
		return errors.New("Invalid current password")
	}
	if govalidator.IsNull(payload.NewPassword) {
		return errors.New("Invalid new password")
	}
	return nil
}

// @id UserUpdatePassword
// @summary Update password for a user
// @description Update password for the specified user.
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "identifier"
// @param body body userUpdatePasswordPayload true "details"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/passwd [put]
func (handler *Handler) userUpdatePassword(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	if handler.demoService.IsDemoUser(portaineree.UserID(userID)) {
		return httperror.Forbidden(httperrors.ErrNotAvailableInDemo.Error(), httperrors.ErrNotAvailableInDemo)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.Role != portaineree.AdministratorRole && tokenData.ID != portaineree.UserID(userID) {
		return httperror.Forbidden("Permission denied to update user", httperrors.ErrUnauthorized)
	}

	var payload userUpdatePasswordPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	user, err := handler.DataStore.User().User(portaineree.UserID(userID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a user with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	err = handler.CryptoService.CompareHashAndData(user.Password, payload.Password)
	if err != nil {
		return httperror.Forbidden("Current password doesn't match", errors.New("Current password does not match the password provided. Please try again"))
	}

	if !handler.passwordStrengthChecker.Check(payload.NewPassword) {
		return httperror.BadRequest("Password does not meet the requirements", nil)
	}

	user.Password, err = handler.CryptoService.Hash(payload.NewPassword)
	if err != nil {
		return httperror.InternalServerError("Unable to hash user password", errCryptoHashFailure)
	}
	user.TokenIssueAt = time.Now().Unix()
	err = handler.DataStore.User().UpdateUser(user.ID, user)
	if err != nil {
		return httperror.InternalServerError("Unable to persist user changes inside the database", err)
	}

	return response.Empty(w)
}

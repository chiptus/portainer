package users

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
)

type userCreatePayload struct {
	Username string `validate:"required" example:"bob"`
	Password string `validate:"required" example:"cg9Wgky3"`
	// User role
	// 1 = administrator account
	// 2 = regular account
	Role portainer.UserRole `validate:"required" enums:"1,2" example:"1"`
}

func (payload *userCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Username) || govalidator.Contains(payload.Username, " ") {
		return errors.New("invalid username. Must not contain any whitespace")
	}

	if payload.Role != portaineree.AdministratorRole && payload.Role != portaineree.StandardUserRole {
		return errors.New("invalid role value. Value must be one of: 1 (administrator) or 2 (regular user)")
	}
	return nil
}

// @id UserCreate
// @summary Create a new user
// @description Create a new Portainer user.
// @description Only administrators can create users.
// @description **Access policy**: restricted
// @tags users
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body userCreatePayload true "User details"
// @success 200 {object} portaineree.User "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 409 "User already exists"
// @failure 500 "Server error"
// @router /users [post]
func (handler *Handler) userCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload userCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	user, err := handler.DataStore.User().UserByUsername(payload.Username)
	if err != nil && !handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.InternalServerError("Unable to retrieve users from the database", err)
	}
	if user != nil {
		return httperror.Conflict("Another user with the same username already exists", errUserAlreadyExists)
	}

	user = &portaineree.User{
		Username:                payload.Username,
		Role:                    payload.Role,
		PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
		UseCache:                true,
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	// when ldap/oauth is on, can only add users without password
	if (settings.AuthenticationMethod == portaineree.AuthenticationLDAP || settings.AuthenticationMethod == portaineree.AuthenticationOAuth) && payload.Password != "" {
		errMsg := "a user with password can not be created when authentication method is Oauth or LDAP"
		return httperror.BadRequest(errMsg, errors.New(errMsg))
	}

	if settings.AuthenticationMethod == portaineree.AuthenticationInternal {
		if !handler.passwordStrengthChecker.Check(payload.Password) {
			return httperror.BadRequest("Password does not meet the requirements", nil)
		}

		user.Password, err = handler.CryptoService.Hash(payload.Password)
		if err != nil {
			return httperror.InternalServerError("Unable to hash user password", errCryptoHashFailure)
		}
	}

	err = handler.DataStore.User().Create(user)
	if err != nil {
		return httperror.InternalServerError("Unable to persist user inside the database", err)
	}

	hideFields(user)
	return response.JSON(w, user)
}

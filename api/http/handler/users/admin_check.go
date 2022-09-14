package users

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/dataservices/errors"
)

// @id UserAdminCheck
// @summary Check administrator account existence
// @description Check if an administrator account exists in the database.
// @description **Access policy**: public
// @tags users
// @success 204 "Success"
// @failure 404 "User not found"
// @router /users/admin/check [get]
func (handler *Handler) adminCheck(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	users, err := handler.DataStore.User().UsersByRole(portaineree.AdministratorRole)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve users from the database", err)
	}

	if len(users) == 0 {
		return httperror.NotFound("No administrator account found inside the database", errors.ErrObjectNotFound)
	}

	return response.Empty(w)
}

package users

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id UserInspect
// @summary Inspect a user
// @description Retrieve details about a user.
// @description User passwords are filtered out, and should never be accessible.
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 200 {object} portaineree.User "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id} [get]
func (handler *Handler) userInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	// IF user is not admin
	// AND is not team leader
	// AND requesting user is not target user
	// EE-6176 TODO later: move this check to RBAC layer performed before the handler exec
	if !security.IsAdminOrEdgeAdminContext(securityContext) && !securityContext.IsTeamLeader && securityContext.UserID != portainer.UserID(userID) {
		return httperror.Forbidden("Permission denied inspect user", errors.ErrResourceAccessDenied)
	}

	user, err := handler.DataStore.User().Read(portainer.UserID(userID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a user with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	if securityContext.IsTeamLeader && security.IsAdminOrEdgeAdmin(user.Role) {
		return httperror.Forbidden("Permission denied inspect user", nil)
	}

	hideFields(user)
	return response.JSON(w, user)
}

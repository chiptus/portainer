package users

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
)

// @id UserMembershipsInspect
// @summary Inspect a user memberships
// @description Inspect a user memberships.
// @description **Access policy**: restricted
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 200 {object} portaineree.TeamMembership "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @router /users/{id}/memberships [get]
func (handler *Handler) userMemberships(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.Role != portaineree.AdministratorRole && tokenData.ID != portaineree.UserID(userID) {
		return httperror.Forbidden("Permission denied to update user memberships", errors.ErrUnauthorized)
	}

	memberships, err := handler.DataStore.TeamMembership().TeamMembershipsByUserID(portaineree.UserID(userID))
	if err != nil {
		return httperror.InternalServerError("Unable to persist membership changes inside the database", err)
	}

	return response.JSON(w, memberships)
}

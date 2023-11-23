package teammemberships

import (
	"errors"
	"net/http"

	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type teamMembershipUpdatePayload struct {
	// User identifier
	UserID int `validate:"required" example:"1"`
	// Team identifier
	TeamID int `validate:"required" example:"1"`
	// Role for the user inside the team (1 for leader and 2 for regular member)
	Role int `validate:"required" example:"1" enums:"1,2"`
}

func (payload *teamMembershipUpdatePayload) Validate(r *http.Request) error {
	if payload.UserID == 0 {
		return errors.New("Invalid UserID")
	}
	if payload.TeamID == 0 {
		return errors.New("Invalid TeamID")
	}
	if payload.Role != 1 && payload.Role != 2 {
		return errors.New("Invalid role value. Value must be one of: 1 (leader) or 2 (member)")
	}
	return nil
}

// @id TeamMembershipUpdate
// @summary Update a team membership
// @description Update a team membership. Access is only available to administrators or leaders of the associated team.
// @description **Access policy**: administrator or leaders of the associated team
// @tags team_memberships
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Team membership identifier"
// @param body body teamMembershipUpdatePayload true "Team membership details"
// @success 200 {object} portainer.TeamMembership "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "TeamMembership not found"
// @failure 500 "Server error"
// @router /team_memberships/{id} [put]
func (handler *Handler) teamMembershipUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	membershipID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid membership identifier route variable", err)
	}

	var payload teamMembershipUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	membership, err := handler.DataStore.TeamMembership().Read(portainer.TeamMembershipID(membershipID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a team membership with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a team membership with the specified identifier inside the database", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	isLeadingBothTeam := security.AuthorizedTeamManagement(portainer.TeamID(payload.TeamID), securityContext) &&
		security.AuthorizedTeamManagement(membership.TeamID, securityContext)
	// EE-6176 TODO later: move this check to RBAC layer performed before the handler exec
	if !(security.IsAdminContext(securityContext) || isLeadingBothTeam) {
		return httperror.Forbidden("Permission denied to update the membership", httperrors.ErrResourceAccessDenied)
	}

	previousUserID := int(membership.UserID)
	membership.UserID = portainer.UserID(payload.UserID)
	membership.TeamID = portainer.TeamID(payload.TeamID)
	membership.Role = portainer.MembershipRole(payload.Role)

	err = handler.DataStore.TeamMembership().Update(membership.ID, membership)
	if err != nil {
		return httperror.InternalServerError("Unable to persist membership changes inside the database", err)
	}

	handler.AuthorizationService.TriggerUserAuthUpdate(payload.UserID)
	handler.AuthorizationService.TriggerUserAuthUpdate(previousUserID)

	defer handler.updateUserServiceAccounts(membership)

	return response.JSON(w, membership)
}

package teams

import (
	"net/http"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

// @id TeamDelete
// @summary Remove a team
// @description Remove a team.
// @description **Access policy**: administrator
// @tags teams
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Team Id"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Team not found"
// @failure 500 "Server error"
// @router /teams/{id} [delete]
func (handler *Handler) teamDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	teamID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid team identifier route variable", err)
	}

	_, err = handler.DataStore.Team().Read(portaineree.TeamID(teamID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a team with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a team with the specified identifier inside the database", err)
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to get user environment access", err)
	}

	for _, endpoint := range endpoints {
		if endpoint.Type != portaineree.KubernetesLocalEnvironment &&
			endpoint.Type != portaineree.AgentOnKubernetesEnvironment &&
			endpoint.Type != portaineree.EdgeAgentOnKubernetesEnvironment {
			continue
		}

		kcl, err := handler.K8sClientFactory.GetKubeClient(&endpoint)
		if err != nil {
			return httperror.InternalServerError("Unable to get k8s environment access", err)
		}

		accessPolicies, err := kcl.GetNamespaceAccessPolicies()
		if err != nil {
			break
		}

		accessPolicies, hasChange, err := handler.AuthorizationService.RemoveTeamNamespaceAccessPolicies(
			handler.DataStore, teamID, int(endpoint.ID), accessPolicies,
		)
		if hasChange {
			err = kcl.UpdateNamespaceAccessPolicies(accessPolicies)
			if err != nil {
				break
			}
		}
	}

	err = handler.DataStore.Team().Delete(portaineree.TeamID(teamID))
	if err != nil {
		return httperror.InternalServerError("Unable to delete the team from the database", err)
	}

	err = handler.DataStore.TeamMembership().DeleteTeamMembershipByTeamID(portaineree.TeamID(teamID))
	if err != nil {
		return httperror.InternalServerError("Unable to delete associated team memberships from the database", err)
	}

	err = handler.AuthorizationService.RemoveTeamAccessPolicies(handler.DataStore, portaineree.TeamID(teamID))
	if err != nil {
		return httperror.InternalServerError("Unable to clean-up team access policies", err)
	}

	// update default team if deleted team was default
	err = handler.updateDefaultTeamIfDeleted(portaineree.TeamID(teamID))
	if err != nil {
		return httperror.InternalServerError("Unable to reset default team", err)
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	return response.Empty(w)
}

// updateDefaultTeamIfDeleted resets the default team to nil if default team was the deleted team
func (handler *Handler) updateDefaultTeamIfDeleted(teamID portaineree.TeamID) error {
	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return errors.Wrap(err, "failed to fetch settings")
	}

	if teamID != settings.OAuthSettings.DefaultTeamID {
		return nil
	}

	settings.OAuthSettings.DefaultTeamID = 0
	err = handler.DataStore.Settings().UpdateSettings(settings)
	return errors.Wrap(err, "failed to update settings")
}

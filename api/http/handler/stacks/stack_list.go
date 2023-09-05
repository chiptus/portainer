package stacks

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type stackListOperationFilters struct {
	SwarmID               string `json:"SwarmID"`
	EndpointID            int    `json:"EndpointID"`
	IncludeOrphanedStacks bool   `json:"IncludeOrphanedStacks"`
}

// @id StackList
// @summary List stacks
// @description List all stacks based on the current user authorizations.
// @description Will return all stacks if using an administrator account otherwise it
// @description will only return the list of stacks the user have access to.
// @description **Access policy**: authenticated
// @tags stacks
// @security ApiKeyAuth
// @security jwt
// @param filters query string false "Filters to process on the stack list. Encoded as JSON (a map[string]string). For example, {'SwarmID': 'jpofkc0i9uo9wtx1zesuk649w'} will only return stacks that are part of the specified Swarm cluster. Available filters: EndpointID, SwarmID."
// @success 200 {array} portaineree.Stack "Success"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /stacks [get]
func (handler *Handler) stackList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var filters stackListOperationFilters
	err := request.RetrieveJSONQueryParameter(r, "filters", &filters, true)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: filters", err)
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments from database", err)
	}

	stacks, err := handler.DataStore.Stack().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve stacks from the database", err)
	}
	stacks = filterStacks(stacks, &filters, endpoints)

	resourceControls, err := handler.DataStore.ResourceControl().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve resource controls from the database", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	stacks = authorization.DecorateStacks(stacks, resourceControls)

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user information from the database", err)
	}

	_, hasEndpointResourcesAccess := user.EndpointAuthorizations[portaineree.EndpointID(filters.EndpointID)][portaineree.EndpointResourcesAccess]

	if !securityContext.IsAdmin && !hasEndpointResourcesAccess {
		if filters.IncludeOrphanedStacks {
			return httperror.Forbidden("Permission denied to access orphaned stacks", httperrors.ErrUnauthorized)
		}

		userTeamIDs := make([]portaineree.TeamID, 0)
		for _, membership := range securityContext.UserMemberships {
			userTeamIDs = append(userTeamIDs, membership.TeamID)
		}

		stacks = authorization.FilterAuthorizedStacks(stacks, user, userTeamIDs)
	}

	for _, stack := range stacks {
		if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
			// sanitize password in the http response to minimise possible security leaks
			stack.GitConfig.Authentication.Password = ""
		}
	}

	return response.JSON(w, stacks)
}

func filterStacks(stacks []portaineree.Stack, filters *stackListOperationFilters, endpoints []portaineree.Endpoint) []portaineree.Stack {
	if filters.EndpointID == 0 && filters.SwarmID == "" {
		return stacks
	}

	filteredStacks := make([]portaineree.Stack, 0, len(stacks))
	for _, stack := range stacks {
		if filters.IncludeOrphanedStacks && isOrphanedStack(stack, endpoints) {
			if (stack.Type == portaineree.DockerComposeStack && filters.SwarmID == "") || (stack.Type == portaineree.DockerSwarmStack && filters.SwarmID != "") {
				filteredStacks = append(filteredStacks, stack)
			}
			continue
		}

		if stack.Type == portaineree.DockerComposeStack && stack.EndpointID == portaineree.EndpointID(filters.EndpointID) {
			filteredStacks = append(filteredStacks, stack)
		}
		if stack.Type == portaineree.DockerSwarmStack && stack.SwarmID == filters.SwarmID {
			filteredStacks = append(filteredStacks, stack)
		}
	}

	return filteredStacks
}

func isOrphanedStack(stack portaineree.Stack, endpoints []portaineree.Endpoint) bool {
	for _, endpoint := range endpoints {
		if stack.EndpointID == endpoint.ID {
			return false
		}
	}

	return true
}

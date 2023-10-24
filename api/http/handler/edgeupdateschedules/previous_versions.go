package edgeupdateschedules

import (
	"net/http"
	"slices"

	"github.com/portainer/portainer-ee/api/http/utils"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeUpdatePreviousVersions
// @summary Fetches the previous versions of updated agents
// @description
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param environmentIds query []int true "Environment IDs"
// @success 200 {object} map[portainer.EndpointID]string{}
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /edge_update_schedules/previous_versions [get]
func (handler *Handler) previousVersions(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	envIDs, err := utils.RetrieveNumberArrayQueryParameter[portainer.EndpointID](r, "environmentIds")
	if err != nil {
		return httperror.BadRequest(err.Error(), err)
	}

	if envIDs == nil {
		return httperror.BadRequest("Missing environmentIds query parameter", nil)
	}

	envs, err := handler.dataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments", err)
	}

	versionMap := map[portainer.EndpointID]string{}
	for _, env := range envs {
		if !slices.Contains(envIDs, env.ID) {
			continue
		}

		if endpointutils.IsEdgeEndpoint(&env) {
			versionMap[env.ID] = env.Agent.PreviousVersion
		}
	}

	return response.JSON(w, versionMap)
}

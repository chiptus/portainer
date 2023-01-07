package edgeupdateschedules

import (
	"net/http"
	"os"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

var Versions = []string{
	"2.17.0",
	"2.18.0",
}

func init() {
	env := os.Getenv("TEST_UPDATE_AGENT_VERSIONS")
	if env != "" {
		testVersions := strings.Split(env, ",")
		Versions = append(Versions, testVersions...)
	}
}

// @id AgentVersions
// @summary Fetches the supported versions of the agent to update/rollback
// @description
// @description **Access policy**: authenticated
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} string
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /edge_update_schedules/agent_versions [get]
func (h *Handler) agentVersions(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	return response.JSON(w, Versions)
}

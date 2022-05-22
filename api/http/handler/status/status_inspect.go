package status

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/demo"
)

type status struct {
	*portaineree.Status
	DemoEnvironment demo.EnvironmentDetails
}

// @id StatusInspect
// @summary Check Portainer status
// @description Retrieve Portainer status
// @description **Access policy**: public
// @tags status
// @produce json
// @success 200 {object} status "Success"
// @router /status [get]
func (handler *Handler) statusInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	return response.JSON(w, &status{
		Status:          handler.Status,
		DemoEnvironment: handler.demoService.Details(),
	})
}

package system

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/demo"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

type status struct {
	*portaineree.Status
	DemoEnvironment demo.EnvironmentDetails
}

// @id systemStatus
// @summary Check Portainer status
// @description Retrieve Portainer status
// @description **Access policy**: public
// @tags system
// @produce json
// @success 200 {object} status "Success"
// @router /system/status [get]
func (handler *Handler) systemStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	return response.JSON(w, &status{
		Status:          handler.status,
		DemoEnvironment: handler.demoService.Details(),
	})
}

// swagger docs for deprecated route:
// @id StatusInspect
// @summary Check Portainer status
// @deprecated
// @description Deprecated: use the `/system/status` endpoint instead.
// @description Retrieve Portainer status
// @description **Access policy**: public
// @tags status
// @produce json
// @success 200 {object} status "Success"
// @router /status [get]
func (handler *Handler) statusInspectDeprecated(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	log.Warn().Msg("The /status endpoint is deprecated and will be removed in a future version of Portainer. Please use the /system/status endpoint instead.")

	return handler.systemStatus(w, r)
}

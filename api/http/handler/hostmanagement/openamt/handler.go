package openamt

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle OpenAMT operations.
type Handler struct {
	*mux.Router
	OpenAMTService      portainer.OpenAMTService
	DataStore           dataservices.DataStore
	DockerClientFactory *client.ClientFactory
}

// NewHandler returns a new Handler
func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess)

	adminRouter.Handle("/open_amt/configure", httperror.LoggerHandler(h.openAMTConfigure)).Methods(http.MethodPost)
	adminRouter.Handle("/open_amt/{id}/info", httperror.LoggerHandler(h.openAMTHostInfo)).Methods(http.MethodGet)
	adminRouter.Handle("/open_amt/{id}/activate", httperror.LoggerHandler(h.openAMTActivate)).Methods(http.MethodPost)
	adminRouter.Handle("/open_amt/{id}/devices", httperror.LoggerHandler(h.openAMTDevices)).Methods(http.MethodGet)
	adminRouter.Handle("/open_amt/{id}/devices/{deviceId}/action", httperror.LoggerHandler(h.deviceAction)).Methods(http.MethodPost)
	adminRouter.Handle("/open_amt/{id}/devices/{deviceId}/features", httperror.LoggerHandler(h.deviceFeatures)).Methods(http.MethodPost)

	return h
}

package openamt

import (
	"net/http"

	"github.com/gorilla/mux"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
)

// Handler is the HTTP handler used to handle OpenAMT operations.
type Handler struct {
	*mux.Router
	OpenAMTService      portainer.OpenAMTService
	DataStore           dataservices.DataStore
	DockerClientFactory *docker.ClientFactory
}

// NewHandler returns a new Handler
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}

	h.Handle("/open_amt/configure", bouncer.AdminAccess(httperror.LoggerHandler(h.openAMTConfigure))).Methods(http.MethodPost)
	h.Handle("/open_amt/{id}/info", bouncer.AdminAccess(httperror.LoggerHandler(h.openAMTHostInfo))).Methods(http.MethodGet)
	h.Handle("/open_amt/{id}/activate", bouncer.AdminAccess(httperror.LoggerHandler(h.openAMTActivate))).Methods(http.MethodPost)
	h.Handle("/open_amt/{id}/devices", bouncer.AdminAccess(httperror.LoggerHandler(h.openAMTDevices))).Methods(http.MethodGet)
	h.Handle("/open_amt/{id}/devices/{deviceId}/action", bouncer.AdminAccess(httperror.LoggerHandler(h.deviceAction))).Methods(http.MethodPost)
	h.Handle("/open_amt/{id}/devices/{deviceId}/features", bouncer.AdminAccess(httperror.LoggerHandler(h.deviceFeatures))).Methods(http.MethodPost)

	return h
}

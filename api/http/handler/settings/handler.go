package settings

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
)

func hideFields(settings *portaineree.Settings) {
	settings.LDAPSettings.Password = ""
	settings.OAuthSettings.ClientSecret = ""
	settings.OAuthSettings.KubeSecretKey = nil

	cloudKeys := &settings.CloudApiKeys
	cloudKeys.CivoApiKey = strings.Repeat("*", len(cloudKeys.CivoApiKey))
	cloudKeys.LinodeToken = strings.Repeat("*", len(cloudKeys.LinodeToken))
	cloudKeys.DigitalOceanToken = strings.Repeat("*", len(cloudKeys.DigitalOceanToken))
}

// Handler is the HTTP handler used to handle settings operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            dataservices.DataStore
	FileService          portainer.FileService
	JWTService           portaineree.JWTService
	LDAPService          portaineree.LDAPService
	SnapshotService      portaineree.SnapshotService
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage settings operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)

	adminRouter.Handle("/settings", httperror.LoggerHandler(h.settingsInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/settings", httperror.LoggerHandler(h.settingsUpdate)).Methods(http.MethodPut)

	publicRouter.Handle("/settings/public", httperror.LoggerHandler(h.settingsPublic)).Methods(http.MethodGet)

	return h
}

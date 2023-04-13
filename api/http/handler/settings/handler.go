package settings

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/ssl"
)

func hideFields(settings *portaineree.Settings) {
	settings.LDAPSettings.Password = ""
	settings.OAuthSettings.ClientSecret = ""
	settings.OAuthSettings.KubeSecretKey = nil
	settings.Edge.MTLS.CaCertFile = ""
	settings.Edge.MTLS.CertFile = ""
	settings.Edge.MTLS.KeyFile = ""
}

// Handler is the HTTP handler used to handle settings operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            dataservices.DataStore
	FileService          portaineree.FileService
	JWTService           portaineree.JWTService
	LDAPService          portaineree.LDAPService
	SnapshotService      portaineree.SnapshotService
	SSLService           *ssl.Service
	userActivityService  portaineree.UserActivityService
	demoService          *demo.Service
}

// NewHandler creates a handler to manage settings operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService, demoService *demo.Service) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
		demoService:         demoService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)

	adminRouter.Handle("/settings", httperror.LoggerHandler(h.settingsInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/settings", httperror.LoggerHandler(h.settingsUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/settings/default_registry", httperror.LoggerHandler(h.defaultRegistryUpdate)).Methods(http.MethodPut)

	publicRouter.Handle("/settings/public", httperror.LoggerHandler(h.settingsPublic)).Methods(http.MethodGet)

	return h
}

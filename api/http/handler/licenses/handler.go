package licenses

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle Edge job operations.
type Handler struct {
	*mux.Router
	LicenseService      portaineree.LicenseService
	userActivityService portaineree.UserActivityService
	demoService         *demo.Service
}

// NewHandler creates a handler to manage Edge job operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService, demoService *demo.Service) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
		demoService:         demoService,
	}

	adminRouter := h.Router.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess, useractivity.LogUserActivity(h.userActivityService))
	adminRouter.Handle("/licenses", httperror.LoggerHandler(h.licensesList)).Methods(http.MethodGet)
	adminRouter.Handle("/licenses/add", httperror.LoggerHandler(h.licensesAttach)).Methods(http.MethodPost)
	adminRouter.Handle("/licenses/remove", httperror.LoggerHandler(h.licensesDelete)).Methods(http.MethodPost)

	publicRouter := h.Router.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)
	publicRouter.Handle("/licenses/info", httperror.LoggerHandler(h.licensesInfo)).Methods(http.MethodGet)

	return h
}

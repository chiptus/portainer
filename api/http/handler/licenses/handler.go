package licenses

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
)

// Handler is the HTTP handler used to handle Edge job operations.
type Handler struct {
	*mux.Router
	LicenseService      portaineree.LicenseService
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage Edge job operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	adminRouter := h.Router.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.Router.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)

	adminRouter.Handle("/licenses", httperror.LoggerHandler(h.licensesList)).Methods(http.MethodGet)
	adminRouter.Handle("/licenses", httperror.LoggerHandler(h.licensesAttach)).Methods(http.MethodPost)
	adminRouter.Handle("/licenses/remove", httperror.LoggerHandler(h.licensesDelete)).Methods(http.MethodPost)
	publicRouter.Handle("/licenses/info", httperror.LoggerHandler(h.licensesInfo)).Methods(http.MethodGet)
	return h
}

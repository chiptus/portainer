package tags

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	*mux.Router
	DataStore           portainer.DataStore
	userActivityService portainer.UserActivityService
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portainer.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)

	adminRouter.Handle("/tags", httperror.LoggerHandler(h.tagCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/tags/{id}", httperror.LoggerHandler(h.tagDelete)).Methods(http.MethodDelete)

	authenticatedRouter.Handle("/tags", httperror.LoggerHandler(h.tagList)).Methods(http.MethodGet)

	return h
}

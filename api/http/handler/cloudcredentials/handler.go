package cloudcredentials

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	*mux.Router
	DataStore           dataservices.DataStore
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/cloudcredentials", httperror.LoggerHandler(h.getAll)).Methods(http.MethodGet)
	adminRouter.Handle("/cloudcredentials", httperror.LoggerHandler(h.create)).Methods(http.MethodPost)
	adminRouter.Handle("/cloudcredentials/{id}", httperror.LoggerHandler(h.getByID)).Methods(http.MethodGet)
	adminRouter.Handle("/cloudcredentials/{id}", httperror.LoggerHandler(h.delete)).Methods(http.MethodDelete)
	adminRouter.Handle("/cloudcredentials/{id}", httperror.LoggerHandler(h.update)).Methods(http.MethodPut)

	return h
}

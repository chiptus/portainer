package cloudcredentials

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	*mux.Router
	DataStore           dataservices.DataStore
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/cloud/credentials", httperror.LoggerHandler(h.getAll)).Methods(http.MethodGet)
	adminRouter.Handle("/cloud/credentials", httperror.LoggerHandler(h.create)).Methods(http.MethodPost)
	adminRouter.Handle("/cloud/credentials/{id}", httperror.LoggerHandler(h.getByID)).Methods(http.MethodGet)
	adminRouter.Handle("/cloud/credentials/{id}", httperror.LoggerHandler(h.delete)).Methods(http.MethodDelete)
	adminRouter.Handle("/cloud/credentials/{id}", httperror.LoggerHandler(h.update)).Methods(http.MethodPut)

	return h
}

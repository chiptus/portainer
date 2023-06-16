package tags

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"

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

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)

	adminRouter.Handle("/tags", httperror.LoggerHandler(h.tagCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/tags/{id}", httperror.LoggerHandler(h.tagDelete)).Methods(http.MethodDelete)

	authenticatedRouter.Handle("/tags", httperror.LoggerHandler(h.tagList)).Methods(http.MethodGet)

	return h
}

func txResponse(w http.ResponseWriter, r any, err error) *httperror.HandlerError {
	if err != nil {
		var handlerError *httperror.HandlerError
		if errors.As(err, &handlerError) {
			return handlerError
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, r)
}

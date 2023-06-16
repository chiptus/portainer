package chat

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	*mux.Router
	DataStore       dataservices.DataStore
	SnapshotService portaineree.SnapshotService
	// userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		// userActivityService: userActivityService,
	}

	// adminRouter := h.NewRoute().Subrouter()
	// adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)

	authenticatedRouter.Handle("/chat", httperror.LoggerHandler(h.chatQuery)).Methods(http.MethodPost)

	return h
}

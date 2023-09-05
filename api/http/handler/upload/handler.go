package upload

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle upload operations.
type Handler struct {
	*mux.Router
	FileService         portainer.FileService
	userActivityService portaineree.UserActivityService
}

// NewHandler creates a handler to manage upload operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/upload/tls/{certificate:(?:ca|cert|key)}", httperror.LoggerHandler(h.uploadTLS)).Methods(http.MethodPost)
	return h
}

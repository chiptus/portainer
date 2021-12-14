package upload

import (
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"

	"net/http"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle upload operations.
type Handler struct {
	*mux.Router
	FileService         portainer.FileService
	userActivityService portainer.UserActivityService
}

// NewHandler creates a handler to manage upload operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portainer.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/upload/tls/{certificate:(?:ca|cert|key)}", httperror.LoggerHandler(h.uploadTLS)).Methods(http.MethodPost)
	return h
}

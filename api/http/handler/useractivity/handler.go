package useractivity

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle user activity operations
type Handler struct {
	*mux.Router
	UserActivityStore portaineree.UserActivityStore
}

// NewHandler creates a handler.
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}

	h.Use(bouncer.PureAdminAccess)
	h.Handle("/useractivity/authlogs", httperror.LoggerHandler(h.authLogsList)).Methods(http.MethodGet)
	h.Handle("/useractivity/authlogs.csv", httperror.LoggerHandler(h.authLogsCSV)).Methods(http.MethodGet)
	h.Handle("/useractivity/logs", httperror.LoggerHandler(h.logsList)).Methods(http.MethodGet)
	h.Handle("/useractivity/logs.csv", httperror.LoggerHandler(h.logsCSV)).Methods(http.MethodGet)

	return h
}

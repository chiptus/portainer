package useractivity

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
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

	h.Handle("/useractivity/authlogs",
		bouncer.AdminAccess(httperror.LoggerHandler(h.authLogsList))).Methods(http.MethodGet)
	h.Handle("/useractivity/authlogs.csv",
		bouncer.AdminAccess(httperror.LoggerHandler(h.authLogsCSV))).Methods(http.MethodGet)

	h.Handle("/useractivity/logs",
		bouncer.AdminAccess(httperror.LoggerHandler(h.logsList))).Methods(http.MethodGet)
	h.Handle("/useractivity/logs.csv",
		bouncer.AdminAccess(httperror.LoggerHandler(h.logsCSV))).Methods(http.MethodGet)

	return h
}

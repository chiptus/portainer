package teams

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Handler is the HTTP handler used to handle team operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            portainer.DataStore
	K8sClientFactory     *cli.ClientFactory
	userActivityService  portainer.UserActivityService
}

// NewHandler creates a handler to manage team operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portainer.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/teams", httperror.LoggerHandler(h.teamCreate)).Methods(http.MethodPost)
	h.Handle("/teams", httperror.LoggerHandler(h.teamList)).Methods(http.MethodGet)
	h.Handle("/teams/{id}", httperror.LoggerHandler(h.teamInspect)).Methods(http.MethodGet)
	h.Handle("/teams/{id}", httperror.LoggerHandler(h.teamUpdate)).Methods(http.MethodPut)
	h.Handle("/teams/{id}", httperror.LoggerHandler(h.teamDelete)).Methods(http.MethodDelete)
	h.Handle("/teams/{id}/memberships", httperror.LoggerHandler(h.teamMemberships)).Methods(http.MethodGet)

	return h
}

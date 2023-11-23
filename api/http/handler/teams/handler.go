package teams

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle team operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            dataservices.DataStore
	K8sClientFactory     *cli.ClientFactory
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage team operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess, useractivity.LogUserActivity(h.userActivityService))
	adminRouter.Handle("/teams", httperror.LoggerHandler(h.teamCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/teams/{id}", httperror.LoggerHandler(h.teamUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/teams/{id}", httperror.LoggerHandler(h.teamDelete)).Methods(http.MethodDelete)

	restrictedRouter := h.NewRoute().Subrouter()
	restrictedRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))
	restrictedRouter.Handle("/teams", httperror.LoggerHandler(h.teamList)).Methods(http.MethodGet)

	teamLeaderRouter := h.NewRoute().Subrouter()
	teamLeaderRouter.Use(bouncer.TeamLeaderAccess, useractivity.LogUserActivity(h.userActivityService))
	teamLeaderRouter.Handle("/teams/{id}", httperror.LoggerHandler(h.teamInspect)).Methods(http.MethodGet)
	teamLeaderRouter.Handle("/teams/{id}/memberships", httperror.LoggerHandler(h.teamMemberships)).Methods(http.MethodGet)

	return h
}

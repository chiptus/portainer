package teammemberships

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle team membership operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            dataservices.DataStore
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage team membership operations.
func NewHandler(bouncer security.BouncerService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.TeamLeaderAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/team_memberships", httperror.LoggerHandler(h.teamMembershipCreate)).Methods(http.MethodPost)
	h.Handle("/team_memberships", httperror.LoggerHandler(h.teamMembershipList)).Methods(http.MethodGet)
	h.Handle("/team_memberships/{id}", httperror.LoggerHandler(h.teamMembershipUpdate)).Methods(http.MethodPut)
	h.Handle("/team_memberships/{id}", httperror.LoggerHandler(h.teamMembershipDelete)).Methods(http.MethodDelete)

	return h
}

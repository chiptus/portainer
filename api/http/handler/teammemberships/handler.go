package teammemberships

import (
	"net/http"

	"github.com/rs/zerolog/log"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle team membership operations.
type Handler struct {
	*mux.Router
	AuthorizationService *authorization.Service
	DataStore            dataservices.DataStore
	userActivityService  portaineree.UserActivityService
	K8sClientFactory     *cli.ClientFactory
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

func (handler *Handler) updateUserServiceAccounts(membership *portainer.TeamMembership) {
	endpoints, err := handler.DataStore.Endpoint().EndpointsByTeamID(membership.TeamID)
	if err != nil {
		log.Error().Err(err).Msgf("failed fetching environments for team %d", membership.TeamID)
		return
	}
	for _, endpoint := range endpoints {
		// update kubernenets service accounts if the team is associated with a kubernetes environment
		if endpointutils.IsKubernetesEndpoint(&endpoint) {

			kubecli, err := handler.K8sClientFactory.GetKubeClient(&endpoint)
			if err != nil {
				log.Error().Err(err).Msgf("failed getting kube client for environment %d", endpoint.ID)
				continue
			}

			tokenManager, err := kubernetes.NewTokenManager(kubecli, handler.DataStore, nil, true, handler.AuthorizationService)
			if err != nil {
				log.Error().Err(err).Msgf("failed getting token manager client for environment %d", endpoint.ID)
				continue
			}

			err = tokenManager.SetupUserServiceAccounts(int(membership.UserID), &endpoint)
			if err != nil {
				log.Error().Err(err).Msgf("failed updating user service account for environment %d", endpoint.ID)
			}
		}
	}
}

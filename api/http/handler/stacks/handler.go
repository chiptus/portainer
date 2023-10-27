package stacks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/docker/consts"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks/deployments"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/docker/docker/api/types"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Handler is the HTTP handler used to handle stack operations.
type Handler struct {
	*mux.Router
	stackCreationMutex      *sync.Mutex
	stackDeletionMutex      *sync.Mutex
	requestBouncer          security.BouncerService
	userActivityService     portaineree.UserActivityService
	DataStore               dataservices.DataStore
	DockerClientFactory     *client.ClientFactory
	FileService             portainer.FileService
	GitService              portainer.GitService
	SwarmStackManager       portaineree.SwarmStackManager
	ComposeStackManager     portaineree.ComposeStackManager
	KubernetesDeployer      portaineree.KubernetesDeployer
	KubernetesClientFactory *cli.ClientFactory
	AuthorizationService    *authorization.Service
	Scheduler               *scheduler.Scheduler
	StackDeployer           deployments.StackDeployer
}

func stackExistsError(name string) *httperror.HandlerError {
	msg := fmt.Sprintf("A stack with the normalized name '%s' already exists", name)
	err := errors.New(msg)
	return httperror.Conflict(msg, err)
}

// NewHandler creates a handler to manage stack operations.
func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		DataStore:           dataStore,
		stackCreationMutex:  &sync.Mutex{},
		stackDeletionMutex:  &sync.Mutex{},
		requestBouncer:      bouncer,
		userActivityService: userActivityService,
	}

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)

	authenticatedRouter.Handle("/stacks/create/{type}/{method}", httperror.LoggerHandler(h.stackCreate)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/stacks", middlewares.Deprecated(authenticatedRouter, deprecatedStackCreateUrlParser)).Methods(http.MethodPost) // Deprecated
	authenticatedRouter.Handle("/stacks", httperror.LoggerHandler(h.stackList)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/stacks/{id}", httperror.LoggerHandler(h.stackInspect)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/stacks/{id}", httperror.LoggerHandler(h.stackDelete)).Methods(http.MethodDelete)
	authenticatedRouter.Handle("/stacks/name/{name}", httperror.LoggerHandler(h.stackDeleteKubernetesByName)).Methods(http.MethodDelete)
	authenticatedRouter.Handle("/stacks/{id}", httperror.LoggerHandler(h.stackUpdate)).Methods(http.MethodPut)
	authenticatedRouter.Handle("/stacks/{id}/git", httperror.LoggerHandler(h.stackUpdateGit)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/stacks/{id}/git/redeploy", httperror.LoggerHandler(h.stackGitRedeploy)).Methods(http.MethodPut)
	authenticatedRouter.Handle("/stacks/{id}/file", httperror.LoggerHandler(h.stackFile)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/stacks/{id}/migrate", httperror.LoggerHandler(h.stackMigrate)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/stacks/{id}/start", httperror.LoggerHandler(h.stackStart)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/stacks/{id}/stop", httperror.LoggerHandler(h.stackStop)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/stacks/{id}/images_status", httperror.LoggerHandler(h.stackImagesStatus)).Methods(http.MethodGet)

	adminRouter.Handle("/stacks/{id}/associate", httperror.LoggerHandler(h.stackAssociate)).Methods(http.MethodPut)

	publicRouter.Handle("/stacks/webhooks/{webhookID}", httperror.LoggerHandler(h.webhookInvoke)).Methods(http.MethodPost)

	return h
}

func (handler *Handler) userCanAccessStack(securityContext *security.RestrictedRequestContext, endpointID portainer.EndpointID, resourceControl *portainer.ResourceControl) (bool, error) {
	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return false, err
	}

	userTeamIDs := make([]portainer.TeamID, 0)
	for _, membership := range securityContext.UserMemberships {
		userTeamIDs = append(userTeamIDs, membership.TeamID)
	}

	if resourceControl != nil && authorization.UserCanAccessResource(securityContext.UserID, userTeamIDs, resourceControl) {
		return true, nil
	}

	return stackutils.UserIsAdminOrEndpointAdmin(user, endpointID)
}

func (handler *Handler) userIsAdmin(userID portainer.UserID) (bool, error) {
	user, err := handler.DataStore.User().Read(userID)
	if err != nil {
		return false, err
	}

	isAdmin := user.Role == portaineree.AdministratorRole

	return isAdmin, nil
}

func (handler *Handler) userCanCreateStack(securityContext *security.RestrictedRequestContext, endpointID portainer.EndpointID) (bool, error) {
	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return false, err
	}

	return stackutils.UserIsAdminOrEndpointAdmin(user, endpointID)
}

// if stack management is disabled for non admins and the user isn't an admin, then return false. Otherwise return true
func (handler *Handler) userCanManageStacks(securityContext *security.RestrictedRequestContext, endpoint *portaineree.Endpoint) (bool, error) {
	// When the endpoint is deleted, stacks that the deleted endpoint created will be tagged as an orphan stack
	// An orphan stack can be adopted by admins
	if endpoint == nil {
		return true, nil
	}

	if endpointutils.IsDockerEndpoint(endpoint) && !endpoint.SecuritySettings.AllowStackManagementForRegularUsers {
		canCreate, err := handler.userCanCreateStack(securityContext, portainer.EndpointID(endpoint.ID))

		if err != nil {
			return false, fmt.Errorf("failed to get user from the database: %w", err)
		}

		return canCreate, nil
	}
	return true, nil
}

func (handler *Handler) checkUniqueStackName(endpoint *portaineree.Endpoint, name string, stackID portainer.StackID) (bool, error) {
	stacks, err := handler.DataStore.Stack().ReadAll()
	if err != nil {
		return false, err
	}

	for _, stack := range stacks {
		if strings.EqualFold(stack.Name, name) && (stackID == 0 || stackID != stack.ID) && stack.EndpointID == endpoint.ID {
			return false, nil
		}
	}

	return true, nil
}

func (handler *Handler) checkUniqueStackNameInDocker(endpoint *portaineree.Endpoint, name string, stackID portainer.StackID, swarmMode bool) (bool, error) {
	isUniqueStackName, err := handler.checkUniqueStackName(endpoint, name, stackID)
	if err != nil {
		return false, err
	}

	dockerClient, err := handler.DockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return false, err
	}
	defer dockerClient.Close()
	if swarmMode {
		services, err := dockerClient.ServiceList(context.Background(), types.ServiceListOptions{})
		if err != nil {
			return false, err
		}

		for _, service := range services {
			serviceNS, ok := service.Spec.Labels["com.docker.stack.namespace"]
			if ok && serviceNS == name {
				return false, nil
			}
		}
	}

	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for _, container := range containers {
		containerNS, ok := container.Labels[consts.ComposeStackNameLabel]

		if ok && containerNS == name {
			return false, nil
		}
	}

	return isUniqueStackName, nil
}

func (handler *Handler) isUniqueWebhookID(webhookID string) (bool, error) {
	_, err := handler.DataStore.Stack().StackByWebhookID(webhookID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return true, nil
	}
	return false, err
}

func (handler *Handler) checkUniqueWebhookID(webhookID string) *httperror.HandlerError {
	if webhookID == "" {
		return nil
	}
	isUnique, err := handler.isUniqueWebhookID(webhookID)
	if err != nil {
		return httperror.InternalServerError("Unable to check for webhook ID collision", err)
	}
	if !isUnique {
		return httperror.Conflict(fmt.Sprintf("Webhook ID: %s already exists", webhookID), stackutils.ErrWebhookIDAlreadyExists)
	}
	return nil

}

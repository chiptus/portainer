package stacks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/portainer/portainer-ee/api/internal/endpointutils"

	"github.com/docker/docker/api/types"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks"
	portainer "github.com/portainer/portainer/api"
	dberrors "github.com/portainer/portainer/api/dataservices/errors"
)

var (
	errStackAlreadyExists     = errors.New("A stack already exists with this name")
	errWebhookIDAlreadyExists = errors.New("A webhook ID already exists")
	errInvalidGitCredential   = errors.New("Invalid git credential")
)

// Handler is the HTTP handler used to handle stack operations.
type Handler struct {
	*mux.Router
	stackCreationMutex      *sync.Mutex
	stackDeletionMutex      *sync.Mutex
	requestBouncer          *security.RequestBouncer
	userActivityService     portaineree.UserActivityService
	DataStore               dataservices.DataStore
	DockerClientFactory     *docker.ClientFactory
	FileService             portainer.FileService
	GitService              portaineree.GitService
	SwarmStackManager       portaineree.SwarmStackManager
	ComposeStackManager     portaineree.ComposeStackManager
	KubernetesDeployer      portaineree.KubernetesDeployer
	KubernetesClientFactory *cli.ClientFactory
	AuthorizationService    *authorization.Service
	Scheduler               *scheduler.Scheduler
	StackDeployer           stacks.StackDeployer
}

func stackExistsError(name string) *httperror.HandlerError {
	msg := fmt.Sprintf("A stack with the normalized name '%s' already exists", name)
	err := errors.New(msg)
	return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: msg, Err: err}
}

// NewHandler creates a handler to manage stack operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore, userActivityService portaineree.UserActivityService) *Handler {
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

	authenticatedRouter.Handle("/stacks", httperror.LoggerHandler(h.stackCreate)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/stacks", httperror.LoggerHandler(h.stackList)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/stacks/{id}", httperror.LoggerHandler(h.stackInspect)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/stacks/{id}", httperror.LoggerHandler(h.stackDelete)).Methods(http.MethodDelete)
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

func (handler *Handler) userCanAccessStack(securityContext *security.RestrictedRequestContext, endpointID portaineree.EndpointID, resourceControl *portaineree.ResourceControl) (bool, error) {
	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return false, err
	}

	userTeamIDs := make([]portaineree.TeamID, 0)
	for _, membership := range securityContext.UserMemberships {
		userTeamIDs = append(userTeamIDs, membership.TeamID)
	}

	if resourceControl != nil && authorization.UserCanAccessResource(securityContext.UserID, userTeamIDs, resourceControl) {
		return true, nil
	}

	return handler.userIsAdminOrEndpointAdmin(user, endpointID)
}

func (handler *Handler) userIsAdmin(userID portaineree.UserID) (bool, error) {
	user, err := handler.DataStore.User().User(userID)
	if err != nil {
		return false, err
	}

	isAdmin := user.Role == portaineree.AdministratorRole

	return isAdmin, nil
}

func (handler *Handler) userIsAdminOrEndpointAdmin(user *portaineree.User, endpointID portaineree.EndpointID) (bool, error) {
	isAdmin := user.Role == portaineree.AdministratorRole

	_, endpointResourceAccess := user.EndpointAuthorizations[portaineree.EndpointID(endpointID)][portaineree.EndpointResourcesAccess]

	return isAdmin || endpointResourceAccess, nil
}

func (handler *Handler) userCanCreateStack(securityContext *security.RestrictedRequestContext, endpointID portaineree.EndpointID) (bool, error) {
	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return false, err
	}

	return handler.userIsAdminOrEndpointAdmin(user, endpointID)
}

// if stack management is disabled for non admins and the user isn't an admin, then return false. Otherwise return true
func (handler *Handler) userCanManageStacks(securityContext *security.RestrictedRequestContext, endpoint *portaineree.Endpoint) (bool, error) {
	if endpointutils.IsDockerEndpoint(endpoint) && !endpoint.SecuritySettings.AllowStackManagementForRegularUsers {
		canCreate, err := handler.userCanCreateStack(securityContext, portaineree.EndpointID(endpoint.ID))

		if err != nil {
			return false, fmt.Errorf("Failed to get user from the database: %w", err)
		}

		return canCreate, nil
	}
	return true, nil
}

func (handler *Handler) checkUniqueStackName(endpoint *portaineree.Endpoint, name string, stackID portaineree.StackID) (bool, error) {
	stacks, err := handler.DataStore.Stack().Stacks()
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

func (handler *Handler) checkUniqueStackNameInKubernetes(endpoint *portaineree.Endpoint, name string, stackID portaineree.StackID, namespace string) (bool, error) {
	isUniqueStackName, err := handler.checkUniqueStackName(endpoint, name, stackID)
	if err != nil {
		return false, err
	}

	if !isUniqueStackName {
		// Check if this stack name is really used in the kubernetes.
		// Because the stack with this name could be removed via kubectl cli outside and the datastore does not be informed of this action.
		if namespace == "" {
			namespace = "default"
		}

		kubeCli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
		if err != nil {
			return false, err
		}
		isUniqueStackName, err = kubeCli.HasStackName(namespace, name)
		if err != nil {
			return false, err
		}
	}
	return isUniqueStackName, nil
}

func (handler *Handler) checkUniqueStackNameInDocker(endpoint *portaineree.Endpoint, name string, stackID portaineree.StackID, swarmMode bool) (bool, error) {
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
		containerNS, ok := container.Labels["com.docker.compose.project"]

		if ok && containerNS == name {
			return false, nil
		}
	}

	return isUniqueStackName, nil
}

func (handler *Handler) isUniqueWebhookID(webhookID string) (bool, error) {
	_, err := handler.DataStore.Stack().StackByWebhookID(webhookID)
	if err == dberrors.ErrObjectNotFound {
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
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: fmt.Sprintf("Webhook ID: %s already exists", webhookID), Err: errWebhookIDAlreadyExists}
	}
	return nil

}

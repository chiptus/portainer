package edgestacks

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	edgestackservice "github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	"github.com/portainer/portainer-ee/api/internal/edge/updateschedules"
	"github.com/portainer/portainer-ee/api/scheduler"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var errInvalidGitCredential = errors.New("Invalid git credential")

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	requestBouncer      *security.RequestBouncer
	DataStore           dataservices.DataStore
	FileService         portainer.FileService
	GitService          portainer.GitService
	userActivityService portaineree.UserActivityService
	edgeAsyncService    *edgeasync.Service
	edgeStacksService   *edgestackservice.Service
	edgeUpdateService   *updateschedules.Service
	KubernetesDeployer  portaineree.KubernetesDeployer
	scheduler           *scheduler.Scheduler
}

const contextKey = "edgeStack_item"

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(
	bouncer *security.RequestBouncer,
	userActivityService portaineree.UserActivityService,
	dataStore dataservices.DataStore,
	edgeAsyncService *edgeasync.Service,
	edgeStacksService *edgestackservice.Service,
	edgeUpdateService *updateschedules.Service,
	scheduler *scheduler.Scheduler,
) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		DataStore:           dataStore,
		requestBouncer:      bouncer,
		userActivityService: userActivityService,
		edgeAsyncService:    edgeAsyncService,
		edgeStacksService:   edgeStacksService,
		edgeUpdateService:   edgeUpdateService,
		scheduler:           scheduler,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/edge_stacks/create/{method}", httperror.LoggerHandler(h.edgeStackCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/edge_stacks", httperror.LoggerHandler(h.edgeStackList)).Methods(http.MethodGet)

	edgeStackAdminRouter := adminRouter.PathPrefix("/edge_stacks/{id}").Subrouter()
	edgeStackAdminRouter.Use(middlewares.WithItem(func(id portaineree.EdgeStackID) (*portaineree.EdgeStack, error) {
		return dataStore.EdgeStack().EdgeStack(id)
	}, "id", contextKey))

	edgeStackAdminRouter.Handle("/git", httperror.LoggerHandler(h.edgeStackUpdateFromGitHandler)).Methods(http.MethodPut)

	adminRouter.Handle("/edge_stacks/{id}", httperror.LoggerHandler(h.edgeStackInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/edge_stacks/{id}", httperror.LoggerHandler(h.edgeStackUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/edge_stacks/{id}", httperror.LoggerHandler(h.edgeStackDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/edge_stacks/{id}/file", httperror.LoggerHandler(h.edgeStackFile)).Methods(http.MethodGet)

	adminRouter.Handle("/edge_stacks/{id}/logs/{endpoint_id}", httperror.LoggerHandler(h.edgeStackLogsStatusGet)).Methods(http.MethodGet)
	adminRouter.Handle("/edge_stacks/{id}/logs/{endpoint_id}", httperror.LoggerHandler(h.edgeStackLogsCollect)).Methods(http.MethodPut)
	adminRouter.Handle("/edge_stacks/{id}/logs/{endpoint_id}", httperror.LoggerHandler(h.edgeStackLogsDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/edge_stacks/{id}/logs/{endpoint_id}/file", httperror.LoggerHandler(h.edgeStackLogsDownload)).Methods(http.MethodGet)

	publicRouter.Handle("/edge_stacks/{id}/status", httperror.LoggerHandler(h.edgeStackStatusUpdate)).Methods(http.MethodPut)
	publicRouter.Handle("/edge_stacks/webhooks/{webhookID}", httperror.LoggerHandler(h.webhookInvoke)).Methods(http.MethodPost)

	edgeStackStatusRouter := publicRouter.NewRoute().Subrouter()
	edgeStackStatusRouter.Use(middlewares.WithEndpoint(h.DataStore.Endpoint(), "endpoint_id"))
	edgeStackStatusRouter.PathPrefix("/edge_stacks/{id}/status/{endpoint_id}").Handler(httperror.LoggerHandler(h.edgeStackStatusDelete)).Methods(http.MethodDelete)

	return h
}

func (handler *Handler) convertAndStoreKubeManifestIfNeeded(stackFolder string, projectPath, composePath string, relatedEndpointIds []portaineree.EndpointID) (manifestPath string, err error) {
	hasKubeEndpoint, err := hasKubeEndpoint(handler.DataStore.Endpoint(), relatedEndpointIds)
	if err != nil {
		return "", fmt.Errorf("unable to check if edge stack has kube environments: %w", err)
	}

	if !hasKubeEndpoint {
		return "", nil
	}

	composeConfig, err := handler.FileService.GetFileContent(projectPath, composePath)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve Compose file from disk: %w", err)
	}

	kompose, err := handler.KubernetesDeployer.ConvertCompose(composeConfig)
	if err != nil {
		return "", fmt.Errorf("failed converting compose file to kubernetes manifest: %w", err)
	}

	komposeFileName := filesystem.ManifestFileDefaultName
	_, err = handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, komposeFileName, kompose)
	if err != nil {
		return "", fmt.Errorf("failed to store kube manifest file: %w", err)
	}

	return komposeFileName, nil
}

func (handler *Handler) handlerDBErr(err error, msg string) *httperror.HandlerError {
	httpErr := httperror.InternalServerError(msg, err)

	if handler.DataStore.IsErrObjectNotFound(err) {
		httpErr.StatusCode = http.StatusNotFound
	}

	return httpErr
}

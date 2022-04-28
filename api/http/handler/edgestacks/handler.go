package edgestacks

import (
	"fmt"
	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/edge"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"net/http"
	"strconv"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	requestBouncer      *security.RequestBouncer
	DataStore           dataservices.DataStore
	FileService         portainer.FileService
	GitService          portaineree.GitService
	userActivityService portaineree.UserActivityService
	edgeService         *edge.Service
	KubernetesDeployer  portaineree.KubernetesDeployer
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService, dataStore dataservices.DataStore, edgeService *edge.Service) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		DataStore:           dataStore,
		requestBouncer:      bouncer,
		userActivityService: userActivityService,
		edgeService:         edgeService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/edge_stacks", httperror.LoggerHandler(h.edgeStackCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/edge_stacks", httperror.LoggerHandler(h.edgeStackList)).Methods(http.MethodGet)
	adminRouter.Handle("/edge_stacks/{id}", httperror.LoggerHandler(h.edgeStackInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/edge_stacks/{id}", httperror.LoggerHandler(h.edgeStackUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/edge_stacks/{id}", httperror.LoggerHandler(h.edgeStackDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/edge_stacks/{id}/file", httperror.LoggerHandler(h.edgeStackFile)).Methods(http.MethodGet)

	publicRouter.Handle("/edge_stacks/{id}/status", httperror.LoggerHandler(h.edgeStackStatusUpdate)).Methods(http.MethodPut)

	edgeStackStatusRouter := publicRouter.NewRoute().Subrouter()
	edgeStackStatusRouter.Use(middlewares.WithEndpoint(h.DataStore.Endpoint(), "endpoint_id"))
	edgeStackStatusRouter.PathPrefix("/edge_stacks/{id}/status/{endpoint_id}").Handler(httperror.LoggerHandler(h.edgeStackStatusDelete)).Methods(http.MethodDelete)

	return h
}

func (handler *Handler) convertAndStoreKubeManifestIfNeeded(edgeStack *portaineree.EdgeStack, relatedEndpointIds []portaineree.EndpointID) error {
	hasKubeEndpoint, err := hasKubeEndpoint(handler.DataStore.Endpoint(), relatedEndpointIds)
	if err != nil {
		return fmt.Errorf("unable to check if edge stack has kube environments: %w", err)
	}

	if !hasKubeEndpoint {
		return nil
	}

	composeConfig, err := handler.FileService.GetFileContent(edgeStack.ProjectPath, edgeStack.EntryPoint)
	if err != nil {
		return fmt.Errorf("unable to retrieve Compose file from disk: %w", err)
	}

	kompose, err := handler.KubernetesDeployer.ConvertCompose(composeConfig)
	if err != nil {
		return fmt.Errorf("failed converting compose file to kubernetes manifest: %w", err)
	}

	KomposeFileName := filesystem.ManifestFileDefaultName
	_, err = handler.FileService.StoreEdgeStackFileFromBytes(strconv.Itoa(int(edgeStack.ID)), KomposeFileName, kompose)
	if err != nil {
		return fmt.Errorf("failed to store kube manifest file: %w", err)
	}

	edgeStack.ManifestPath = KomposeFileName

	return nil
}

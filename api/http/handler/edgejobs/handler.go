package edgejobs

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	portainer "github.com/portainer/portainer/api"
)

// Handler is the HTTP handler used to handle Edge job operations.
type Handler struct {
	*mux.Router
	DataStore            dataservices.DataStore
	FileService          portainer.FileService
	ReverseTunnelService portaineree.ReverseTunnelService
	userActivityService  portaineree.UserActivityService
	edgeService          *edgeasync.Service
}

// NewHandler creates a handler to manage Edge job operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portaineree.UserActivityService, edgeService *edgeasync.Service) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
		edgeService:         edgeService,
	}

	h.Use(bouncer.AdminAccess, bouncer.EdgeComputeOperation, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/edge_jobs", httperror.LoggerHandler(h.edgeJobList)).Methods(http.MethodGet)
	h.Handle("/edge_jobs", httperror.LoggerHandler(h.edgeJobCreate)).Methods(http.MethodPost)
	h.Handle("/edge_jobs/{id}", httperror.LoggerHandler(h.edgeJobInspect)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}", httperror.LoggerHandler(h.edgeJobUpdate)).Methods(http.MethodPut)
	h.Handle("/edge_jobs/{id}", httperror.LoggerHandler(h.edgeJobDelete)).Methods(http.MethodDelete)
	h.Handle("/edge_jobs/{id}/file", httperror.LoggerHandler(h.edgeJobFile)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}/tasks", httperror.LoggerHandler(h.edgeJobTasksList)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}/tasks/{taskID}/logs", httperror.LoggerHandler(h.edgeJobTaskLogsInspect)).Methods(http.MethodGet)
	h.Handle("/edge_jobs/{id}/tasks/{taskID}/logs", httperror.LoggerHandler(h.edgeJobTasksCollect)).Methods(http.MethodPost)
	h.Handle("/edge_jobs/{id}/tasks/{taskID}/logs", httperror.LoggerHandler(h.edgeJobTasksClear)).Methods(http.MethodDelete)

	return h
}

func convertEndpointsToMetaObject(endpoints []portaineree.EndpointID) map[portaineree.EndpointID]portaineree.EdgeJobEndpointMeta {
	endpointsMap := map[portaineree.EndpointID]portaineree.EdgeJobEndpointMeta{}

	for _, endpointID := range endpoints {
		endpointsMap[endpointID] = portaineree.EdgeJobEndpointMeta{}
	}

	return endpointsMap
}

func txResponse(w http.ResponseWriter, r any, err error) *httperror.HandlerError {
	if err != nil {
		if httpErr, ok := err.(*httperror.HandlerError); ok {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, r)
}

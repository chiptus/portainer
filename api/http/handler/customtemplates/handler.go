package customtemplates

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	portainer "github.com/portainer/portainer/api"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	DataStore           dataservices.DataStore
	FileService         portainer.FileService
	GitService          portainer.GitService
	userActivityService portaineree.UserActivityService
	gitFetchMutexs      map[portaineree.TemplateID]*sync.Mutex
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer *security.RequestBouncer, dataStore dataservices.DataStore, fileService portainer.FileService, gitService portainer.GitService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		DataStore:           dataStore,
		FileService:         fileService,
		GitService:          gitService,
		userActivityService: userActivityService,
		gitFetchMutexs:      make(map[portaineree.TemplateID]*sync.Mutex),
	}

	h.Use(bouncer.AuthenticatedAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/custom_templates/create/{method}", httperror.LoggerHandler(h.customTemplateCreate)).Methods(http.MethodPost)
	h.Handle("/custom_templates", httperror.LoggerHandler(h.customTemplateList)).Methods(http.MethodGet)
	h.Handle("/custom_templates/{id}", httperror.LoggerHandler(h.customTemplateInspect)).Methods(http.MethodGet)
	h.Handle("/custom_templates/{id}/file", httperror.LoggerHandler(h.customTemplateFile)).Methods(http.MethodGet)
	h.Handle("/custom_templates/{id}", httperror.LoggerHandler(h.customTemplateUpdate)).Methods(http.MethodPut)
	h.Handle("/custom_templates/{id}", httperror.LoggerHandler(h.customTemplateDelete)).Methods(http.MethodDelete)
	h.Handle("/custom_templates/{id}/git_fetch", httperror.LoggerHandler(h.customTemplateGitFetch)).Methods(http.MethodPut)
	return h
}

func userCanEditTemplate(customTemplate *portaineree.CustomTemplate, securityContext *security.RestrictedRequestContext) bool {
	return securityContext.IsAdmin || customTemplate.CreatedByUserID == securityContext.UserID
}

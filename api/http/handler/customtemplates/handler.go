package customtemplates

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
)

// Handler is the HTTP handler used to handle environment(endpoint) group operations.
type Handler struct {
	*mux.Router
	DataStore           portainer.DataStore
	FileService         portainer.FileService
	GitService          portainer.GitService
	userActivityService portainer.UserActivityService
}

// NewHandler creates a handler to manage environment(endpoint) group operations.
func NewHandler(bouncer *security.RequestBouncer, userActivityService portainer.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		userActivityService: userActivityService,
	}

	h.Use(bouncer.AuthenticatedAccess, useractivity.LogUserActivity(h.userActivityService))

	h.Handle("/custom_templates", httperror.LoggerHandler(h.customTemplateCreate)).Methods(http.MethodPost)
	h.Handle("/custom_templates", httperror.LoggerHandler(h.customTemplateList)).Methods(http.MethodGet)
	h.Handle("/custom_templates/{id}", httperror.LoggerHandler(h.customTemplateInspect)).Methods(http.MethodGet)
	h.Handle("/custom_templates/{id}/file", httperror.LoggerHandler(h.customTemplateFile)).Methods(http.MethodGet)
	h.Handle("/custom_templates/{id}", httperror.LoggerHandler(h.customTemplateUpdate)).Methods(http.MethodPut)
	h.Handle("/custom_templates/{id}", httperror.LoggerHandler(h.customTemplateDelete)).Methods(http.MethodDelete)
	return h
}

func userCanEditTemplate(customTemplate *portainer.CustomTemplate, securityContext *security.RestrictedRequestContext) bool {
	return securityContext.IsAdmin || customTemplate.CreatedByUserID == securityContext.UserID
}

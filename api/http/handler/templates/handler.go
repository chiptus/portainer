package templates

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler represents an HTTP API handler for managing templates.
type Handler struct {
	*mux.Router
	DataStore   dataservices.DataStore
	GitService  portainer.GitService
	FileService portainer.FileService
}

// NewHandler returns a new instance of Handler.
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}

	h.Handle("/templates",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.templateList))).Methods(http.MethodGet)
	h.Handle("/templates/file",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.templateFile))).Methods(http.MethodPost)
	return h
}

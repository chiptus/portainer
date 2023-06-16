package edgetemplates

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle edge environment(endpoint) operations.
type Handler struct {
	*mux.Router
	requestBouncer security.BouncerService
	DataStore      dataservices.DataStore
}

// NewHandler creates a handler to manage environment(endpoint) operations.
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}

	h.Handle("/edge_templates",
		bouncer.AdminAccess(httperror.LoggerHandler(h.edgeTemplateList))).Methods(http.MethodGet)

	return h
}

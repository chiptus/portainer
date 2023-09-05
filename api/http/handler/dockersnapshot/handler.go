package dockersnapshot

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
	dataStore dataservices.DataStore
}

// NewHandler creates a handler to process non-proxied requests to docker APIs directly.
func NewHandler(routePrefix string, bouncer security.BouncerService, dataStore dataservices.DataStore) *Handler {
	h := &Handler{
		Router:    mux.NewRouter(),
		dataStore: dataStore,
	}

	router := h.PathPrefix(routePrefix).Subrouter()
	router.Use(bouncer.AuthenticatedAccess)

	router.Handle("", httperror.LoggerHandler(h.snapshotInspect)).Methods(http.MethodGet)
	router.Handle("/containers", httperror.LoggerHandler(h.containersList)).Methods(http.MethodGet)
	router.Handle("/containers/{containerId}", httperror.LoggerHandler(h.containerInspect)).Methods(http.MethodGet)

	return h
}

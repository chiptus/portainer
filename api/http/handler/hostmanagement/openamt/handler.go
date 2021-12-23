package openamt

import (
	"net/http"

	"github.com/gorilla/mux"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
)

// Handler is the HTTP handler used to handle OpenAMT operations.
type Handler struct {
	*mux.Router
	OpenAMTService portaineree.OpenAMTService
	DataStore      portaineree.DataStore
}

// NewHandler returns a new Handler
func NewHandler(bouncer *security.RequestBouncer, dataStore portaineree.DataStore) (*Handler, error) {
	if !dataStore.Settings().IsFeatureFlagEnabled(portaineree.FeatOpenAMT) {
		return nil, nil
	}

	h := &Handler{
		Router: mux.NewRouter(),
	}

	h.Handle("/open_amt", bouncer.AdminAccess(httperror.LoggerHandler(h.openAMTConfigureDefault))).Methods(http.MethodPost)

	return h, nil
}

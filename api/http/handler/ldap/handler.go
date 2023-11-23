package ldap

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle LDAP search Operations
type Handler struct {
	*mux.Router
	DataStore   dataservices.DataStore
	FileService portainer.FileService
	LDAPService portaineree.LDAPService
}

// NewHandler returns a new Handler
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess)
	adminRouter.Handle("/ldap/check", httperror.LoggerHandler(h.ldapCheck)).Methods(http.MethodPost)
	adminRouter.Handle("/ldap/groups", httperror.LoggerHandler(h.ldapGroups)).Methods(http.MethodPost)
	adminRouter.Handle("/ldap/admin-groups", httperror.LoggerHandler(h.ldapAdminGroups)).Methods(http.MethodPost)
	adminRouter.Handle("/ldap/users", httperror.LoggerHandler(h.ldapUsers)).Methods(http.MethodPost)
	adminRouter.Handle("/ldap/test", httperror.LoggerHandler(h.ldapTestLogin)).Methods(http.MethodPost)

	return h
}

func (handler *Handler) prefillSettings(ldapSettings *portaineree.LDAPSettings) error {
	if !ldapSettings.AnonymousMode && ldapSettings.Password == "" {
		settings, err := handler.DataStore.Settings().Settings()
		if err != nil {
			return err
		}

		ldapSettings.Password = settings.LDAPSettings.Password
	}

	if (ldapSettings.TLSConfig.TLS || ldapSettings.StartTLS) && !ldapSettings.TLSConfig.TLSSkipVerify {
		caCertPath, err := handler.FileService.GetPathForTLSFile(filesystem.LDAPStorePath, portainer.TLSFileCA)
		if err != nil {
			return err
		}

		ldapSettings.TLSConfig.TLSCACertPath = caCertPath
	}

	return nil
}

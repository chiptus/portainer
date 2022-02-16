package users

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/apikey"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

var (
	errUserAlreadyExists          = errors.New("User already exists")
	errAdminAlreadyInitialized    = errors.New("An administrator user already exists")
	errAdminCannotRemoveSelf      = errors.New("Cannot remove your own user account. Contact another administrator")
	errCannotRemoveLastLocalAdmin = errors.New("Cannot remove the last local administrator account")
	errCryptoHashFailure          = errors.New("Unable to hash data")
)

func hideFields(user *portaineree.User) {
	user.Password = ""
}

// Handler is the HTTP handler used to handle user operations.
type Handler struct {
	*mux.Router
	bouncer              *security.RequestBouncer
	apiKeyService        apikey.APIKeyService
	AuthorizationService *authorization.Service
	CryptoService        portaineree.CryptoService
	DataStore            dataservices.DataStore
	K8sClientFactory     *cli.ClientFactory
	userActivityService  portaineree.UserActivityService
}

// NewHandler creates a handler to manage user operations.
func NewHandler(bouncer *security.RequestBouncer, rateLimiter *security.RateLimiter, apiKeyService apikey.APIKeyService, userActivityService portaineree.UserActivityService) *Handler {
	h := &Handler{
		Router:              mux.NewRouter(),
		bouncer:             bouncer,
		apiKeyService:       apiKeyService,
		userActivityService: userActivityService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.AdminAccess, useractivity.LogUserActivity(h.userActivityService))

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess, useractivity.LogUserActivity(h.userActivityService))

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess, useractivity.LogUserActivity(h.userActivityService))

	adminRouter.Handle("/users", httperror.LoggerHandler(h.userCreate)).Methods(http.MethodPost)
	adminRouter.Handle("/users", httperror.LoggerHandler(h.userList)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/users/{id}", httperror.LoggerHandler(h.userInspect)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/users/{id}", httperror.LoggerHandler(h.userUpdate)).Methods(http.MethodPut)
	adminRouter.Handle("/users/{id}", httperror.LoggerHandler(h.userDelete)).Methods(http.MethodDelete)
	adminRouter.Handle("/users/{id}/tokens", httperror.LoggerHandler(h.userGetAccessTokens)).Methods(http.MethodGet)
	adminRouter.Handle("/users/{id}/tokens", rateLimiter.LimitAccess(httperror.LoggerHandler(h.userCreateAccessToken))).Methods(http.MethodPost)
	adminRouter.Handle("/users/{id}/tokens/{keyID}", httperror.LoggerHandler(h.userRemoveAccessToken)).Methods(http.MethodDelete)
	adminRouter.Handle("/users/{id}/memberships", httperror.LoggerHandler(h.userMemberships)).Methods(http.MethodGet)
	adminRouter.Handle("/users/{id}/namespaces", httperror.LoggerHandler(h.userNamespaces)).Methods(http.MethodGet)
	authenticatedRouter.Handle("/users/{id}/passwd", rateLimiter.LimitAccess(httperror.LoggerHandler(h.userUpdatePassword))).Methods(http.MethodPut)
	publicRouter.Handle("/users/admin/check", httperror.LoggerHandler(h.adminCheck)).Methods(http.MethodGet)
	publicRouter.Handle("/users/admin/init", httperror.LoggerHandler(h.adminInit)).Methods(http.MethodPost)

	return h
}

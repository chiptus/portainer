package auth

import (
	"net/http"
	"regexp"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// Handler is the HTTP handler used to handle authentication operations.
type Handler struct {
	*mux.Router
	DataStore                   dataservices.DataStore
	CryptoService               portaineree.CryptoService
	JWTService                  portaineree.JWTService
	LDAPService                 portaineree.LDAPService
	LicenseService              portaineree.LicenseService
	OAuthService                portaineree.OAuthService
	ProxyManager                *proxy.Manager
	KubernetesTokenCacheManager *kubernetes.TokenCacheManager
	AuthorizationService        *authorization.Service
	UserActivityService         portaineree.UserActivityService
	passwordStrengthChecker     security.PasswordStrengthChecker
}

// NewHandler creates a handler to manage authentication operations.
func NewHandler(bouncer *security.RequestBouncer, rateLimiter *security.RateLimiter, passwordStrengthChecker security.PasswordStrengthChecker) *Handler {
	h := &Handler{
		Router:                  mux.NewRouter(),
		passwordStrengthChecker: passwordStrengthChecker,
	}

	h.Handle("/auth/oauth/validate",
		rateLimiter.LimitAccess(bouncer.PublicAccess(httperror.LoggerHandler(h.authActivityMiddleware(h.validateOAuth, portaineree.AuthenticationActivitySuccess))))).Methods(http.MethodPost)
	h.Handle("/auth",
		rateLimiter.LimitAccess(bouncer.PublicAccess(httperror.LoggerHandler(h.authActivityMiddleware(h.authenticate, portaineree.AuthenticationActivitySuccess))))).Methods(http.MethodPost)
	h.Handle("/auth/logout",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.authActivityMiddleware(h.logout, portaineree.AuthenticationActivityLogOut)))).Methods(http.MethodPost)

	return h
}

type authMiddlewareHandler func(http.ResponseWriter, *http.Request) (*authMiddlewareResponse, *httperror.HandlerError)

type authMiddlewareResponse struct {
	Username string
	Method   portaineree.AuthenticationMethod
}

func (handler *Handler) authActivityMiddleware(prev authMiddlewareHandler, defaultActivityType portaineree.AuthenticationActivityType) httperror.LoggerHandler {
	return func(rw http.ResponseWriter, r *http.Request) *httperror.HandlerError {
		resp, respErr := prev(rw, r)

		method := resp.Method
		if int(method) == 0 {
			method = portaineree.AuthenticationInternal
		}

		activityType := defaultActivityType
		if respErr != nil && activityType == portaineree.AuthenticationActivitySuccess {
			activityType = portaineree.AuthenticationActivityFailure
		}

		origin := getOrigin(r.RemoteAddr)

		err := handler.UserActivityService.LogAuthActivity(resp.Username, origin, method, activityType)
		if err != nil {
			log.Error().Err(err).Msg("failed logging auth activity")
		}

		return respErr
	}
}

func getOrigin(addr string) string {
	ipRegex := regexp.MustCompile(`:\d+$`)
	ipSplit := ipRegex.Split(addr, -1)
	if len(ipSplit) == 0 {
		return ""
	}

	return ipSplit[0]
}

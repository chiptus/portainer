package registries

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/portainer/portainer-ee/api/github/packages"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"

	"github.com/rs/zerolog/log"
)

// request on /api/registries/:id/v2
func (handler *Handler) proxyRequestsToRegistryAPI(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	hasAccess, _, err := handler.userHasRegistryAccess(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}
	if !hasAccess {
		return httperror.Forbidden("Access denied to resource", httperrors.ErrResourceAccessDenied)
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid registry identifier route variable", err)
	}

	registry, err := handler.DataStore.Registry().Read(portaineree.RegistryID(registryID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a registry with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a registry with the specified identifier inside the database", err)
	}

	managementConfiguration := syncConfig(registry)

	key := strconv.Itoa(int(registryID))

	forceCreate := false
	forceNew := r.Header.Get("X-RegistryManagement-ForceNew")
	if forceNew != "" {
		forceCreate = true
	}

	managementUrl := getRegistryManagementUrl(registry)
	proxy, err := handler.registryProxyService.GetProxy(key, managementUrl, managementConfiguration, forceCreate)
	if err != nil {
		return httperror.InternalServerError("Unable to create registry proxy", err)
	}

	if registry.Type == portaineree.ProGetRegistry {
		// replacePathRaw function does the following r.URL.RawPath = strings.Replace(r.URL.RawPath, "%2F", "/", -1)
		proxy = replacePathRaw("%2F", "/", proxy)
	}

	if registry.Type == portaineree.GithubRegistry {
		router := mux.NewRouter()
		gpPackages := packages.NewPackages(registry)
		router.Path("/v2/_catalog").Methods(http.MethodGet).Handler(proxyGithubRegistriesCatalog(gpPackages))
		router.Path("/v2/{userName}/{packageName}/manifests/{reference}").Methods(http.MethodDelete).Handler(proxyGithubRegistriesDeleteManifest(gpPackages))
		router.PathPrefix("/").Handler(proxy)
		http.StripPrefix("/registries/"+key, router).ServeHTTP(w, r)
		return nil
	}

	http.StripPrefix("/registries/"+key, proxy).ServeHTTP(w, r)

	return nil
}

func getRegistryManagementUrl(registry *portaineree.Registry) string {
	if registry.Type == portaineree.ProGetRegistry && registry.BaseURL != "" {
		log.Debug().Str("base_URL", registry.BaseURL).Int("registry_id", int(registry.ID)).Msg("")

		return registry.BaseURL
	}

	return registry.URL
}

func replacePathRaw(o, n string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawPath, o) {
			h.ServeHTTP(w, r)
		} else {
			r2 := *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.RawPath = strings.Replace(r.URL.RawPath, o, n, -1)
			h.ServeHTTP(w, &r2)
		}
	})
}

func createDefaultManagementConfiguration(registry *portaineree.Registry) *portaineree.RegistryManagementConfiguration {
	config := &portaineree.RegistryManagementConfiguration{
		Type: registry.Type,
		TLSConfig: portaineree.TLSConfiguration{
			TLS: false,
		},
	}

	if registry.Authentication {
		config.Username = registry.Username
		config.Password = registry.Password
		config.Ecr = registry.Ecr
		config.AccessToken = registry.AccessToken
		config.AccessTokenExpiry = registry.AccessTokenExpiry
	}

	return config
}

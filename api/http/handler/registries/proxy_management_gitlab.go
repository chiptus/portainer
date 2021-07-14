package registries

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
)

// request on /api/registries/{id}/proxies/gitlab
//
// Restricted to admins only
func (handler *Handler) proxyRequestsToGitlabAPIWithRegistry(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}
	if !securityContext.IsAdmin {
		return &httperror.HandlerError{http.StatusForbidden, "Access denied to resource", httperrors.ErrResourceAccessDenied}
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry identifier route variable", err}
	}

	registry, err := handler.DataStore.Registry().Registry(portainer.RegistryID(registryID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a registry with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a registry with the specified identifier inside the database", err}
	}

	config := &portainer.RegistryManagementConfiguration{
		Type:     portainer.GitlabRegistry,
		Password: registry.Password,
	}

	id := strconv.Itoa(int(registryID))

	proxy, err := handler.registryProxyService.GetProxy(id+"-gitlab", registry.Gitlab.InstanceURL, config, false)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create registry proxy", err}
	}

	http.StripPrefix("/registries/"+id+"/proxies/gitlab", proxy).ServeHTTP(w, r)
	return nil
}

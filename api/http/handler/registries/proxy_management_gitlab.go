package registries

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// request on /api/registries/{id}/proxies/gitlab
func (handler *Handler) proxyRequestsToGitlabAPIWithRegistry(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
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

	registry, err := handler.DataStore.Registry().Registry(portaineree.RegistryID(registryID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a registry with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a registry with the specified identifier inside the database", err)
	}

	config := &portaineree.RegistryManagementConfiguration{
		Type:     portaineree.GitlabRegistry,
		Password: registry.Password,
	}

	id := strconv.Itoa(int(registryID))

	proxy, err := handler.registryProxyService.GetProxy(id+"-gitlab", registry.Gitlab.InstanceURL, config, false)
	if err != nil {
		return httperror.InternalServerError("Unable to create registry proxy", err)
	}

	http.StripPrefix("/registries/"+id+"/proxies/gitlab", proxy).ServeHTTP(w, r)
	return nil
}

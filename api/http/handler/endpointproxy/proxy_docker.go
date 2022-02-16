package endpointproxy

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

func (handler *Handler) proxyRequestsToDockerAPI(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access environment", err}
	}

	if endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment {
		if endpoint.EdgeID == "" {
			return &httperror.HandlerError{http.StatusInternalServerError, "No Edge agent registered with the environment", errors.New("No agent available")}
		}

		_, err := handler.ReverseTunnelService.GetActiveTunnel(endpoint)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to get the active tunnel", err}
		}
	}

	var proxy http.Handler
	proxy = handler.ProxyManager.GetEndpointProxy(endpoint)
	if proxy == nil {
		proxy, err = handler.ProxyManager.CreateAndRegisterEndpointProxy(endpoint)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create proxy", err}
		}
	}

	id := strconv.Itoa(endpointID)

	prefix := "/" + id + "/agent/docker"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		prefix = "/" + id + "/docker"
	}

	http.StripPrefix(prefix, proxy).ServeHTTP(w, r)
	return nil
}

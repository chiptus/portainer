package endpointproxy

import (
	"errors"
	"fmt"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	bolterrors "github.com/portainer/portainer-ee/api/bolt/errors"

	"net/http"
)

func (handler *Handler) proxyRequestsToKubernetesAPI(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, false)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access environment", err}
	}

	if endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
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

	//  For KubernetesLocalEnvironment
	requestPrefix := fmt.Sprintf("/%d/kubernetes", endpointID)

	if endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
		requestPrefix = fmt.Sprintf("/%d", endpointID)

		agentPrefix := fmt.Sprintf("/%d/agent/kubernetes", endpointID)
		if strings.HasPrefix(r.URL.Path, agentPrefix) {
			requestPrefix = agentPrefix
		}
	}

	http.StripPrefix(requestPrefix, proxy).ServeHTTP(w, r)
	return nil
}

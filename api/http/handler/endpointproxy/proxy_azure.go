package endpointproxy

import (
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"

	"net/http"
)

func (handler *Handler) proxyRequestsToAzureAPI(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
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

	var proxy http.Handler
	proxy = handler.ProxyManager.GetEndpointProxy(endpoint)
	if proxy == nil {
		proxy, err = handler.ProxyManager.CreateAndRegisterEndpointProxy(endpoint)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create proxy", err}
		}
	}

	id := strconv.Itoa(endpointID)
	http.StripPrefix("/"+id+"/azure", proxy).ServeHTTP(w, r)
	return nil
}

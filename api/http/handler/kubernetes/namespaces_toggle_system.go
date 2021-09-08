package kubernetes

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer/api/http/middlewares"
	"github.com/portainer/portainer/api/http/useractivity"
)

type namespacesToggleSystemPayload struct {
	// Toggle the system state of this namespace to true or false
	System bool `example:"true"`
}

func (payload *namespacesToggleSystemPayload) Validate(r *http.Request) error {
	return nil
}

// @id KubernetesNamespacesToggleSystem
// @summary Toggle the system state for a namespace
// @description  Toggle the system state for a namespace
// @description **Access policy**: administrator or endpoint admin
// @security jwt
// @tags kubernetes
// @accept json
// @param id path int true "Endpoint identifier"
// @param namespace path string true "Namespace name"
// @param body body namespacesToggleSystemPayload true "Update details"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /kubernetes/{id}/namespaces/{namespace}/system [put]
func (handler *Handler) namespacesToggleSystem(rw http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment on request context", err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to perform operation", err}
	}

	namespaceName, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid namespace identifier route variable", err}
	}

	var payload namespacesToggleSystemPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	kubeClient, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create kubernetes client", err}
	}

	err = kubeClient.ToggleSystemState(namespaceName, payload.System)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to toggle system status", err}
	}

	useractivity.LogHttpActivity(handler.UserActivityStore, endpoint.Name, r, payload)

	return response.Empty(rw)

}
